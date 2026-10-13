package control

// RestoredLocked invalidates control caches only after a database restore commits.
// The caller holds the Manager lock and supplies old/new persisted default bits.
func (manager *Manager[CTX]) RestoredLocked(old, next map[string]int64) {
	for name, c := range manager.M {
		if (uint64(old[name])^uint64(next[name]))>>63 != 0 {
			c.Options.DisableOnDefault = !c.Options.DisableOnDefault
		}
		c.Cache = make(map[int64]uint8, 16)
	}
	manager.banCache = make(map[uint64]bool, 8)
	manager.blockCache = make(map[int64]bool, 8)
	manager.respCache = make(map[int64]string, 8)
}
