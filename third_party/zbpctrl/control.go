package control

import (
	"errors"
	"math/bits"

	sql "github.com/FloatTech/sqlite"
	log "github.com/sirupsen/logrus"
)

// Control is to control the plugins.
type Control[CTX any] struct {
	Service string
	Cache   map[int64]uint8 // resolved state per location, invalidated after commit
	Options Options[CTX]
	Manager *Manager[CTX]
	// BeforeData is installed only for configuration data, not business counters.
	BeforeData func() error
}

func (manager *Manager[CTX]) NewControl(service string, o *Options[CTX]) *Control[CTX] {
	m := &Control[CTX]{Service: service, Cache: make(map[int64]uint8, 16), Manager: manager}
	if o != nil {
		m.Options = *o
	}
	manager.Lock()
	defer manager.Unlock()
	var c GroupConfig
	if err := manager.D.Create(service, &c); err != nil {
		panic(err)
	}
	if err := manager.D.Create(service+"ban", &BanStatus{}); err != nil {
		panic(err)
	}
	err := manager.D.Find(service, &c, "WHERE gid=0")
	if err != nil && !errors.Is(err, sql.ErrNullResult) {
		panic(err)
	}
	if err == nil && bits.RotateLeft64(uint64(c.Disable), 1)&1 == 1 {
		m.Options.DisableOnDefault = !m.Options.DisableOnDefault
	}
	manager.M[service] = m
	return m
}

// These methods share the checked writer with the admin API. Existing direct
// calls that ignore the new error return remain source-compatible.
func (m *Control[CTX]) Enable(gid int64) error { _, err := m.SetSwitch(gid, "enabled", ""); return err }
func (m *Control[CTX]) Disable(gid int64) error {
	_, err := m.SetSwitch(gid, "disabled", "")
	return err
}
func (m *Control[CTX]) Reset(gid int64) error { _, err := m.SetSwitch(gid, "inherit", ""); return err }

// Read and resolve under one lock so stale readers cannot reinstall old caches.
func (m *Control[CTX]) IsEnabledIn(gid int64) bool {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	if value, ok := m.Cache[gid]; ok {
		return value == 0
	}
	state, err := m.stateLocked(gid)
	if err != nil {
		log.Error("[control] cannot read plugin switch")
		return false
	}
	if state.EffectiveEnabled {
		m.Cache[gid] = 0
	} else {
		m.Cache[gid] = 1
	}
	return state.EffectiveEnabled
}

func (m *Control[CTX]) Handler(gid, uid int64) bool {
	if m.Manager.IsBlocked(uid) {
		return false
	}
	grp := gid
	if grp == 0 {
		grp = -uid
	}
	if !m.Manager.CanResponse(grp) || m.IsBannedIn(uid, grp) {
		return false
	}
	return m.IsEnabledIn(grp)
}

func (m *Control[CTX]) String() string { return m.Options.Help }
