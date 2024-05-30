package control

import (
	"math/bits"

	log "github.com/sirupsen/logrus"
)

// Flip 改变全局默认启用状态
func (m *Control[CTX]) Flip() error {
	var c GroupConfig
	m.Manager.Lock()
	defer m.Manager.Unlock()
	m.Options.DisableOnDefault = !m.Options.DisableOnDefault
	err := m.Manager.D.Find(m.Service, &c, "WHERE gid=0")
	if err != nil && m.Options.DisableOnDefault {
		c.Disable = 1
	}
	x := bits.RotateLeft64(uint64(c.Disable), 1) &^ 1
	c.Disable = int64(bits.RotateLeft64(x, -1))
	log.Debugf("[control] flip plugin %s of all : %d %v", m.Service, c.GroupID, x&1)
	err = m.Manager.D.Insert(m.Service, &c)
	if err != nil {
		log.Errorf("[control] %v", err)
	}
	return err
}
