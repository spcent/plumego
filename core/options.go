package core

import (
	"github.com/spcent/plumego/log"
)

// AppDependencies carries constructor-owned dependencies for App.
type AppDependencies struct {
	// Logger used by request and lifecycle-adjacent helpers.
	Logger log.StructuredLogger
}

func resolveLogger(dependencies AppDependencies) log.StructuredLogger {
	if dependencies.Logger == nil {
		return log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard})
	}
	return dependencies.Logger
}
