package chat

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	binutils "github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/RomiChan/syncx"
)

var mems = syncx.Map[string, *memstorage]{}

type memstorage struct {
	mu   sync.Mutex
	root string
}

func atomicgetmemstorage(service string) (*memstorage, error) {
	d := path.Join(file.BOTPATH, "data", service, "agent")
	err := os.MkdirAll(d, 0700)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(d)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("Agent 记忆路径无效")
	}
	if err = protectHistoryPath(d, true); err != nil {
		return nil, err
	}
	x, _ := mems.LoadOrStore(service, &memstorage{root: d})
	return x, nil
}

func (ms *memstorage) Save(grp int64, text string) error {
	if ms.root == "" {
		return errors.New("Agent 记忆存储不可用")
	}
	cfg, err := CurrentConfig()
	if err != nil {
		return errors.New("Agent 记忆脱敏配置不可用")
	}
	text = cfg.Redact(text)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	p := path.Join(ms.root, strconv.FormatInt(grp, 10)+".txt")
	if info, err := os.Lstat(p); err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return errors.New("Agent 记忆文件无效")
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = protectHistoryPath(p, false); err != nil {
		return err
	}
	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	_, err = f.WriteString("\n")
	return err
}

func (ms *memstorage) Load(grp int64) []string {
	if ms.root == "" {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	p := path.Join(ms.root, strconv.FormatInt(grp, 10)+".txt")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	text := strings.TrimSpace(binutils.BytesToString(data))
	if cfg, err := CurrentConfig(); err == nil {
		text = cfg.Redact(text)
	} else {
		return nil
	}
	return []string{text}
}
