package glog

import (
	"bytes"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Initialize logging
	Init()
}

// Test helper functions

// captureOutput captures output into a buffer
func captureOutput(f func()) string {
	var buf bytes.Buffer
	std.mu.Lock()
	oldOutput := std.output
	std.output = &buf
	std.mu.Unlock()

	f()
	std.Flush()

	std.mu.Lock()
	std.output = oldOutput
	std.mu.Unlock()

	return buf.String()
}

// createTempDir creates a temporary directory
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "glog_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

// cleanupTempDir removes the temporary directory
func cleanupTempDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("Failed to cleanup temp dir: %v", err)
	}
}

// resetGlobalLogger resets the global logger to its default state
func resetGlobalLogger() {
	std.mu.Lock()
	defer std.mu.Unlock()

	// Close all files
	for _, file := range std.logFiles {
		if file != nil {
			file.Close()
		}
	}

	// Reset to default state
	std.level = INFO
	std.output = os.Stderr
	std.toStderr = true
	std.alsoToStderr = false
	std.verbosity = 0
	std.vmodulePatterns = nil
	std.logBacktraceAt = ""
	std.logFiles = make(map[Level]*os.File)
	std.logDir = ""
	std.program = filepath.Base(os.Args[0])
}

// TestBasicLogging verifies basic logging functions
func TestBasicLogging(t *testing.T) {
	resetGlobalLogger()

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name:     "Info",
			logFunc:  func() { Info("test info message") },
			expected: "test info message",
		},
		{
			name:     "Infof",
			logFunc:  func() { Infof("test %s message", "info") },
			expected: "test info message",
		},
		{
			name:     "Infoln",
			logFunc:  func() { Infoln("test", "info", "message") },
			expected: "test info message",
		},
		{
			name:     "Warning",
			logFunc:  func() { Warning("test warning message") },
			expected: "test warning message",
		},
		{
			name:     "Warningf",
			logFunc:  func() { Warningf("test %s message", "warning") },
			expected: "test warning message",
		},
		{
			name:     "Warningln",
			logFunc:  func() { Warningln("test", "warning", "message") },
			expected: "test warning message",
		},
		{
			name:     "Error",
			logFunc:  func() { Error("test error message") },
			expected: "test error message",
		},
		{
			name:     "Errorf",
			logFunc:  func() { Errorf("test %s message", "error") },
			expected: "test error message",
		},
		{
			name:     "Errorln",
			logFunc:  func() { Errorln("test", "error", "message") },
			expected: "test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(tt.logFunc)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}

			// Verify log format
			if !regexp.MustCompile(`[IWEF]\d{4} \d{2}:\d{2}:\d{2}\.\d{6} +\d+ \w+:\d+\]`).MatchString(output) {
				t.Errorf("Log format is incorrect: %q", output)
			}
		})
	}
}

// TestLogLevels ensures log level filtering works
func TestLogLevels(t *testing.T) {
	resetGlobalLogger()

	tests := []struct {
		name      string
		setLevel  Level
		logLevel  Level
		logFunc   func()
		shouldLog bool
	}{
		{"INFO level allows INFO", INFO, INFO, func() { Info("test") }, true},
		{"INFO level allows WARNING", INFO, WARNING, func() { Warning("test") }, true},
		{"INFO level allows ERROR", INFO, ERROR, func() { Error("test") }, true},
		{"WARNING level blocks INFO", WARNING, INFO, func() { Info("test") }, false},
		{"WARNING level allows WARNING", WARNING, WARNING, func() { Warning("test") }, true},
		{"WARNING level allows ERROR", WARNING, ERROR, func() { Error("test") }, true},
		{"ERROR level blocks INFO", ERROR, INFO, func() { Info("test") }, false},
		{"ERROR level blocks WARNING", ERROR, WARNING, func() { Warning("test") }, false},
		{"ERROR level allows ERROR", ERROR, ERROR, func() { Error("test") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.SetLevel(tt.setLevel)
			output := captureOutput(tt.logFunc)

			hasOutput := len(strings.TrimSpace(output)) > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, got output=%q", tt.shouldLog, output)
			}
		})
	}
}

