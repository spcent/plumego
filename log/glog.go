package glog

import (
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	INFO Level = iota
	WARNING
	ERROR
	FATAL
)

var levelNames = []string{"INFO", "WARN", "ERROR", "FATAL"}

var (
	logDir          = flag.String("log_dir", "", "If non-empty, write log files in this directory")
	alsoLogToStderr = flag.Bool("alsologtostderr", false, "log to standard error as well as files")
	logToStderr     = flag.Bool("logtostderr", false, "log to standard error instead of files")
	verbosity       = flag.Int("v", 0, "log level for V logs")
	vmodule         = flag.String("vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	logBacktraceAt  = flag.String("log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")
)

// logBufferPool is used to recycle log buffers to reduce memory allocations
var logBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 1024) // Preallocate 1KB buffer
	},
}

type vmodulePattern struct {
	pattern string
	level   int
}

type RotationConfig struct {
	MaxSize    int // Maximum size in MB before rotation
	MaxAge     int // Maximum days to retain old log files
	MaxBackups int // Maximum number of old log files to retain
}

type Logger struct {
	mu              sync.RWMutex
	level           Level
	output          io.Writer
	toStderr        bool
	alsoToStderr    bool
	verbosity       int
	vmodulePatterns []vmodulePattern
	logBacktraceAt  string
	logFiles        map[Level]*os.File
	logDir          string
	program         string
	rotationConfig  RotationConfig
	currentSize     map[Level]int64
}

var std = New()

func New() *Logger {
	return &Logger{
		level:        INFO,
		output:       os.Stderr,
		toStderr:     true,
		alsoToStderr: false,
		verbosity:    0,
		logFiles:     make(map[Level]*os.File),
		currentSize:  make(map[Level]int64),
		program:      filepath.Base(os.Args[0]),
	}
}

func Init() {
	if !flag.Parsed() {
		flag.Parse()
	}

	std.mu.Lock()
	defer std.mu.Unlock()

	if *logDir != "" {
		std.logDir = *logDir
		if err := std.initLogFiles(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize log files: %v\n", err)
		}
	}

	std.toStderr = *logToStderr
	std.alsoToStderr = *alsoLogToStderr

	std.verbosity = *verbosity

	if *vmodule != "" {
		std.parseVmodule(*vmodule)
	}

	std.logBacktraceAt = *logBacktraceAt

	if std.toStderr {
		std.output = os.Stderr
	}
}

func (l *Logger) initLogFiles() error {
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return err
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	pid := os.Getpid()
	now := time.Now()

	for level := INFO; level <= ERROR; level++ {
		filename := fmt.Sprintf("%s.%s.%s.%04d%02d%02d-%02d%02d%02d.%d.log",
			l.program, hostname, levelNames[level],
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(), pid)

		logPath := filepath.Join(l.logDir, filename)
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		l.logFiles[level] = file

		linkName := fmt.Sprintf("%s.%s", l.program, levelNames[level])
		linkPath := filepath.Join(l.logDir, linkName)
		os.Remove(linkPath)
		if err := os.Symlink(filename, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink %s: %w", linkPath, err)
		}
	}

	return nil
}

func (l *Logger) parseVmodule(vmodule string) {
	l.vmodulePatterns = nil
	for _, item := range strings.Split(vmodule, ",") {
		if item == "" {
			continue
		}
		parts := strings.Split(item, "=")
		if len(parts) != 2 {
			continue
		}
		pattern := strings.TrimSpace(parts[0])
		levelStr := strings.TrimSpace(parts[1])
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			continue
		}
		l.vmodulePatterns = append(l.vmodulePatterns, vmodulePattern{
			pattern: pattern,
			level:   level,
		})
	}
}

func (l *Logger) getVerbosityForFile(file string) int {
	base := filepath.Base(file)
	base = strings.TrimSuffix(base, ".go")
	for _, pattern := range l.vmodulePatterns {
		if matched, _ := filepath.Match(pattern.pattern, base); matched {
			return pattern.level
		}
	}

	return l.verbosity
}

func (l *Logger) shouldLogBacktrace(file string, line int) bool {
	if l.logBacktraceAt == "" {
		return false
	}

	target := fmt.Sprintf("%s:%d", filepath.Base(file), line)
	return target == l.logBacktraceAt
}

