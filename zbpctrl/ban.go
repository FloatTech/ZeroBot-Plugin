package control

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var banCache = make(map[uint64]bool)

// Ban 禁止某人在某群使用本插件
func (m *Control[CTX]) Ban(uid, gid int64) {
	var err error
	var digest [16]byte
	if gid != 0 { // 特定群
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_%d", m.Service, uid, gid)))
		id := binary.LittleEndian.Uint64(digest[:8])
		m.Manager.Lock()
		err = m.Manager.D.Insert(m.Service+"ban", &BanStatus{ID: int64(id), UserID: uid, GroupID: gid})
		banCache[id] = true
		m.Manager.Unlock()
		if err == nil {
			log.Debugf("[control] plugin %s is banned in grp %d for usr %d.", m.Service, gid, uid)
			return
		}
	}
	// 所有群
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_all", m.Service, uid)))
	id := binary.LittleEndian.Uint64(digest[:8])
	m.Manager.Lock()
	err = m.Manager.D.Insert(m.Service+"ban", &BanStatus{ID: int64(id), UserID: uid, GroupID: 0})
	banCache[id] = true
	m.Manager.Unlock()
	if err == nil {
		log.Debugf("[control] plugin %s is banned in all grp for usr %d.", m.Service, uid)
	}
}

// Permit 允许某人在某群使用本插件
func (m *Control[CTX]) Permit(uid, gid int64) {
	var digest [16]byte
	if gid != 0 { // 特定群
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_%d", m.Service, uid, gid)))
		id := binary.LittleEndian.Uint64(digest[:8])
		m.Manager.Lock()
		_ = m.Manager.D.Del(m.Service+"ban", "WHERE id = "+strconv.FormatInt(int64(id), 10))
		banCache[id] = false
		m.Manager.Unlock()
		log.Debugf("[control] plugin %s is permitted in grp %d for usr %d.", m.Service, gid, uid)
		return
	}
	// 所有群
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_all", m.Service, uid)))
	id := binary.LittleEndian.Uint64(digest[:8])
	m.Manager.Lock()
	_ = m.Manager.D.Del(m.Service+"ban", "WHERE id = "+strconv.FormatInt(int64(id), 10))
	banCache[id] = false
	m.Manager.Unlock()
	log.Debugf("[control] plugin %s is permitted in all grp for usr %d.", m.Service, uid)
}

// IsBannedIn 某人是否在某群被 ban
func (m *Control[CTX]) IsBannedIn(uid, gid int64) bool {
	var b BanStatus
	var err error
	var digest [16]byte
	if gid != 0 {
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_%d", m.Service, uid, gid)))
		id := binary.LittleEndian.Uint64(digest[:8])
		m.Manager.RLock()
		if yes, ok := banCache[id]; ok {
			m.Manager.RUnlock()
			return yes
		}
		err = m.Manager.D.Find(m.Service+"ban", &b, "WHERE id = "+strconv.FormatInt(int64(id), 10))
		m.Manager.RUnlock()
		if err == nil && gid == b.GroupID && uid == b.UserID {
			log.Debugf("[control] plugin %s is banned in grp %d for usr %d.", m.Service, b.GroupID, b.UserID)
			m.Manager.Lock()
			banCache[id] = true
			m.Manager.Unlock()
			return true
		}
		m.Manager.Lock()
		banCache[id] = false
		m.Manager.Unlock()
	}
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_all", m.Service, uid)))
	id := binary.LittleEndian.Uint64(digest[:8])
	m.Manager.RLock()
	if yes, ok := banCache[id]; ok {
		m.Manager.RUnlock()
		return yes
	}
	err = m.Manager.D.Find(m.Service+"ban", &b, "WHERE id = "+strconv.FormatInt(int64(id), 10))
	m.Manager.RUnlock()
	if err == nil && b.GroupID == 0 && uid == b.UserID {
		log.Debugf("[control] plugin %s is banned in all grp for usr %d.", m.Service, b.UserID)
		m.Manager.Lock()
		banCache[id] = true
		m.Manager.Unlock()
		return true
	}
	m.Manager.Lock()
	banCache[id] = false
	m.Manager.Unlock()
	return false
}