// TestVerboseLogging checks verbose logging behavior
func TestVerboseLogging(t *testing.T) {
	resetGlobalLogger()

	tests := []struct {
		name      string
		verbosity int
		logLevel  int
		shouldLog bool
	}{
		{"V(0) with verbosity 0", 0, 0, true},
		{"V(1) with verbosity 0", 0, 1, false},
		{"V(1) with verbosity 1", 1, 1, true},
		{"V(2) with verbosity 1", 1, 2, false},
		{"V(2) with verbosity 2", 2, 2, true},
		{"V(5) with verbosity 10", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.SetVerbose(tt.verbosity)

			if V(tt.logLevel) != tt.shouldLog {
				t.Errorf("V(%d) with verbosity %d: expected %v", tt.logLevel, tt.verbosity, tt.shouldLog)
			}

			// Test VLog behavior
			output := captureOutput(func() {
				VLog(tt.logLevel, "verbose test message")
			})

			hasOutput := len(strings.TrimSpace(output)) > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("VLog(%d) with verbosity %d: expected shouldLog=%v, got output=%q",
					tt.logLevel, tt.verbosity, tt.shouldLog, output)
			}
		})
	}
}

// TestVmodule validates vmodule functionality
func TestVmodule(t *testing.T) {
	resetGlobalLogger()

	// Validate vmodule parsing
	std.parseVmodule("glog_test=2,other=1,pattern*=3")

	expected := []vmodulePattern{
		{"glog_test", 2},
		{"other", 1},
		{"pattern*", 3},
	}

	if len(std.vmodulePatterns) != len(expected) {
		t.Fatalf("Expected %d patterns, got %d", len(expected), len(std.vmodulePatterns))
	}

	for i, pattern := range std.vmodulePatterns {
		if pattern != expected[i] {
			t.Errorf("Pattern %d: expected %+v, got %+v", i, expected[i], pattern)
		}
	}

	// Test file matching
	testFile := "glog_test.go"
	verbosity := std.getVerbosityForFile(testFile)
	if verbosity != 2 {
		t.Errorf("Expected verbosity 2 for %s, got %d", testFile, verbosity)
	}

	// Unmatched files should use the global verbosity
	std.verbosity = 5
	testFile = "unmatched.go"
	verbosity = std.getVerbosityForFile(testFile)
	if verbosity != 5 {
		t.Errorf("Expected verbosity 5 for %s, got %d", testFile, verbosity)
	}
}

// TestLogBacktrace checks stack trace logging selection
func TestLogBacktrace(t *testing.T) {
	resetGlobalLogger()

	// Configure backtrace location (use a line in this file)
	std.logBacktraceAt = "glog_test.go:999" // Use a non-existent line to avoid triggering

	if !std.shouldLogBacktrace("glog_test.go", 999) {
		t.Error("Should log backtrace for matching file:line")
	}

	if std.shouldLogBacktrace("glog_test.go", 1000) {
		t.Error("Should not log backtrace for non-matching line")
	}

	if std.shouldLogBacktrace("other.go", 999) {
		t.Error("Should not log backtrace for non-matching file")
	}
}

// TestFileOutput ensures log output is written to files
func TestFileOutput(t *testing.T) {
	resetGlobalLogger()

	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Configure log directory
	std.logDir = tempDir
	std.program = "testapp"

	err := std.initLogFiles()
	if err != nil {
		t.Fatalf("Failed to initialize log files: %v", err)
	}
	defer std.Close()

	// Write logs at different levels
	Info("info message")
	Warning("warning message")
	Error("error message")

	std.Flush()

	// Verify files are created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	var logFiles []string
	var symlinks []string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".log") {
			logFiles = append(logFiles, file.Name())
		} else if strings.Contains(file.Name(), "testapp.") {
			symlinks = append(symlinks, file.Name())
		}
	}

	// Expect three log files (INFO, WARN, ERROR)
	if len(logFiles) < 3 {
		t.Errorf("Expected at least 3 log files, got %d: %v", len(logFiles), logFiles)
	}

	// Expect three symlinks
	if len(symlinks) < 3 {
		t.Errorf("Expected at least 3 symlinks, got %d: %v", len(symlinks), symlinks)
	}

	// Verify the INFO file contains all log entries
	infoFile := ""
	for _, file := range logFiles {
		if strings.Contains(file, "INFO") {
			infoFile = filepath.Join(tempDir, file)
			break
		}
	}

	if infoFile == "" {
		t.Fatal("INFO log file not found")
	}

	content, err := os.ReadFile(infoFile)
	if err != nil {
		t.Fatalf("Failed to read INFO log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "info message") {
		t.Error("INFO file should contain info message")
	}
	if !strings.Contains(contentStr, "warning message") {
		t.Error("INFO file should contain warning message")
	}
	if !strings.Contains(contentStr, "error message") {
		t.Error("INFO file should contain error message")
	}
}

