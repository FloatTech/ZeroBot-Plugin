package control

import (
	"errors"

	sql "github.com/FloatTech/sqlite"
)

// ResponseState describes the existing OR rule without exposing extension data.
type ResponseState struct {
	GlobalEnabled    bool   `json:"globalEnabled"`
	GroupEnabled     bool   `json:"groupEnabled"`
	EffectiveEnabled bool   `json:"effectiveEnabled"`
	Source           string `json:"source"`
}

func (manager *Manager[CTX]) responseRowActiveLocked(gid int64) (bool, error) {
	var row ResponseGroup
	err := manager.D.Find("__resp", &row, "WHERE gid = ?", gid)
	if errors.Is(err, sql.ErrNullResult) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return row.Extra != "-", nil
}

// ReadResponseState reads global and positive-group rules under one lock. It
// neither changes response caches nor writes rows or invokes backup hooks.
func (manager *Manager[CTX]) ReadResponseState(gid int64) (ResponseState, error) {
	if gid < 0 {
		return ResponseState{}, errors.New("response scope must be global or a positive group")
	}
	manager.Lock()
	defer manager.Unlock()
	global, err := manager.responseRowActiveLocked(0)
	if err != nil {
		return ResponseState{}, err
	}
	local := false
	if gid != 0 {
		local, err = manager.responseRowActiveLocked(gid)
		if err != nil {
			return ResponseState{}, err
		}
	}
	source := "none"
	if global {
		source = "global"
	} else if local {
		source = "group"
	}
	return ResponseState{GlobalEnabled: global, GroupEnabled: local, EffectiveEnabled: global || local, Source: source}, nil
}
