package core

import "context"

// ShutdownHook runs during application shutdown.
type ShutdownHook func(context.Context) error

// OnShutdown registers a shutdown hook.
func (a *App) OnShutdown(hook ShutdownHook) error {
	if hook == nil {
		return nil
	}
	if err := a.ensureMutable("register_shutdown_hook", "register shutdown hook"); err != nil {
		return err
	}
	a.mu.Lock()
	a.shutdownHooks = append(a.shutdownHooks, hook)
	a.mu.Unlock()
	return nil
}

func (a *App) runShutdownHooks(ctx context.Context) {
	a.shutdownOnce.Do(func() {
		a.mu.Lock()
		hooks := append([]ShutdownHook{}, a.shutdownHooks...)
		a.mu.Unlock()

		for i := len(hooks) - 1; i >= 0; i-- {
			if hooks[i] == nil {
				continue
			}
			_ = hooks[i](ctx)
		}
	})
}
