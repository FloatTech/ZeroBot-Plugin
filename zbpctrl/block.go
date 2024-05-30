package control

import (
	"strconv"
)

func (manager *Manager[CTX]) initBlock() error {
	return manager.D.Create("__block", &BlockStatus{})
}

var blockCache = make(map[int64]bool)

// DoBlock 封禁
func (manager *Manager[CTX]) DoBlock(uid int64) error {
	manager.Lock()
	defer manager.Unlock()
	blockCache[uid] = true
	return manager.D.Insert("__block", &BlockStatus{UserID: uid})
}

// DoUnblock 解封
func (manager *Manager[CTX]) DoUnblock(uid int64) error {
	manager.Lock()
	defer manager.Unlock()
	blockCache[uid] = false
	return manager.D.Del("__block", "where uid = "+strconv.FormatInt(uid, 10))
}

// IsBlocked 是否封禁
func (manager *Manager[CTX]) IsBlocked(uid int64) bool {
	manager.RLock()
	isbl, ok := blockCache[uid]
	manager.RUnlock()
	if ok {
		return isbl
	}
	manager.Lock()
	defer manager.Unlock()
	isbl = manager.D.CanFind("__block", "where uid = "+strconv.FormatInt(uid, 10))
	blockCache[uid] = isbl
	return isbl
}
