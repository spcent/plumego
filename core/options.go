package core

import (
	"github.com/spcent/plumego/log"
)

// WithLogger sets a custom logger for the App.
func WithLogger(logger log.StructuredLogger) Option {
	return func(a *App) {
		if logger == nil {
			panic("core logger cannot be nil")
		}
		a.logger = logger
	}
}