// TestConcurrentLogging validates concurrent logging
func TestConcurrentLogging(t *testing.T) {
	resetGlobalLogger()

	var buf bytes.Buffer
	std.SetOutput(&buf)

	const numGoroutines = 10
	const numLogs = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogs; j++ {
				Infof("goroutine %d log %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Verify the number of log lines
	expectedLines := numGoroutines * numLogs
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify each line is in the expected format
	for i, line := range lines {
		if !strings.Contains(line, "goroutine") || !strings.Contains(line, "log") {
			t.Errorf("Line %d has incorrect format: %s", i, line)
		}
	}
}

// TestLoggerInstance validates a custom Logger instance
func TestLoggerInstance(t *testing.T) {
	logger := New()

	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetLevel(WARNING)

	// INFO level logs should be filtered out
	logger.Info("should not appear")
	if buf.Len() > 0 {
		t.Error("INFO log should be filtered when level is WARNING")
	}

	// WARNING level logs should be emitted
	logger.Warning("should appear")
	if buf.Len() == 0 {
		t.Error("WARNING log should not be filtered")
	}

	output := buf.String()
	if !strings.Contains(output, "should appear") {
		t.Errorf("Expected warning message in output: %q", output)
	}
}

// TestCopyStandardLogTo verifies redirecting the standard library logger
func TestCopyStandardLogTo(t *testing.T) {
	resetGlobalLogger()

	var buf bytes.Buffer
	std.SetOutput(&buf)

	CopyStandardLogTo(INFO)

	// Log using the standard library logger

	stdlog.Print("standard log message")

	output := buf.String()
	if !strings.Contains(output, "standard log message") {
		t.Errorf("Standard log message not found in output: %q", output)
	}
}

// TestFlagIntegration validates flag integration
func TestFlagIntegration(t *testing.T) {
	// Save original values
	origLogDir := *logDir
	origAlsoLogToStderr := *alsoLogToStderr
	origLogToStderr := *logToStderr
	origVerbosity := *verbosity
	origVmodule := *vmodule
	origLogBacktraceAt := *logBacktraceAt

	defer func() {
		*logDir = origLogDir
		*alsoLogToStderr = origAlsoLogToStderr
		*logToStderr = origLogToStderr
		*verbosity = origVerbosity
		*vmodule = origVmodule
		*logBacktraceAt = origLogBacktraceAt
	}()

	// Set flag values
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	*logDir = tempDir
	*alsoLogToStderr = true
	*logToStderr = false
	*verbosity = 2
	*vmodule = "test=3"
	*logBacktraceAt = "test.go:123"

	// Reinitialize
	resetGlobalLogger()
	Init()
	defer std.Close()

	// Verify settings are applied
	if std.logDir != tempDir {
		t.Errorf("Expected logDir %s, got %s", tempDir, std.logDir)
	}

	if !std.alsoToStderr {
		t.Error("Expected alsoToStderr to be true")
	}

	if std.verbosity != 2 {
		t.Errorf("Expected verbosity 2, got %d", std.verbosity)
	}

	if len(std.vmodulePatterns) == 0 {
		t.Error("Expected vmodule patterns to be parsed")
	}

	if std.logBacktraceAt != "test.go:123" {
		t.Errorf("Expected logBacktraceAt 'test.go:123', got %s", std.logBacktraceAt)
	}
}

// TestMultiWriter validates writing to multiple outputs
func TestMultiWriter(t *testing.T) {
	resetGlobalLogger()

	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Configure output to both files and stderr
	std.logDir = tempDir
	std.alsoToStderr = true
	std.program = "testapp"

	err := std.initLogFiles()
	if err != nil {
		t.Fatalf("Failed to initialize log files: %v", err)
	}
	defer std.Close()

	// Capture stderr output
	var stderrBuf bytes.Buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	go func() {
		io.Copy(&stderrBuf, r)
	}()

	Info("test multi writer message")
	std.Flush()

	w.Close()
	os.Stderr = oldStderr

	time.Sleep(100 * time.Millisecond) // Wait for the goroutine to finish

	// Verify stderr has output
	stderrOutput := stderrBuf.String()
	if !strings.Contains(stderrOutput, "test multi writer message") {
		t.Errorf("Message not found in stderr: %q", stderrOutput)
	}

	// Verify the file also has output
	files, _ := os.ReadDir(tempDir)
	var infoFile string
	for _, file := range files {
		if strings.Contains(file.Name(), "INFO") && strings.HasSuffix(file.Name(), ".log") {
			infoFile = filepath.Join(tempDir, file.Name())
			break
		}
	}

	if infoFile == "" {
		t.Fatal("INFO log file not found")
	}

	fileContent, err := os.ReadFile(infoFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(fileContent), "test multi writer message") {
		t.Errorf("Message not found in log file: %q", string(fileContent))
	}
}

// BenchmarkLogging measures basic logging performance
func BenchmarkLogging(b *testing.B) {
	resetGlobalLogger()
	std.SetOutput(io.Discard) // Discard output to focus on performance

	b.ResetTimer()

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Info("benchmark test message")
		}
	})

	b.Run("Infof", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Infof("benchmark test message %d", i)
		}
	})

	b.Run("VLog", func(b *testing.B) {
		std.SetVerbose(1)
		for i := 0; i < b.N; i++ {
			VLog(1, "benchmark verbose message")
		}
	})

	b.Run("VLogFiltered", func(b *testing.B) {
		std.SetVerbose(0) // This filters out VLog(1)
		for i := 0; i < b.N; i++ {
			VLog(1, "benchmark filtered message")
		}
	})
}

