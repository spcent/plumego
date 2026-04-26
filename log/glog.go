package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DEBUG   Level = -1
	INFO    Level = 0
	WARNING Level = 1
	ERROR   Level = 2
	FATAL   Level = 3
)

var levelNames = map[Level]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARN",
	ERROR:   "ERROR",
	FATAL:   "FATAL",
}

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

type rotationConfig struct {
	MaxSize    int // Maximum size in MB before rotation
	MaxAge     int // Maximum days to retain old log files
	MaxBackups int // Maximum number of old log files to retain
}

type gLogger struct {
	mu              sync.RWMutex
	writeMu         sync.Mutex
	level           Level
	output          io.Writer
	errorOutput     io.Writer
	toStderr        bool
	alsoToStderr    bool
	verbosity       int
	vmodulePatterns []vmodulePattern
	logBacktraceAt  string
	logFiles        map[Level]*os.File
	logDir          string
	program         string
	rotationConfig  rotationConfig
	currentSize     map[Level]int64
	writeErrOnce    sync.Once
	// cachedWriters caches the per-level io.Writer (possibly an io.MultiWriter)
	// so getLogWriter does not allocate on every log call.
	// Invalidated whenever logFiles, alsoToStderr, toStderr, or output changes.
	cachedWriters map[Level]io.Writer
}

// pid is cached at package init to avoid a getpid syscall on every log line.
var pid = os.Getpid()

