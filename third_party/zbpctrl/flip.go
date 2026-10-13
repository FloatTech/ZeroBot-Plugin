package control

// Flip persists the default-flip flag, preserving explicit switches and data.
func (m *Control[CTX]) Flip() error {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	c, mode, err := m.loadSwitchLocked(0)
	if err != nil {
		return err
	}
	if m.Manager.BeforeSwitch != nil {
		if err = m.Manager.BeforeSwitch(); err != nil {
			return err
		}
	}
	c.Disable = int64(uint64(c.Disable) ^ (uint64(1) << 63))
	if err = m.writeSwitchLocked(c, mode); err != nil {
		return err
	}
	m.Options.DisableOnDefault = !m.Options.DisableOnDefault
	m.Cache = make(map[int64]uint8, 16)
	return nil
}
