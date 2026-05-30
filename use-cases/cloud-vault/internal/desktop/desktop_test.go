package desktop

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	plumelog "github.com/spcent/plumego/log"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.AppName == "" {
		t.Error("AppName should not be empty")
	}

	if cfg.DataDir == "" {
		t.Error("DataDir should not be empty")
	}

	if !cfg.Enabled {
		t.Error("Enabled should be true by default")
	}

	if !cfg.CloseToTray {
		t.Error("CloseToTray should be true by default")
	}

	if !cfg.NativeNotifications {
		t.Error("NativeNotifications should be true by default")
	}
}

func TestGetDefaultDataDir(t *testing.T) {
	dir := getDefaultDataDir()

	if dir == "" {
		t.Error("DataDir should not be empty")
	}

	// Check that the path is absolute
	if !filepath.IsAbs(dir) {
		t.Errorf("DataDir should be absolute path, got %s", dir)
	}
}

func TestNewService(t *testing.T) {
	cfg := DefaultConfig()
	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	if svc == nil {
		t.Fatal("Service should not be nil")
	}

	if svc.config != &cfg {
		t.Error("Service.config should match input config")
	}
}

func TestService_GetRuntimeInfo(t *testing.T) {
	cfg := DefaultConfig()
	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	ctx := context.Background()
	info, err := svc.GetRuntimeInfo(ctx)
	if err != nil {
		t.Fatalf("GetRuntimeInfo failed: %v", err)
	}

	if info == nil {
		t.Fatal("RuntimeInfo should not be nil")
	}

	if info.Mode != "desktop" {
		t.Errorf("Mode should be desktop, got %s", info.Mode)
	}

	if info.Version != "0.9.0" {
		t.Errorf("Version should be 0.9.0, got %s", info.Version)
	}

	if info.DataDir != cfg.DataDir {
		t.Errorf("DataDir should match config, got %s", info.DataDir)
	}

	if info.PID <= 0 {
		t.Error("PID should be positive")
	}
}

func TestService_ScanPreview(t *testing.T) {
	cfg := DefaultConfig()

	// Create a temporary directory with test files — and set DataDir to it
	// so the scan validation (source must be inside DataDir) passes.
	tmpDir := t.TempDir()
	cfg.DataDir = tmpDir

	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	// Create test markdown files
	testFiles := []string{
		"doc1.md",
		"doc2.md",
		"subdir/doc3.md",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("# Test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create a non-markdown file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("text"), 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	ctx := context.Background()

	t.Run("recursive scan", func(t *testing.T) {
		req := &ScanPreviewRequest{
			SourceDir: tmpDir,
			Recursive: true,
		}

		resp, err := svc.ScanPreview(ctx, req)
		if err != nil {
			t.Fatalf("ScanPreview failed: %v", err)
		}

		if resp.TotalFiles != 4 {
			t.Errorf("TotalFiles should be 4, got %d", resp.TotalFiles)
		}

		if resp.MarkdownFiles != 3 {
			t.Errorf("MarkdownFiles should be 3, got %d", resp.MarkdownFiles)
		}
	})

	t.Run("non-recursive scan", func(t *testing.T) {
		req := &ScanPreviewRequest{
			SourceDir: tmpDir,
			Recursive: false,
		}

		resp, err := svc.ScanPreview(ctx, req)
		if err != nil {
			t.Fatalf("ScanPreview failed: %v", err)
		}

		if resp.TotalFiles != 3 {
			t.Errorf("TotalFiles should be 3, got %d", resp.TotalFiles)
		}

		if resp.MarkdownFiles != 2 {
			t.Errorf("MarkdownFiles should be 2, got %d", resp.MarkdownFiles)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		req := &ScanPreviewRequest{
			SourceDir: "/nonexistent/path",
			Recursive: false,
		}

		_, err := svc.ScanPreview(ctx, req)
		if err == nil {
			t.Error("ScanPreview should fail for invalid path")
		}
	})
}

func TestService_GetDataDirInfo(t *testing.T) {
	cfg := DefaultConfig()
	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	// Create data directory
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	ctx := context.Background()
	info, err := svc.GetDataDirInfo(ctx)
	if err != nil {
		t.Fatalf("GetDataDirInfo failed: %v", err)
	}

	if info == nil {
		t.Fatal("DataDirInfo should not be nil")
	}

	if info.Path != cfg.DataDir {
		t.Errorf("Path should match config, got %s", info.Path)
	}

	if !info.Exists {
		t.Error("Exists should be true")
	}

	if !info.IsWritable {
		t.Error("IsWritable should be true")
	}
}

func TestNewTrayManager(t *testing.T) {
	cfg := DefaultConfig()
	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	tray := NewTrayManager(&cfg, nil, svc, logger)
	if tray == nil {
		t.Fatal("TrayManager should not be nil")
	}

	// Window starts visible by default
	if !tray.IsVisible() {
		t.Error("Window should be visible initially")
	}
}

func TestTrayManager_WindowVisibility(t *testing.T) {
	cfg := DefaultConfig()
	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	tray := NewTrayManager(&cfg, nil, svc, logger)

	tray.ShowWindow()
	if !tray.IsVisible() {
		t.Error("Window should be visible after ShowWindow")
	}

	tray.HideWindow()
	if tray.IsVisible() {
		t.Error("Window should not be visible after HideWindow")
	}
}

func TestTrayManager_CloseToTray(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CloseToTray = true
	logger := plumelog.NewLogger()

	svc := NewService(&cfg, nil, logger)

	tray := NewTrayManager(&cfg, nil, svc, logger)

	if !tray.ShouldCloseToTray() {
		t.Error("ShouldCloseToTray should return true when CloseToTray is enabled")
	}

	tray.config.CloseToTray = false
	if tray.ShouldCloseToTray() {
		t.Error("ShouldCloseToTray should return false when CloseToTray is disabled")
	}
}

func TestIsHidden(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{".hidden", true},
		{".git", true},
		{"visible", false},
		{"file.md", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHidden(tt.name)
			if result != tt.expected {
				t.Errorf("isHidden(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestGetPlatform(t *testing.T) {
	platform := getPlatform()

	if platform == "" {
		t.Error("Platform should not be empty")
	}

	// Platform should be one of: windows, darwin, linux
	validPlatforms := map[string]bool{
		"windows": true,
		"darwin":  true,
		"linux":   true,
	}

	if !validPlatforms[platform] {
		t.Errorf("Platform should be windows, darwin, or linux, got %s", platform)
	}
}
