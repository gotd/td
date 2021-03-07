package dcmanager

func (m *Manager) tryMigrate() (func(), bool) {
	m.migmux.Lock()
	defer m.migmux.Unlock()

	if !m.migrating {
		m.migrating = true
		onMigrationDone := func() {
			m.migmux.Lock()
			defer m.migmux.Unlock()

			m.migrating = false
			for _, cb := range m.callbacks {
				cb()
			}

			m.callbacks = nil
		}

		return onMigrationDone, true
	}

	done := make(chan struct{})
	waitFunc := func() { <-done }
	m.callbacks = append(m.callbacks, func() { close(done) })

	return waitFunc, false
}
