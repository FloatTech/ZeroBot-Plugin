package modelstats

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (fn roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

type failedBody struct{ readErr, closeErr error }

func (b failedBody) Read([]byte) (int, error) { return 0, b.readErr }
func (b failedBody) Close() error             { return b.closeErr }
func response(status int, b io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Body: b, ContentLength: -1, Header: make(http.Header)}
}
func exchange(c *Collector, kind Kind, status int, b io.ReadCloser) (*http.Response, error) {
	cl := c.Client(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { return response(status, b), nil })}, kind)
	return cl.Get("http://example.invalid/private-path?secret=never-retained")
}
func invariant(t *testing.T, s Snapshot) {
	t.Helper()
	if s.Started != s.Active+s.Completed+s.Failed+s.Incomplete || s.MaxMilliseconds < s.AverageMilliseconds {
		t.Fatalf("inconsistent counters: %+v", s)
	}
	var total Counts
	for _, row := range s.Categories {
		total.Started += row.Started
		total.Active += row.Active
		total.Completed += row.Completed
		total.Failed += row.Failed
		total.Incomplete += row.Incomplete
	}
	if total.Started != s.Started || total.Active != s.Active || total.Completed != s.Completed || total.Failed != s.Failed || total.Incomplete != s.Incomplete {
		t.Fatal("category totals differ")
	}
}
func TestCompletionIncludesBodyAndIsCountedOnce(t *testing.T) {
	c := New()
	r, err := exchange(c, Vision, 200, io.NopCloser(strings.NewReader("private-response")))
	if err != nil {
		t.Fatal(err)
	}
	if s := c.Read(); s.Started != 1 || s.Active != 1 || s.Completed != 0 {
		t.Fatal(s)
	}
	data, err := io.ReadAll(r.Body)
	if err != nil || string(data) != "private-response" {
		t.Fatal("payload altered")
	}
	for i := 0; i < 2; i++ {
		r.Body.Close()
		r.Body.Read(make([]byte, 1))
	}
	s := c.Read()
	invariant(t, s)
	if s.Active != 0 || s.Completed != 1 || s.Failed != 0 || s.Incomplete != 0 || s.Categories[Vision].Completed != 1 {
		t.Fatal(s)
	}
	b, _ := json.Marshal(s)
	for _, secret := range []string{"private-response", "private-path", "never-retained", "example.invalid"} {
		if strings.Contains(string(b), secret) {
			t.Fatal("payload retained")
		}
	}
	s.Categories[Vision].Started = 999
	if c.Read().Categories[Vision].Started != 1 {
		t.Fatal("snapshot aliases collector")
	}
}
func TestTerminalOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		status                        int
		b                             io.ReadCloser
		read                          bool
		completed, failed, incomplete uint64
	}{
		{"empty", 204, http.NoBody, true, 1, 0, 0},
		{"http failure", 503, io.NopCloser(strings.NewReader("private-error")), true, 0, 1, 0},
		{"early close", 200, io.NopCloser(strings.NewReader("unused")), false, 0, 0, 1},
		{"read failure", 200, failedBody{readErr: io.ErrUnexpectedEOF}, true, 0, 1, 0},
		{"close failure", 200, failedBody{readErr: nil, closeErr: errors.New("secret-close-error")}, false, 0, 1, 0},
		{"http failure early close", 401, io.NopCloser(strings.NewReader("unused")), false, 0, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New()
			r, e := exchange(c, Chat, tc.status, tc.b)
			if e != nil {
				t.Fatal(e)
			}
			if tc.read {
				io.ReadAll(r.Body)
			}
			r.Body.Close()
			s := c.Read()
			invariant(t, s)
			if s.Active != 0 || s.Completed != tc.completed || s.Failed != tc.failed || s.Incomplete != tc.incomplete {
				t.Fatal(s)
			}
		})
	}
}
func TestTransportErrorAndPanic(t *testing.T) {
	for _, panics := range []bool{false, true} {
		c := New()
		cl := c.Client(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			if panics {
				panic("private-panic")
			}
			return nil, errors.New("private-network-error")
		})}, Agent)
		func() {
			defer func() {
				if recovered := recover(); panics != (recovered != nil) {
					t.Fatal("panic changed")
				}
			}()
			cl.Get("http://example.invalid/")
		}()
		s := c.Read()
		if s.Active != 0 || s.Failed != 1 {
			t.Fatal(s)
		}
		b, _ := json.Marshal(s)
		if strings.Contains(string(b), "private-") {
			t.Fatal("error retained")
		}
	}
}
func TestRedirectPolicyAndClientPreserved(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/first" {
			http.Redirect(w, r, server.URL+"/last", 302)
			return
		}
		io.WriteString(w, "answer")
	}))
	defer server.Close()
	jar, _ := cookiejar.New(nil)
	base := &http.Client{Timeout: time.Second, Jar: jar}
	c := New()
	cl := c.Client(base, GPT)
	if base.Transport != nil || cl == base || cl.Jar != base.Jar || cl.Timeout != base.Timeout {
		t.Fatal("client mutated")
	}
	r, err := cl.Get(server.URL + "/first")
	if err != nil {
		t.Fatal(err)
	}
	io.ReadAll(r.Body)
	r.Body.Close()
	s := c.Read()
	invariant(t, s)
	if s.Started != 2 || s.Completed != 1 || s.Failed != 1 {
		t.Fatal(s)
	}
	base.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c = New()
	r, err = c.Client(base, GPT).Get(server.URL + "/first")
	if err != nil || r.StatusCode != 302 {
		t.Fatal("redirect policy altered")
	}
	io.ReadAll(r.Body)
	r.Body.Close()
	if s = c.Read(); s.Started != 1 || s.Failed != 1 {
		t.Fatal(s)
	}
}
func TestConcurrentRequestsBoundedAndNoHistory(t *testing.T) {
	c := New()
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := exchange(c, Kind(i%4), 200, io.NopCloser(strings.NewReader("answer")))
			if err != nil {
				t.Error(err)
				return
			}
			invariant(t, c.Read())
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
		}(i)
	}
	wg.Wait()
	s := c.Read()
	invariant(t, s)
	if s.Started != 200 || s.Completed != 200 || s.Active != 0 || len(s.Categories) != 4 {
		t.Fatal(s)
	}
}

