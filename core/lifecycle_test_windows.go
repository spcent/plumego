//go:build windows

package core

import (
	"testing"
)

// sendShutdownSignal sends a shutdown signal to the current process
// On Windows, it skips the test as signal handling works differently
func sendShutdownSignal(t *testing.T) {
	t.Skip("skipping test on Windows due to signal handling differences")
}
