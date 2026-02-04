//go:build !windows

package core

import (
	"syscall"
	"testing"
)

// sendShutdownSignal sends a shutdown signal to the current process
func sendShutdownSignal(t *testing.T) {
	t.Helper()
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		t.Fatalf("failed to send SIGTERM signal: %v", err)
	}
}