// BenchmarkConcurrentLogging measures concurrent logging performance
func BenchmarkConcurrentLogging(b *testing.B) {
	resetGlobalLogger()
	std.SetOutput(io.Discard)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info("concurrent benchmark message")
		}
	})
}

// BenchmarkBufferPoolEffectiveness demonstrates the memory optimization from using sync.Pool
func BenchmarkBufferPoolEffectiveness(b *testing.B) {
	resetGlobalLogger()
	std.SetOutput(io.Discard)

	// This benchmark measures the allocation rate when logging
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Info("benchmark message for buffer pool effectiveness test")
	}
}

// BenchmarkBufferPoolWithLargeMessages demonstrates buffer pool benefits with large messages
func BenchmarkBufferPoolWithLargeMessages(b *testing.B) {
	resetGlobalLogger()
	std.SetOutput(io.Discard)

	largeMessage := strings.Repeat("x", 1000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Info(largeMessage)
	}
}

// TestLnMethodFormatting ensures *ln methods correctly format arguments with spaces
func TestLnMethodFormatting(t *testing.T) {
	resetGlobalLogger()

	tests := []struct {
		name     string
		logFunc  func()
		args     []any
		expected string
	}{
		{
			name:     "Infoln with multiple strings",
			logFunc:  func() { Infoln("hello", "world", "test") },
			args:     []any{"hello", "world", "test"},
			expected: "hello world test",
		},
		{
			name:     "Warningln with mixed types",
			logFunc:  func() { Warningln("count:", 42, "status:", true) },
			args:     []any{"count:", 42, "status:", true},
			expected: "count: 42 status: true",
		},
		{
			name:     "Errorln with numbers",
			logFunc:  func() { Errorln(1, 2, 3, "sum") },
			args:     []any{1, 2, 3, "sum"},
			expected: "1 2 3 sum",
		},
		{
			name:     "Infoln with zero arguments",
			logFunc:  func() { Infoln() },
			args:     []any{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(tt.logFunc)

			// Check that the formatted output is present
			if !strings.Contains(output, tt.expected) {
				t.Errorf("%s: Expected output to contain %q, got %q", tt.name, tt.expected, output)
			}

			// For non-empty expected strings, ensure there's a newline at the end
			if tt.expected != "" {
				if !strings.HasSuffix(strings.TrimRight(output, "\n"), tt.expected) {
					t.Errorf("%s: Expected output to end with %q, got %q", tt.name, tt.expected, output)
				}
			}
		})
	}
}

