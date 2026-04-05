package core

// RuntimeSnapshot returns the stable runtime/config snapshot consumed by
// first-party tooling.
func (a *App) RuntimeSnapshot() RuntimeSnapshot {
	if a == nil || a.config == nil {
		return RuntimeSnapshot{}
	}

	a.mu.RLock()
	cfg := *a.config
	configFrozen := a.configFrozen
	serverPrepared := a.httpServer != nil
	a.mu.RUnlock()

	snapshot := projectServerSettings(cfg).runtimeSnapshot()
	switch {
	case serverPrepared:
		snapshot.PreparationState = PreparationStateServerPrepared
	case configFrozen:
		snapshot.PreparationState = PreparationStateHandlerPrepared
	default:
		snapshot.PreparationState = PreparationStateMutable
	}
	return snapshot
}
