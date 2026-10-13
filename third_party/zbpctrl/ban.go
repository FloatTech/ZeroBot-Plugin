package control

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"

	sql "github.com/FloatTech/sqlite"
)

func (m *Control[CTX]) banID(uid, gid int64) uint64 {
	var text string
	if gid == 0 {
		text = fmt.Sprintf("[%s]%d_all", m.Service, uid)
	} else {
		text = fmt.Sprintf("[%s]%d_%d", m.Service, uid, gid)
	}
	digest := md5.Sum([]byte(text))
	return binary.LittleEndian.Uint64(digest[:8])
}

// SQLite stores signed IDs. Reads/deletes use the same representation as writes.
func (m *Control[CTX]) Ban(uid, gid int64) error {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	id := m.banID(uid, gid)
	var row BanStatus
	err := m.Manager.D.Find(m.Service+"ban", &row, "WHERE id = ?", int64(id))
	if err == nil && row.UserID == uid && row.GroupID == gid {
		return nil
	}
	if err != nil && !errors.Is(err, sql.ErrNullResult) {
		return err
	}
	if err := m.Manager.beforePolicyLocked(); err != nil {
		return err
	}
	if err := m.Manager.D.Insert(m.Service+"ban", &BanStatus{ID: int64(id), UserID: uid, GroupID: gid}); err != nil {
		return err
	}
	m.Manager.banCache[id] = true
	return nil
}
func (m *Control[CTX]) Permit(uid, gid int64) error {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	id := m.banID(uid, gid)
	var row BanStatus
	err := m.Manager.D.Find(m.Service+"ban", &row, "WHERE id = ?", int64(id))
	if errors.Is(err, sql.ErrNullResult) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := m.Manager.beforePolicyLocked(); err != nil {
		return err
	}
	if err := m.Manager.D.Del(m.Service+"ban", "WHERE id = ?", int64(id)); err != nil {
		return err
	}
	m.Manager.banCache[id] = false
	return nil
}

// Lookup/cache insertion must not straddle restore. Group misses check global bans.
func (m *Control[CTX]) IsBannedIn(uid, gid int64) bool {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	groups := []int64{gid}
	if gid != 0 {
		groups = append(groups, 0)
	}
	for _, group := range groups {
		id := m.banID(uid, group)
		value, ok := m.Manager.banCache[id]
		if !ok {
			var row BanStatus
			err := m.Manager.D.Find(m.Service+"ban", &row, "WHERE id = ?", int64(id))
			value = err == nil && row.GroupID == group && row.UserID == uid
			m.Manager.banCache[id] = value
		}
		if value {
			return true
		}
	}
	return false
}
