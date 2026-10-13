package control

import (
	"errors"
	"fmt"

	sql "github.com/FloatTech/sqlite"
)

var (
	ErrAlreadyResponding = errors.New("already in response")
	ErrAlreadySilent     = errors.New("already in silence")
)

// Mutations read persisted state and distinguish missing rows from DB failures.
// Reads here do not populate caches, so a rejected write leaves them unchanged.
func (manager *Manager[CTX]) persistedResponseLocked(gid int64) (bool, error) {
	ids := []int64{0}
	if gid != 0 {
		ids = append(ids, gid)
	}
	for _, id := range ids {
		var row ResponseGroup
		err := manager.D.Find("__resp", &row, "WHERE gid = ?", id)
		if errors.Is(err, sql.ErrNullResult) {
			continue
		}
		if err != nil {
			return false, err
		}
		if row.Extra != "-" {
			return true, nil
		}
	}
	return false, nil
}

func (manager *Manager[CTX]) initResponse() error {
	return manager.D.Create("__resp", &ResponseGroup{})
}

func (manager *Manager[CTX]) responseLocked(gid int64) bool {
	for _, id := range []int64{0, gid} {
		extra, ok := manager.respCache[id]
		if !ok {
			var row ResponseGroup
			if err := manager.D.Find("__resp", &row, "WHERE gid = ?", id); err == nil {
				extra = row.Extra
			} else {
				extra = "-"
			}
			manager.respCache[id] = extra
		}
		if extra != "-" {
			return true
		}
	}
	return false
}

func (manager *Manager[CTX]) Response(gid int64) error {
	manager.Lock()
	defer manager.Unlock()
	active, err := manager.persistedResponseLocked(gid)
	if err != nil {
		return err
	}
	if active {
		return fmt.Errorf("group %d: %w", gid, ErrAlreadyResponding)
	}
	if err := manager.beforePolicyLocked(); err != nil {
		return err
	}
	if err := manager.D.Insert("__resp", &ResponseGroup{GroupID: gid}); err != nil {
		return err
	}
	manager.respCache[gid] = ""
	return nil
}

// Silence retains the original row deletion semantics.
func (manager *Manager[CTX]) Silence(gid int64) error {
	manager.Lock()
	defer manager.Unlock()
	active, err := manager.persistedResponseLocked(gid)
	if err != nil {
		return err
	}
	if !active {
		return fmt.Errorf("group %d: %w", gid, ErrAlreadySilent)
	}
	var row ResponseGroup
	err = manager.D.Find("__resp", &row, "WHERE gid = ?", gid)
	if errors.Is(err, sql.ErrNullResult) {
		// Global response can be active without a local row to delete.
		return nil
	}
	if err != nil {
		return err
	}
	if err := manager.beforePolicyLocked(); err != nil {
		return err
	}
	if err := manager.D.Del("__resp", "WHERE gid = ?", gid); err != nil {
		return err
	}
	manager.respCache[gid] = "-"
	return nil
}

func (manager *Manager[CTX]) CanResponse(gid int64) bool {
	manager.Lock()
	defer manager.Unlock()
	return manager.responseLocked(gid)
}
