package glog

import (
	"context"
	"os"
)

// NoOpLogger silently discards all log messages.
// Fatal still calls os.Exit(1) to honour the fatal contract.
// Use this in tests or when logging must be fully disabled.
type NoOpLogger struct{}

// NewNoOpLogger returns a NoOpLogger that implements StructuredLogger.
func NewNoOpLogger() *NoOpLogger { return &NoOpLogger{} }

func (n *NoOpLogger) WithFields(_ Fields) StructuredLogger { return n }

func (n *NoOpLogger) Debug(_ string, _ ...Fields) {}
func (n *NoOpLogger) Info(_ string, _ ...Fields)  {}
func (n *NoOpLogger) Warn(_ string, _ ...Fields)  {}
func (n *NoOpLogger) Error(_ string, _ ...Fields) {}
func (n *NoOpLogger) Fatal(_ string, _ ...Fields) { os.Exit(1) }

func (n *NoOpLogger) DebugCtx(_ context.Context, _ string, _ ...Fields) {}
func (n *NoOpLogger) InfoCtx(_ context.Context, _ string, _ ...Fields)  {}
func (n *NoOpLogger) WarnCtx(_ context.Context, _ string, _ ...Fields)  {}
func (n *NoOpLogger) ErrorCtx(_ context.Context, _ string, _ ...Fields) {}
func (n *NoOpLogger) FatalCtx(_ context.Context, _ string, _ ...Fields) { os.Exit(1) }
