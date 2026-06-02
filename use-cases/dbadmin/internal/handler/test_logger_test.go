package handler

import (
	"context"

	plumelog "github.com/spcent/plumego/log"
)

type testLogger struct{}

func (testLogger) WithFields(plumelog.Fields) plumelog.StructuredLogger { return testLogger{} }
func (testLogger) Debug(string, ...plumelog.Fields)                     {}
func (testLogger) Info(string, ...plumelog.Fields)                      {}
func (testLogger) Warn(string, ...plumelog.Fields)                      {}
func (testLogger) Error(string, ...plumelog.Fields)                     {}
func (testLogger) Fatal(string, ...plumelog.Fields)                     {}
func (testLogger) DebugCtx(context.Context, string, ...plumelog.Fields) {}
func (testLogger) InfoCtx(context.Context, string, ...plumelog.Fields)  {}
func (testLogger) WarnCtx(context.Context, string, ...plumelog.Fields)  {}
func (testLogger) ErrorCtx(context.Context, string, ...plumelog.Fields) {}
func (testLogger) FatalCtx(context.Context, string, ...plumelog.Fields) {}