func (l *Logger) formatHeader(level Level, file string, line int) []byte {
	buf := logBufferPool.Get().([]byte)
	buf = buf[:0] // Reset buffer to empty

	now := time.Now()
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	// Build header using buffer
	buf = append(buf, levelNames[level][0]) // First letter of level
	buf = append(buf, fmt.Sprintf("%02d%02d %02d:%02d:%02d.%06d %7d [%s:%d] ",
		now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000,
		os.Getpid(),
		file, line)...)

	return buf
}

func (l *Logger) getLogWriter(level Level) io.Writer {
	var writers []io.Writer
	if l.toStderr {
		return os.Stderr
	}

	if l.logDir != "" && l.logFiles[level] != nil {
		writers = append(writers, l.logFiles[level])
		for lv := INFO; lv < level; lv++ {
			if l.logFiles[lv] != nil {
				writers = append(writers, l.logFiles[lv])
			}
		}
	}

	if l.alsoToStderr {
		writers = append(writers, os.Stderr)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stderr)
	}

	if len(writers) == 1 {
		return writers[0]
	}

	return io.MultiWriter(writers...)
}

func (l *Logger) log(level Level, calldepth int, args ...any) {
	// First check if we should log at this level to avoid unnecessary work
	l.mu.RLock()
	shouldLog := level >= l.level
	l.mu.RUnlock()

	if !shouldLog {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 1
	}

	// Format header and message
	header := l.formatHeader(level, file, line)
	message := fmt.Sprint(args...)
	logLine := append(header, message...)
	logLine = append(logLine, '\n')

	// Get writer and write log
	l.mu.RLock()
	writer := l.getLogWriter(level)
	shouldBacktrace := l.shouldLogBacktrace(file, line)
	l.mu.RUnlock()

	fmt.Fprint(writer, string(logLine))

	// Return buffer to pool
	logBufferPool.Put(header[:0])

	// Check if log file needs rotation
	if l.logDir != "" {
		l.checkLogRotation(level, int64(len(logLine)))
	}

	// Handle backtrace if needed
	if shouldBacktrace {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		fmt.Fprint(writer, string(stack))
	}

	// Handle fatal errors
	if level == FATAL {
		l.Flush()
		os.Exit(1)
	}
}

func (l *Logger) logf(level Level, calldepth int, format string, args ...any) {
	// First check if we should log at this level to avoid unnecessary work
	l.mu.RLock()
	shouldLog := level >= l.level
	l.mu.RUnlock()

	if !shouldLog {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 1
	}

	// Format header and message
	header := l.formatHeader(level, file, line)
	message := fmt.Sprintf(format, args...)
	logLine := append(header, message...)
	logLine = append(logLine, '\n')

	// Get writer and write log
	l.mu.RLock()
	writer := l.getLogWriter(level)
	shouldBacktrace := l.shouldLogBacktrace(file, line)
	l.mu.RUnlock()

	fmt.Fprint(writer, string(logLine))

	// Return buffer to pool
	logBufferPool.Put(header[:0])

	// Check if log file needs rotation
	if l.logDir != "" {
		l.checkLogRotation(level, int64(len(logLine)))
	}

	// Handle backtrace if needed
	if shouldBacktrace {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		fmt.Fprint(writer, string(stack))
	}

	// Handle fatal errors
	if level == FATAL {
		l.Flush()
		os.Exit(1)
	}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

func (l *Logger) SetVerbose(v int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbosity = v
}

func (l *Logger) V(level int) bool {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return level <= l.verbosity
	}

	l.mu.RLock()
	verbosity := l.getVerbosityForFile(file)
	l.mu.RUnlock()

	return level <= verbosity
}