// TestEdgeCases covers edge scenarios
func TestEdgeCases(t *testing.T) {
	resetGlobalLogger()

	// Test empty message
	output := captureOutput(func() {
		Info("")
	})
	if !strings.Contains(output, "I") { // At minimum the level marker should exist
		t.Error("Empty message should still produce log header")
	}

	// Test a very long message
	longMessage := strings.Repeat("a", 10000)
	output = captureOutput(func() {
		Info(longMessage)
	})
	if !strings.Contains(output, longMessage) {
		t.Error("Long message should be logged completely")
	}

	// Test special characters
	specialMessage := "Test Chinese characters\n\tSpecial chars"
	output = captureOutput(func() {
		Info(specialMessage)
	})
	if !strings.Contains(output, specialMessage) {
		t.Error("Special characters should be logged correctly")
	}
}

// TestErrorHandling validates error cases
func TestErrorHandling(t *testing.T) {
	resetGlobalLogger()

	// Test invalid directory
	std.logDir = "/invalid/nonexistent/directory"
	err := std.initLogFiles()
	if err == nil {
		t.Error("Should get error for invalid directory")
	}

	// Test invalid vmodule
	std.parseVmodule("invalid=abc,=123,normal=1")

	// Only the valid portion should be parsed
	validCount := 0
	for _, pattern := range std.vmodulePatterns {
		if pattern.pattern == "normal" && pattern.level == 1 {
			validCount++
		}
	}

	if validCount != 1 {
		t.Error("Should parse only valid vmodule patterns")
	}
}

// TestLogRotation validates log rotation functionality
func TestLogRotation(t *testing.T) {
	resetGlobalLogger()

	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Configure log directory and rotation settings
	std.logDir = tempDir
	std.program = "testapp"
	std.SetRotationConfig(RotationConfig{
		MaxSize:    1, // 1MB max size
		MaxAge:     30,
		MaxBackups: 5,
	})

	// Initialize log files
	err := std.initLogFiles()
	if err != nil {
		t.Fatalf("Failed to initialize log files: %v", err)
	}
	defer std.Close()

	// Write enough data to trigger rotation
	message := strings.Repeat("x", 500000) // 500KB message
	for i := 0; i < 5; i++ {               // Write 2.5MB total
		Info(message)
	}
	std.Flush()

	// Get all log files
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	// Count log files (should be more than 1 due to rotation)
	infoLogFiles := 0
	for _, file := range files {
		if strings.Contains(file.Name(), "INFO") && strings.HasSuffix(file.Name(), ".log") {
			infoLogFiles++
		}
	}

	if infoLogFiles < 2 {
		t.Errorf("Expected at least 2 INFO log files after rotation, got %d", infoLogFiles)
	}

	// Check that currentSize was reset after rotation
	std.mu.RLock()
	currentSize := std.currentSize[INFO]
	std.mu.RUnlock()

	if currentSize > int64(std.rotationConfig.MaxSize)*1024*1024 {
		t.Errorf("Current log size %d should be less than max size %d after rotation",
			currentSize, int64(std.rotationConfig.MaxSize)*1024*1024)
	}
}

