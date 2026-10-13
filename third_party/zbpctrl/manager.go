// Package control 控制插件的启用与优先级等
package control

import (
	"crypto/rand"
	"os"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// Manager 管理
type Manager[CTX any] struct {
	sync.RWMutex
	M           map[string]*Control[CTX]
	D           sql.Sqlite
	banCache    map[uint64]bool
	blockCache  map[int64]bool
	respCache   map[int64]string
	revisionKey [32]byte
	// BeforeSwitch is called under the Manager lock, before a switch/default
	// mutation. Admin installs a database snapshot hook shared by chat commands.
	BeforeSwitch func() error
	// BeforeExtra protects persistent extension writes with the same backup.
	BeforeExtra func() error
	// BeforePolicy protects response, user block and service ban mutations.
	// It runs under the Manager lock, after a persisted-state no-op check.
	BeforePolicy func() error
}

func (m *Manager[CTX]) beforePolicyLocked() error {
	if m.BeforePolicy != nil {
		return m.BeforePolicy()
	}
	return nil
}

// NewManager 打开管理数据库
func NewManager[CTX any](dbpath string) (m Manager[CTX]) {
	switch {
	case dbpath == "":
		dbpath = "ctrl.db"
	case strings.HasSuffix(dbpath, "/"):
		err := os.MkdirAll(dbpath, 0755)
		if err != nil {
			panic(err)
		}
		dbpath += "ctrl.db"
	default:
		i := strings.LastIndex(dbpath, "/")
		if i > 0 {
			err := os.MkdirAll(dbpath[:i], 0755)
			if err != nil {
				panic(err)
			}
		}
	}
	m = Manager[CTX]{
		M:          map[string]*Control[CTX]{},
		D:          sql.New(dbpath),
		banCache:   make(map[uint64]bool, 8),
		blockCache: make(map[int64]bool, 8),
		respCache:  make(map[int64]string, 8),
	}
	if _, err := rand.Read(m.revisionKey[:]); err != nil {
		panic(err)
	}
	err := m.D.Open(time.Hour)
	if err != nil {
		panic(err)
	}
	err = m.initBlock()
	if err != nil {
		panic(err)
	}
	err = m.initResponse()
	if err != nil {
		panic(err)
	}
	err = m.D.Create("__switch_inherit", &switchInheritance{})
	if err != nil {
		panic(err)
	}
	return
}

// Lookup returns a Manager by the service name, if
// not exist, it will return nil.
func (manager *Manager[CTX]) Lookup(service string) (*Control[CTX], bool) {
	manager.RLock()
	m, ok := manager.M[service]
	manager.RUnlock()
	return m, ok
}

// ForEach iterates through managers.
func (manager *Manager[CTX]) ForEach(iterator func(key string, manager *Control[CTX]) bool) {
	manager.RLock()
	m := cpmp(manager.M)
	manager.RUnlock()
	for k, v := range m {
		if !iterator(k, v) {
			return
		}
	}
}

func cpmp[CTX any](m map[string]*Control[CTX]) map[string]*Control[CTX] {
	ret := make(map[string]*Control[CTX], len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}