func newGLogger() *gLogger {
	return &gLogger{
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

func (l *gLogger) initLogFiles() error {
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return err
	}
	initialized := false
	openedLevels := make([]Level, 0, 3)
	defer func() {
		if initialized {
			return
		}
		for _, level := range openedLevels {
			if file := l.logFiles[level]; file != nil {
				_ = file.Close()
				delete(l.logFiles, level)
				delete(l.currentSize, level)
			}
		}
		l.rebuildWriterCache()
	}()

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	currentPID := pid
	now := time.Now()

	for level := INFO; level <= ERROR; level++ {
		lname := levelName(level)
		filename := logFilename(l.program, hostname, lname, now, currentPID)

		logPath := filepath.Join(l.logDir, filename)
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		l.logFiles[level] = file
		openedLevels = append(openedLevels, level)
		if info, err := file.Stat(); err == nil {
			l.currentSize[level] = info.Size()
		} else {
			l.currentSize[level] = 0
		}

		linkName := fmt.Sprintf("%s.%s", l.program, lname)
		linkPath := filepath.Join(l.logDir, linkName)
		os.Remove(linkPath)
		if err := os.Symlink(filename, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink %s: %w", linkPath, err)
		}
	}

	cleanupOldLogs(l.logDir, l.program, l.rotationConfig, currentLogFiles(l.logFiles))
	l.rebuildWriterCache()
	initialized = true
	return nil
}

func (l *gLogger) parseVmodule(vmodule string) {
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

func (l *gLogger) getVerbosityForFile(file string) int {
	base := filepath.Base(file)
	base = strings.TrimSuffix(base, ".go")
	for _, pattern := range l.vmodulePatterns {
		if matched, _ := filepath.Match(pattern.pattern, base); matched {
			return pattern.level
		}
	}

	return l.verbosity
}

func (l *gLogger) shouldLogBacktrace(file string, line int) bool {
	if l.logBacktraceAt == "" {
		return false
	}

	target := fmt.Sprintf("%s:%d", filepath.Base(file), line)
	return target == l.logBacktraceAt
}

func (l *gLogger) formatHeader(level Level, file string, line int) []byte {
	buf := logBufferPool.Get().([]byte)
	buf = buf[:0] // Reset buffer to empty

	now := time.Now()
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	// Build header using buffer
	buf = append(buf, levelInitial(level)) // First letter of level
	buf = append(buf, fmt.Sprintf("%02d%02d %02d:%02d:%02d.%06d %7d [%s:%d] ",
		now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000,
		pid,
		file, line)...)

	return buf
}

// rebuildWriterCache pre-computes the effective io.Writer for each log level
// and stores it in cachedWriters. Must be called under l.mu write lock whenever
// logFiles, alsoToStderr, toStderr, or output changes.
func (l *gLogger) rebuildWriterCache() {
	if l.toStderr {
		// All levels go to stderr; no need to cache per-level.
		l.cachedWriters = nil
		return
	}
	cache := make(map[Level]io.Writer, 4)
	for level := INFO; level <= FATAL; level++ {
		var writers []io.Writer
		if l.logDir != "" && l.logFiles[level] != nil {
			writers = append(writers, l.logFiles[level])
			for lv := INFO; lv < level; lv++ {
				if l.logFiles[lv] != nil {
					writers = append(writers, l.logFiles[lv])
				}
			}
		}
		if l.alsoToStderr {
			writers = append(writers, l.stderrWriter())
		}
		if len(writers) == 0 {
			writers = append(writers, l.stderrWriter())
		}
		if len(writers) == 1 {
			cache[level] = writers[0]
		} else {
			cache[level] = io.MultiWriter(writers...)
		}
	}
	l.cachedWriters = cache
}

// getLogWriter returns the effective io.Writer for the given level.
// Must be called with at least l.mu read lock held.
func (l *gLogger) getLogWriter(level Level) io.Writer {
	if level >= ERROR && l.errorOutput != nil {
		return l.errorOutput
	}
	if l.toStderr {
		return l.stderrWriter()
	}
	if l.cachedWriters != nil {
		if w, ok := l.cachedWriters[level]; ok {
			return w
		}
	}
	// Fallback (should not happen if rebuildWriterCache is called on every config change).
	return l.stderrWriter()
}

func (l *gLogger) stderrWriter() io.Writer {
	if l.output != nil {
		return l.output
	}
	return os.Stderr
}

func (l *gLogger) log(level Level, calldepth int, args ...any) {
	l.logInternal(level, calldepth+1, func() string {
		return fmt.Sprint(args...)
	})
}

func (l *gLogger) logf(level Level, calldepth int, format string, args ...any) {
	l.logInternal(level, calldepth+1, func() string {
		return fmt.Sprintf(format, args...)
	})
}

func (l *gLogger) logInternal(level Level, calldepth int, messageBuilder func() string) {
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
	message := messageBuilder()
	logLine := append(header, message...)
	logLine = append(logLine, '\n')

	// Get writer and write log
	l.mu.RLock()
	writer := l.getLogWriter(level)
	shouldBacktrace := l.shouldLogBacktrace(file, line)
	l.mu.RUnlock()

	l.writeMu.Lock()
	if err := writeFull(writer, logLine); err != nil {
		l.reportWriteError(err)
	}

	// Return buffer to pool
	logBufferPool.Put(header[:0])

	// Check if log file needs rotation
	if l.logDir != "" {
		for _, writtenLevel := range l.fileLevelsForLog(level) {
			_ = l.checkLogRotation(writtenLevel, int64(len(logLine)))
		}
	}

	// Handle backtrace if needed
	if shouldBacktrace {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		if err := writeFull(writer, stack); err != nil {
			l.reportWriteError(err)
		}
	}
	l.writeMu.Unlock()

	// Handle fatal errors
	if level == FATAL {
		l.Flush()
		os.Exit(1)
	}
}

func (l *gLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *gLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if w == nil {
		w = os.Stderr
	}
	l.output = w
	l.rebuildWriterCache()
}

func (l *gLogger) SetErrorOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorOutput = w
}

func (l *gLogger) SetVerbose(v int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbosity = v
}

func (l *gLogger) vAt(level int, calldepth int) bool {
	_, file, _, ok := runtime.Caller(calldepth)
	if !ok {
		return level <= l.verbosity
	}

	l.mu.RLock()
	verbosity := l.getVerbosityForFile(file)
	l.mu.RUnlock()

	return level <= verbosity
}

func (l *gLogger) V(level int) bool {
	return l.vAt(level, 2)
}

func (l *gLogger) Info(args ...any) {
	l.log(INFO, 2, args...)
}

func (l *gLogger) Infof(format string, args ...any) {
	l.logf(INFO, 2, format, args...)
}

func (l *gLogger) Infoln(args ...any) {
	l.log(INFO, 2, fmt.Sprintln(args...))
}

func (l *gLogger) Warning(args ...any) {
	l.log(WARNING, 2, args...)
}

func (l *gLogger) Warningf(format string, args ...any) {
	l.logf(WARNING, 2, format, args...)
}

func (l *gLogger) Warningln(args ...any) {
	l.log(WARNING, 2, fmt.Sprintln(args...))
}

func (l *gLogger) Error(args ...any) {
	l.log(ERROR, 2, args...)
}

func (l *gLogger) Errorf(format string, args ...any) {
	l.logf(ERROR, 2, format, args...)
}

func (l *gLogger) Errorln(args ...any) {
	l.log(ERROR, 2, fmt.Sprintln(args...))
}

func (l *gLogger) Fatal(args ...any) {
	l.log(FATAL, 2, args...)
}

func (l *gLogger) Fatalf(format string, args ...any) {
	l.logf(FATAL, 2, format, args...)
}

func (l *gLogger) Fatalln(args ...any) {
	l.log(FATAL, 2, fmt.Sprintln(args...))
}

func (l *gLogger) VLog(level int, args ...any) {
	if l.vAt(level, 2) {
		l.log(INFO, 2, args...)
	}
}

func (l *gLogger) VLogf(level int, format string, args ...any) {
	if l.vAt(level, 2) {
		l.logf(INFO, 2, format, args...)
	}
}

func (l *gLogger) fileLevelsForLog(level Level) []Level {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.logDir == "" || l.toStderr {
		return nil
	}

	levels := make([]Level, 0, 4)
	if l.logFiles[level] != nil {
		levels = append(levels, level)
	}
	for lv := INFO; lv < level; lv++ {
		if l.logFiles[lv] != nil {
			levels = append(levels, lv)
		}
	}
	return levels
}

func (l *gLogger) Flush() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, file := range l.logFiles {
		if file != nil {
			file.Sync()
		}
	}
}

