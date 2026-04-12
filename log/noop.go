package log

import (
	"context"
	"os"
)

// discardLogger silently discards all log messages.
// Fatal/FatalCtx still call os.Exit(1) to honour the fatal contract.
type discardLogger struct{}

func newDiscardLogger() *discardLogger { return &discardLogger{} }

func (n *discardLogger) WithFields(_ Fields) StructuredLogger  { return n }
func (n *discardLogger) With(_ string, _ any) StructuredLogger { return n }

func (n *discardLogger) Debug(_ string, _ ...Fields) {}
func (n *discardLogger) Info(_ string, _ ...Fields)  {}
func (n *discardLogger) Warn(_ string, _ ...Fields)  {}
func (n *discardLogger) Error(_ string, _ ...Fields) {}
func (n *discardLogger) Fatal(_ string, _ ...Fields) {
	os.Exit(1)
}

func (n *discardLogger) DebugCtx(_ context.Context, _ string, _ ...Fields) {}
func (n *discardLogger) InfoCtx(_ context.Context, _ string, _ ...Fields)  {}
func (n *discardLogger) WarnCtx(_ context.Context, _ string, _ ...Fields)  {}
func (n *discardLogger) ErrorCtx(_ context.Context, _ string, _ ...Fields) {}
func (n *discardLogger) FatalCtx(_ context.Context, _ string, _ ...Fields) {
	os.Exit(1)
}
