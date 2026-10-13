package wordcount

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/fumiama/jieba"
)

type wordResourceCache struct {
	mu    sync.Mutex
	seg   *jieba.Segmenter
	words []string
}

func (c *wordResourceCache) get(readDictionary, readStopwords func() ([]byte, error)) (*jieba.Segmenter, []string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.seg != nil {
		return c.seg, c.words, nil
	}
	seg, words, err := loadWordResources(readDictionary, readStopwords)
	if err != nil {
		return nil, nil, err
	}
	c.seg, c.words = seg, words
	return seg, words, nil
}

// Publish resources only after both have loaded. Callers can retry failures.
// This runs on a matching command, never during package initialization.
func loadWordResources(readDictionary, readStopwords func() ([]byte, error)) (*jieba.Segmenter, []string, error) {
	dat, err := readDictionary()
	if err != nil {
		return nil, nil, fmt.Errorf("分词词典读取失败: %w", err)
	}
	if len(bytes.TrimSpace(dat)) == 0 {
		return nil, nil, fmt.Errorf("分词词典为空")
	}
	seg, err := jieba.LoadDictionary(bytes.NewReader(dat))
	if err != nil {
		return nil, nil, fmt.Errorf("分词词典解析失败: %w", err)
	}
	data, err := readStopwords()
	if err != nil {
		return nil, nil, fmt.Errorf("停用词读取失败: %w", err)
	}
	words := strings.Split(strings.ReplaceAll(string(data), "\r", ""), "\n")
	sort.Strings(words)
	return seg, words, nil
}
