package wordcount

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func TestWordResourcesFailuresDoNotPublishPartialState(t *testing.T) {
	for _, tc := range []struct {
		name     string
		data     []byte
		err      error
		stopErr  error
		wantStop bool
	}{
		{"download EOF", nil, errors.New("EOF"), nil, false},
		{"missing dictionary", nil, errors.New("missing"), nil, false},
		{"empty dictionary", []byte(" \n"), nil, nil, false},
		{"invalid frequency", []byte("测试 invalid n\n"), nil, nil, false},
		{"missing stopwords", []byte("测试 10 n\n"), nil, errors.New("missing stopwords"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			seg, words, err := loadWordResources(func() ([]byte, error) { return tc.data, tc.err }, func() ([]byte, error) { called = true; return nil, tc.stopErr })
			if err == nil || seg != nil || words != nil || called != tc.wantStop {
				t.Fatal("failure published resources or proceeded past invalid dictionary")
			}
		})
	}
}

func TestWordResourcesRetryAndSuccessfulSegmentation(t *testing.T) {
	attempt := 0
	read := func() ([]byte, error) {
		attempt++
		if attempt == 1 {
			return nil, errors.New("temporary EOF")
		}
		return []byte("测试 100 n\n词典 100 n\n"), nil
	}
	stop := func() ([]byte, error) { return []byte("词典\r\n测试"), nil }
	if _, _, err := loadWordResources(read, stop); err == nil {
		t.Fatal("first failure unexpectedly passed")
	}
	seg, words, err := loadWordResources(read, stop)
	if err != nil || attempt != 2 || !reflect.DeepEqual(words, []string{"测试", "词典"}) {
		t.Fatalf("retry or sorted stopwords failed: %v %v", words, err)
	}
	if frequency, exists := seg.Frequency("测试"); !exists || frequency <= 0 {
		t.Fatal("successful retry did not load dictionary")
	}
	if got := seg.Cut("测试词典", true); !reflect.DeepEqual(got, []string{"测试", "词典"}) {
		t.Fatalf("segmentation failed: %v", got)
	}
}

func TestWordResourceCacheConcurrentRetryAndReuse(t *testing.T) {
	var cache wordResourceCache
	var dictCalls, stopCalls atomic.Int32
	read := func() ([]byte, error) {
		if dictCalls.Add(1) == 1 {
			return nil, errors.New("temporary EOF")
		}
		return []byte("测试 100 n\n词典 100 n\n"), nil
	}
	stop := func() ([]byte, error) { stopCalls.Add(1); return []byte("词典\n测试"), nil }
	var wg sync.WaitGroup
	var failures, successes atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seg, words, err := cache.get(read, stop)
			if err != nil {
				failures.Add(1)
				if seg != nil || words != nil {
					t.Error("failed call received partial state")
				}
				return
			}
			if seg == nil || len(words) != 2 {
				t.Error("successful waiter received no resources")
			}
			successes.Add(1)
		}()
	}
	wg.Wait()
	if failures.Load() != 1 || successes.Load() != 19 || dictCalls.Load() != 2 || stopCalls.Load() != 1 {
		t.Fatal("concurrent failures/retries did not serialize or cached resources were reloaded")
	}
	if _, _, err := cache.get(func() ([]byte, error) { t.Error("cached dictionary read again"); return nil, errors.New("offline") }, func() ([]byte, error) { t.Error("cached stopwords read again"); return nil, errors.New("offline") }); err != nil {
		t.Fatal("cached resources lost")
	}
}
