package desktop

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config holds desktop application configuration.
type Config struct {
	Enabled             bool   `toml:"enabled"`
	AppName             string `toml:"app_name"`
	DataDir             string `toml:"data_dir"`
	CloseToTray         bool   `toml:"close_to_tray"`
	LaunchAtLogin       bool   `toml:"launch_at_login"`
	NativeNotifications bool   `toml:"native_notifications"`
	AutoUpdate          bool   `toml:"auto_update"`

	Import ImportConfig `toml:"import"`
}

// ImportConfig holds import-related configuration for desktop.
type ImportConfig struct {
	Recursive            bool `toml:"recursive"`
	AutoTagFromPath      bool `toml:"auto_tag_from_path"`
	ShowScanPreview      bool `toml:"show_scan_preview"`
	MaxFileSizeMB        int  `toml:"max_file_size_mb"`
	IgnoreHidden         bool `toml:"ignore_hidden"`
	IgnoreEmpty          bool `toml:"ignore_empty"`
	DefaultIndexAfter    bool `toml:"default_index_after"`
	DefaultSkipDuplicate bool `toml:"default_skip_duplicate"`
}

// DefaultConfig returns the default desktop configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:             true,
		AppName:             "Cloud Vault",
		DataDir:             getDefaultDataDir(),
		CloseToTray:         true,
		LaunchAtLogin:       false,
		NativeNotifications: true,
		AutoUpdate:          false,
		Import: ImportConfig{
			Recursive:            true,
			AutoTagFromPath:      true,
			ShowScanPreview:      true,
			MaxFileSizeMB:        50,
			IgnoreHidden:         true,
			IgnoreEmpty:          true,
			DefaultIndexAfter:    true,
			DefaultSkipDuplicate: true,
		},
	}
}

// getDefaultDataDir returns the default data directory based on platform.
func getDefaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "CloudVault")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "CloudVault")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "CloudVault")
	default: // linux and others
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, "cloudvault")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "cloudvault")
	}
}

// EnsureDataDir creates the data directory and all subdirectories if they don't exist.
func (c *Config) EnsureDataDir() error {
	dirs := []string{
		c.DataDir,
		filepath.Join(c.DataDir, "data"),
		filepath.Join(c.DataDir, "objects"),
		filepath.Join(c.DataDir, "backups"),
		filepath.Join(c.DataDir, "logs"),
		filepath.Join(c.DataDir, "cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