func TestLogRetentionCleanup(t *testing.T) {
	resetGlobalLogger()

	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	std.logDir = tempDir
	std.program = "testapp"
	std.SetRotationConfig(RotationConfig{
		MaxAge:     1, // days
		MaxBackups: 1,
	})

	oldTime := time.Now().Add(-48 * time.Hour)
	oldName := fmt.Sprintf("%s.host.INFO.20200101-000000.1.log", std.program)
	oldPath := filepath.Join(tempDir, oldName)
	if err := os.WriteFile(oldPath, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create old log file: %v", err)
	}
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to update old log time: %v", err)
	}

	recentBase := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 2; i++ {
		name := fmt.Sprintf("%s.host.INFO.20220101-000000.%d.log", std.program, i)
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte("recent"), 0644); err != nil {
			t.Fatalf("failed to create recent log file: %v", err)
		}
		modTime := recentBase.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("failed to update recent log time: %v", err)
		}
	}

	if err := std.initLogFiles(); err != nil {
		t.Fatalf("Failed to initialize log files: %v", err)
	}
	defer std.Close()

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	infoLogs := 0
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "INFO") && strings.HasSuffix(entry.Name(), ".log") {
			infoLogs++
		}
	}

	if infoLogs > 2 {
		t.Fatalf("expected at most 2 INFO logs after cleanup, got %d", infoLogs)
	}

	if _, err := os.Stat(oldPath); err == nil {
		t.Fatalf("expected old log file to be removed: %s", oldPath)
	}
}

// TestRotationConfig verifies rotation configuration is applied correctly
func TestRotationConfig(t *testing.T) {
	resetGlobalLogger()

	// Set initial rotation config
	std.SetRotationConfig(RotationConfig{
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
	})

	std.mu.RLock()
	config := std.rotationConfig
	std.mu.RUnlock()

	if config.MaxSize != 10 {
		t.Errorf("Expected MaxSize 10, got %d", config.MaxSize)
	}
	if config.MaxAge != 7 {
		t.Errorf("Expected MaxAge 7, got %d", config.MaxAge)
	}
	if config.MaxBackups != 3 {
		t.Errorf("Expected MaxBackups 3, got %d", config.MaxBackups)
	}

	// Update rotation config
	std.SetRotationConfig(RotationConfig{
		MaxSize:    20,
		MaxAge:     14,
		MaxBackups: 5,
	})

	std.mu.RLock()
	config = std.rotationConfig
	std.mu.RUnlock()

	if config.MaxSize != 20 {
		t.Errorf("Expected MaxSize 20 after update, got %d", config.MaxSize)
	}
	if config.MaxAge != 14 {
		t.Errorf("Expected MaxAge 14 after update, got %d", config.MaxAge)
	}
	if config.MaxBackups != 5 {
		t.Errorf("Expected MaxBackups 5 after update, got %d", config.MaxBackups)
	}
}

// TestClose validates that the Close() method correctly closes log files and cleans up resources
func TestClose(t *testing.T) {
	resetGlobalLogger()

	// Configure file logging
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	std.logDir = tempDir
	std.program = "testapp"

	// Initialize log files
	err := std.initLogFiles()
	if err != nil {
		t.Fatalf("Failed to initialize log files: %v", err)
	}

	// Write some logs to ensure files are active
	Info("test message before close")
	Warning("test warning before close")
	Error("test error before close")

	// Verify log files are open
	std.mu.RLock()
	openFiles := len(std.logFiles)
	std.mu.RUnlock()

	if openFiles == 0 {
		t.Error("Expected log files to be open before Close()")
	}

	// Close the logger
	std.Close()

	// Verify log files are closed and removed from map
	std.mu.RLock()
	closedFiles := len(std.logFiles)
	std.mu.RUnlock()

	if closedFiles != 0 {
		t.Errorf("Expected no open log files after Close(), got %d", closedFiles)
	}

	// Try to log after close - should not panic but may not write to files
	defer func() {
		if r := recover(); r != nil {
			t.Error("Logging after Close() should not panic")
		}
	}()

	Info("test message after close")

	// Reopen and log again to ensure we can still use the logger after close
	err = std.initLogFiles()
	if err != nil {
		t.Fatalf("Failed to reinitialize log files after Close(): %v", err)
	}

	Info("test message after reopen")
	std.Flush()
	std.Close()
}
