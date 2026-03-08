package log

import (
	"context"
	"os"
)

// NoOpLogger silently discards all log messages.
// Fatal/FatalCtx call os.Exit(1) by default to honour the fatal contract.
// Set FatalHook to override this behaviour (e.g. in unit tests):
//
//	n := log.NewNoOpLogger()
//	n.FatalHook = func(msg string) { panic("fatal: " + msg) }
//
// Use NoOpLogger when logging must be fully disabled.
type NoOpLogger struct {
	// FatalHook is called instead of os.Exit(1) on Fatal/FatalCtx.
	// When nil, os.Exit(1) is used.
	FatalHook func(msg string)
}

// NewNoOpLogger returns a NoOpLogger that implements StructuredLogger.
func NewNoOpLogger() *NoOpLogger { return &NoOpLogger{} }

func (n *NoOpLogger) WithFields(_ Fields) StructuredLogger  { return n }
func (n *NoOpLogger) With(_ string, _ any) StructuredLogger { return n }

func (n *NoOpLogger) Debug(_ string, _ ...Fields) {}
func (n *NoOpLogger) Info(_ string, _ ...Fields)  {}
func (n *NoOpLogger) Warn(_ string, _ ...Fields)  {}
func (n *NoOpLogger) Error(_ string, _ ...Fields) {}
func (n *NoOpLogger) Fatal(msg string, _ ...Fields) {
	if n.FatalHook != nil {
		n.FatalHook(msg)
		return
	}
	os.Exit(1)
}

func (n *NoOpLogger) DebugCtx(_ context.Context, _ string, _ ...Fields) {}
func (n *NoOpLogger) InfoCtx(_ context.Context, _ string, _ ...Fields)  {}
func (n *NoOpLogger) WarnCtx(_ context.Context, _ string, _ ...Fields)  {}
func (n *NoOpLogger) ErrorCtx(_ context.Context, _ string, _ ...Fields) {}
func (n *NoOpLogger) FatalCtx(_ context.Context, msg string, _ ...Fields) {
	if n.FatalHook != nil {
		n.FatalHook(msg)
		return
	}
	os.Exit(1)
}
