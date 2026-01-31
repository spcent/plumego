//go:build !windows

package core

import (
	"syscall"
	"testing"
)

// sendShutdownSignal sends a shutdown signal to the current process
func sendShutdownSignal(t *testing.T) {
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
