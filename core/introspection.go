package core

// RuntimeSnapshot returns the stable runtime/config snapshot consumed by
// first-party tooling.
func (a *App) RuntimeSnapshot() RuntimeSnapshot {
	if a == nil || a.config == nil {
		return RuntimeSnapshot{}
	}

	a.mu.RLock()
	cfg := *a.config
	state := a.preparationState
	a.mu.RUnlock()

	snapshot := projectServerSettings(cfg).runtimeSnapshot()
	snapshot.PreparationState = state
	return snapshot
}