type idleTransport struct{ closed bool }

func (*idleTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return response(204, http.NoBody), nil
}
func (t *idleTransport) CloseIdleConnections() { t.closed = true }
func TestDefaultsInvalidKindAndIdleConnections(t *testing.T) {
	c := New()
	if cl := c.Client(nil, Chat); cl.Timeout != http.DefaultClient.Timeout || cl.Transport == nil {
		t.Fatal("default client changed")
	}
	base := &http.Client{}
	if c.Client(base, kindCount) != base || c.Read().Started != 0 {
		t.Fatal("invalid category accepted")
	}
	idle := &idleTransport{}
	c.Client(&http.Client{Transport: idle}, Chat).CloseIdleConnections()
	if !idle.closed {
		t.Fatal("idle close not forwarded")
	}
}

func TestCancellationAfterHeadersEndsActiveRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	c := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	response, err := c.Client(&http.Client{Timeout: time.Second}, Agent).Do(r)
	if err != nil {
		t.Fatal(err)
	}
	if s := c.Read(); s.Active != 1 || s.Completed != 0 {
		t.Fatal("headers ended request")
	}
	cancel()
	_, err = io.ReadAll(response.Body)
	response.Body.Close()
	if err == nil {
		t.Fatal("cancelled read succeeded")
	}
	s := c.Read()
	invariant(t, s)
	if s.Active != 0 || s.Failed != 1 {
		t.Fatal(s)
	}
}

func TestDeclaredLengthCompletesWithoutDecoderReadingEOF(t *testing.T) {
	for _, size := range []int{512, 1024, 65536} {
		c := New()
		data := `{"text":"` + strings.Repeat("a", size-11) + `"}`
		cl := c.Client(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			r := response(200, io.NopCloser(strings.NewReader(data)))
			r.ContentLength = int64(len(data))
			return r, nil
		})}, GPT)
		r, err := cl.Get("http://example.invalid/")
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]string
		if err = json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if s := c.Read(); s.Completed != 1 || s.Incomplete != 0 || s.Active != 0 {
			t.Fatal("valid complete JSON mislabeled", size, s)
		}
	}
}
