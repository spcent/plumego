package app

import (
	"testing"

	"mini-saas-api/internal/config"
)

// testConfig returns dev defaults with kv state isolated in a temp directory
// so tests never write into the package directory.
func testConfig(t *testing.T) config.Config {
	t.Helper()
	cfg := config.Defaults()
	cfg.App.DataDir = t.TempDir()
	return cfg
}
