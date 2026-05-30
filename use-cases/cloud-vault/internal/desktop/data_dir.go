package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// OpenDataDirectory opens the data directory in the system file manager.
func OpenDataDirectory(dataDir string) error {
	absPath, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	// Verify directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("data directory does not exist: %s", absPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", absPath)
	case "darwin":
		cmd = exec.Command("open", absPath)
	default: // linux and others
		cmd = exec.Command("xdg-open", absPath)
	}

	return cmd.Start()
}

// GetDataDirInfo returns information about the data directory.
type DataDirInfo struct {
	Path       string `json:"path"`
	Exists     bool   `json:"exists"`
	Size       int64  `json:"size"`
	FileCount  int    `json:"file_count"`
	IsWritable bool   `json:"is_writable"`
}

// GetDataDirInfo retrieves information about the data directory.
func GetDataDirInfo(dataDir string) (*DataDirInfo, error) {
	absPath, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	info := &DataDirInfo{
		Path: absPath,
	}

	// Check if directory exists
	stat, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return info, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat directory: %w", err)
	}

	info.Exists = stat.IsDir()
	if !info.Exists {
		return info, nil
	}

	// Calculate size and file count
	var size int64
	var count int
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	info.Size = size
	info.FileCount = count

	// Check if writable
	testFile := filepath.Join(absPath, ".write_test")
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
		info.IsWritable = true
	}

	return info, nil
}
