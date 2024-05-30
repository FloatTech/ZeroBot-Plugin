package control

import (
	"errors"
	"strconv"
)

// InitResponse ...
func (manager *Manager[CTX]) initResponse() error {
	return manager.D.Create("__resp", &ResponseGroup{})
}

var respCache = make(map[int64]string)

// Response opens the resp of the gid
func (manager *Manager[CTX]) Response(gid int64) error {
	if manager.CanResponse(gid) {
		return errors.New("group " + strconv.FormatInt(gid, 10) + " already in response")
	}
	manager.Lock()
	respCache[gid] = ""
	err := manager.D.Insert("__resp", &ResponseGroup{GroupID: gid})
	manager.Unlock()
	return err
}

// Silence will drop its extra data
func (manager *Manager[CTX]) Silence(gid int64) error {
	if !manager.CanResponse(gid) {
		return errors.New("group " + strconv.FormatInt(gid, 10) + " already in silence")
	}
	manager.Lock()
	respCache[gid] = "-"
	err := manager.D.Del("__resp", "where gid = "+strconv.FormatInt(gid, 10))
	manager.Unlock()
	return err
}

// CanResponse ...
func (manager *Manager[CTX]) CanResponse(gid int64) bool {
	manager.RLock()
	ext, ok := respCache[0] // all status
	manager.RUnlock()
	if ok && ext != "-" {
		return true
	}
	manager.RLock()
	ext, ok = respCache[gid]
	manager.RUnlock()
	if ok {
		return ext != "-"
	}
	manager.RLock()
	var rsp ResponseGroup
	err := manager.D.Find("__resp", &rsp, "where gid = 0") // all status
	manager.RUnlock()
	if err == nil && rsp.Extra != "-" {
		manager.Lock()
		respCache[0] = rsp.Extra
		manager.Unlock()
		return true
	}
	manager.RLock()
	err = manager.D.Find("__resp", &rsp, "where gid = "+strconv.FormatInt(gid, 10))
	manager.RUnlock()
	if err != nil {
		manager.Lock()
		respCache[gid] = "-"
		manager.Unlock()
		return false
	}
	manager.Lock()
	respCache[gid] = rsp.Extra
	manager.Unlock()
	return rsp.Extra != "-"
}