// checkLogRotation checks if the log file needs rotation based on size
func (l *gLogger) checkLogRotation(level Level, logSize int64) error {
	l.mu.Lock()
	config := l.rotationConfig
	if config.MaxSize <= 0 {
		l.mu.Unlock()
		return nil
	}

	l.currentSize[level] += logSize
	if l.currentSize[level] < int64(config.MaxSize)*1024*1024 {
		l.mu.Unlock()
		return nil
	}

	// Reset current size and rotate log file
	l.currentSize[level] = 0
	err := l.rotateLogFile(level)
	logDir := l.logDir
	program := l.program
	current := currentLogFiles(l.logFiles)
	l.mu.Unlock()

	if err != nil {
		return err
	}

	cleanupOldLogs(logDir, program, config, current)
	return nil
}

// rotateLogFile rotates the log file for the given level
func (l *gLogger) rotateLogFile(level Level) error {
	// Close the current log file
	if file, ok := l.logFiles[level]; ok {
		file.Close()
	}

	// Generate new log filename with timestamp
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	currentPID := pid
	now := time.Now()

	lname := levelName(level)
	filename := logFilename(l.program, hostname, lname, now, currentPID)

	logPath := filepath.Join(l.logDir, filename)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Update log file map
	l.logFiles[level] = file
	l.currentSize[level] = 0

	// The log file changed, so cached writers for this level are stale.
	l.rebuildWriterCache()

	// Update symlink to point to the newest log file.
	// Symlink creation is best-effort: failures are logged to stderr but do not
	// abort the rotation because the log file itself was opened successfully.
	linkName := fmt.Sprintf("%s.%s", l.program, lname)
	linkPath := filepath.Join(l.logDir, linkName)
	os.Remove(linkPath) // ignore removal error; the next Symlink may still succeed
	if err := os.Symlink(filename, linkPath); err != nil {
		// Write directly to stderr to avoid recursion through the logger.
		fmt.Fprintf(os.Stderr, "log: failed to update symlink %s → %s: %v\n", linkPath, filename, err)
	}

	return nil
}

