package manager

func (m *Manager) setupHandlers() {
	m.handlers = make(map[string]cmdHandler)
	m.handlers["exit"] = m.doExit
}

func (m *Manager) doExit(name string, argsLine string, args ...string) {
	m.stopped = true
}
