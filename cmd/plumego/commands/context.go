package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// Context carries shared CLI dependencies and configuration.
type Context struct {
	Out     *output.Formatter
	EnvFile string
}

// resolveDir converts a directory flag to an absolute path and verifies it exists.
func resolveDir(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %w", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory not found: %s", absDir)
	}

	return absDir, nil
}