type logFileInfo struct {
	name    string
	path    string
	modTime time.Time
}

func logFilename(program, hostname, lname string, now time.Time, currentPID int) string {
	return fmt.Sprintf("%s.%s.%s.%04d%02d%02d-%02d%02d%02d.%09d.%d.log",
		program, hostname, lname,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond(), currentPID)
}

func currentLogFiles(files map[Level]*os.File) map[string]struct{} {
	current := make(map[string]struct{}, len(files))
	for _, file := range files {
		if file == nil {
			continue
		}
		current[filepath.Base(file.Name())] = struct{}{}
	}
	return current
}

func cleanupOldLogs(logDir, program string, config rotationConfig, current map[string]struct{}) {
	if logDir == "" {
		return
	}
	if config.MaxAge <= 0 && config.MaxBackups <= 0 {
		return
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	now := time.Now()
	perLevel := make(map[Level][]logFileInfo)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		name := entry.Name()
		if _, ok := current[name]; ok {
			continue
		}
		if !strings.HasSuffix(name, ".log") {
			continue
		}
		if !strings.HasPrefix(name, program+".") {
			continue
		}

		level, ok := levelFromFilename(name)
		if !ok {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if config.MaxAge > 0 {
			maxAge := time.Duration(config.MaxAge) * 24 * time.Hour
			if now.Sub(info.ModTime()) > maxAge {
				_ = os.Remove(filepath.Join(logDir, name))
				continue
			}
		}

		if config.MaxBackups > 0 {
			perLevel[level] = append(perLevel[level], logFileInfo{
				name:    name,
				path:    filepath.Join(logDir, name),
				modTime: info.ModTime(),
			})
		}
	}

	if config.MaxBackups <= 0 {
		return
	}

	for _, files := range perLevel {
		sort.Slice(files, func(i, j int) bool {
			return files[i].modTime.After(files[j].modTime)
		})
		for i := config.MaxBackups; i < len(files); i++ {
			_ = os.Remove(files[i].path)
		}
	}
}

func levelFromFilename(name string) (Level, bool) {
	for level := INFO; level <= ERROR; level++ {
		if strings.Contains(name, "."+levelName(level)+".") {
			return level, true
		}
	}
	return INFO, false
}

func levelName(level Level) string {
	if name, ok := levelNames[level]; ok {
		return name
	}
	return "INFO"
}

func levelInitial(level Level) byte {
	switch level {
	case DEBUG:
		return 'D'
	case INFO:
		return 'I'
	case WARNING:
		return 'W'
	case ERROR:
		return 'E'
	case FATAL:
		return 'F'
	default:
		return 'I'
	}
}

func (l *gLogger) reportWriteError(err error) {
	l.writeErrOnce.Do(func() {
		fmt.Fprintf(os.Stderr, "glog: failed to write log output: %v\n", err)
	})
}

func writeFull(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

// SetRotationConfig sets the log rotation configuration
func (l *gLogger) SetRotationConfig(config rotationConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rotationConfig = config
}

// Close closes all log files and releases resources
func (l *gLogger) Close() {
	l.writeMu.Lock()
	defer l.writeMu.Unlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	for level, file := range l.logFiles {
		if file != nil {
			_ = file.Close()
			delete(l.logFiles, level)
			delete(l.currentSize, level)
		}
	}
	l.rebuildWriterCache()
}
