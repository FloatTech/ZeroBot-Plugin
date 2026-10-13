package control

import (
	"errors"

	sql "github.com/FloatTech/sqlite"
)

func (manager *Manager[CTX]) initBlock() error {
	return manager.D.Create("__block", &BlockStatus{})
}

// DoBlock 封禁
func (manager *Manager[CTX]) DoBlock(uid int64) error {
	manager.Lock()
	defer manager.Unlock()
	var row BlockStatus
	err := manager.D.Find("__block", &row, "WHERE uid = ?", uid)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNullResult) {
		return err
	}
	if err := manager.beforePolicyLocked(); err != nil {
		return err
	}
	if err := manager.D.Insert("__block", &BlockStatus{UserID: uid}); err != nil {
		return err
	}
	manager.blockCache[uid] = true
	return nil
}

// DoUnblock 解封
func (manager *Manager[CTX]) DoUnblock(uid int64) error {
	manager.Lock()
	defer manager.Unlock()
	var row BlockStatus
	err := manager.D.Find("__block", &row, "WHERE uid = ?", uid)
	if errors.Is(err, sql.ErrNullResult) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := manager.beforePolicyLocked(); err != nil {
		return err
	}
	if err := manager.D.Del("__block", "where uid = ?", uid); err != nil {
		return err
	}
	manager.blockCache[uid] = false
	return nil
}

// IsBlocked 是否封禁
func (manager *Manager[CTX]) IsBlocked(uid int64) bool {
	manager.Lock()
	defer manager.Unlock()
	isbl, ok := manager.blockCache[uid]
	if ok {
		return isbl
	}
	isbl = manager.D.CanFind("__block", "where uid = ?", uid)
	manager.blockCache[uid] = isbl
	return isbl
}