func (l *Logger) Info(args ...any) {
	l.log(INFO, 2, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(INFO, 2, format, args...)
}

func (l *Logger) Infoln(args ...any) {
	l.log(INFO, 2, fmt.Sprintln(args...))
}

func (l *Logger) Warning(args ...any) {
	l.log(WARNING, 2, args...)
}

func (l *Logger) Warningf(format string, args ...any) {
	l.logf(WARNING, 2, format, args...)
}

func (l *Logger) Warningln(args ...any) {
	l.log(WARNING, 2, fmt.Sprintln(args...))
}

func (l *Logger) Error(args ...any) {
	l.log(ERROR, 2, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(ERROR, 2, format, args...)
}

func (l *Logger) Errorln(args ...any) {
	l.log(ERROR, 2, fmt.Sprintln(args...))
}

func (l *Logger) Fatal(args ...any) {
	l.log(FATAL, 2, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.logf(FATAL, 2, format, args...)
}

func (l *Logger) Fatalln(args ...any) {
	l.log(FATAL, 2, fmt.Sprintln(args...))
}

func (l *Logger) VLog(level int, args ...any) {
	if l.V(level) {
		l.log(INFO, 2, args...)
	}
}

func (l *Logger) VLogf(level int, format string, args ...any) {
	if l.V(level) {
		l.logf(INFO, 2, format, args...)
	}
}

func (l *Logger) Flush() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, file := range l.logFiles {
		if file != nil {
			file.Sync()
		}
	}
}

// checkLogRotation checks if the log file needs rotation based on size
func (l *Logger) checkLogRotation(level Level, logSize int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	config := l.rotationConfig
	if config.MaxSize <= 0 {
		return nil
	}

	l.currentSize[level] += logSize
	if l.currentSize[level] < int64(config.MaxSize)*1024*1024 {
		return nil
	}

	// Reset current size and rotate log file
	l.currentSize[level] = 0
	return l.rotateLogFile(level)
}

// rotateLogFile rotates the log file for the given level
func (l *Logger) rotateLogFile(level Level) error {
	// Close the current log file
	if file, ok := l.logFiles[level]; ok {
		file.Close()
	}

	// Generate new log filename with timestamp
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	pid := os.Getpid()
	now := time.Now()

	filename := fmt.Sprintf("%s.%s.%s.%04d%02d%02d-%02d%02d%02d.%d.log",
		l.program, hostname, levelNames[level],
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), pid)

	logPath := filepath.Join(l.logDir, filename)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Update log file map
	l.logFiles[level] = file

	// Update symlink
	linkName := fmt.Sprintf("%s.%s", l.program, levelNames[level])
	linkPath := filepath.Join(l.logDir, linkName)
	os.Remove(linkPath)            // Ignore error
	os.Symlink(filename, linkPath) // Ignore error for backward compatibility

	return nil
}

// SetRotationConfig sets the log rotation configuration
func (l *Logger) SetRotationConfig(config RotationConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rotationConfig = config
}

// Close closes all log files and releases resources
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for level, file := range l.logFiles {
		if file != nil {
			file.Close()
			delete(l.logFiles, level)
		}
	}
}

func V(level int) bool {
	return std.V(level)
}

func Info(args ...any) {
	std.log(INFO, 2, args...)
}

func Infof(format string, args ...any) {
	std.logf(INFO, 2, format, args...)
}

func Infoln(args ...any) {
	std.log(INFO, 2, fmt.Sprintln(args...))
}

func Warning(args ...any) {
	std.log(WARNING, 2, args...)
}

func Warningf(format string, args ...any) {
	std.logf(WARNING, 2, format, args...)
}

func Warningln(args ...any) {
	std.log(WARNING, 2, fmt.Sprintln(args...))
}

func Error(args ...any) {
	std.log(ERROR, 2, args...)
}

func Errorf(format string, args ...any) {
	std.logf(ERROR, 2, format, args...)
}

func Errorln(args ...any) {
	std.log(ERROR, 2, fmt.Sprintln(args...))
}

func Fatal(args ...any) {
	std.log(FATAL, 2, args...)
}

func Fatalf(format string, args ...any) {
	std.logf(FATAL, 2, format, args...)
}

func Fatalln(args ...any) {
	std.log(FATAL, 2, fmt.Sprintln(args...))
}

func VLog(level int, args ...any) {
	if V(level) {
		std.log(INFO, 2, args...)
	}
}

func VLogf(level int, format string, args ...any) {
	if V(level) {
		std.logf(INFO, 2, format, args...)
	}
}

func Flush() {
	std.Flush()
}

func Close() {
	std.Close()
}

func CopyStandardLogTo(level Level) {
	stdlog.SetOutput(&logWriter{level: level})
	stdlog.SetFlags(0)
}

type logWriter struct {
	level Level
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	std.log(w.level, 3, msg) // Adjust calldepth to 3 for correct caller info
	return len(p), nil
}
