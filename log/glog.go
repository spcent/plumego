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

type vmodulePattern struct {
	pattern string
	level   int
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
		os.Symlink(filename, linkPath)
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

func (l *Logger) formatHeader(level Level, file string, line int) string {
	now := time.Now()

	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s%02d%02d %02d:%02d:%02d.%06d %7d [%s:%d] ",
		levelNames[level][0:1], // Get the first letter
		now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000,
		os.Getpid(),
		file, line)
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
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 1
	}

	header := l.formatHeader(level, file, line)
	message := fmt.Sprint(args...)
	logLine := header + message + "\n"

	writer := l.getLogWriter(level)
	fmt.Fprint(writer, logLine)

	if l.shouldLogBacktrace(file, line) {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		fmt.Fprint(writer, string(stack))
	}

	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) logf(level Level, calldepth int, format string, args ...any) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 1
	}

	header := l.formatHeader(level, file, line)
	message := fmt.Sprintf(format, args...)
	logLine := header + message + "\n"

	writer := l.getLogWriter(level)
	fmt.Fprint(writer, logLine)

	if l.shouldLogBacktrace(file, line) {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		fmt.Fprint(writer, string(stack))
	}

	if level == FATAL {
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
	l.log(INFO, 2, args...)
}

func (l *Logger) Warning(args ...any) {
	l.log(WARNING, 2, args...)
}

func (l *Logger) Warningf(format string, args ...any) {
	l.logf(WARNING, 2, format, args...)
}

func (l *Logger) Warningln(args ...any) {
	l.log(WARNING, 2, args...)
}

func (l *Logger) Error(args ...any) {
	l.log(ERROR, 2, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(ERROR, 2, format, args...)
}

func (l *Logger) Errorln(args ...any) {
	l.log(ERROR, 2, args...)
}

func (l *Logger) Fatal(args ...any) {
	l.log(FATAL, 2, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.logf(FATAL, 2, format, args...)
}

func (l *Logger) Fatalln(args ...any) {
	l.log(FATAL, 2, args...)
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
	std.log(INFO, 2, args...)
}

func Warning(args ...any) {
	std.log(WARNING, 2, args...)
}

func Warningf(format string, args ...any) {
	std.logf(WARNING, 2, format, args...)
}

func Warningln(args ...any) {
	std.log(WARNING, 2, args...)
}

func Error(args ...any) {
	std.log(ERROR, 2, args...)
}

func Errorf(format string, args ...any) {
	std.logf(ERROR, 2, format, args...)
}

func Errorln(args ...any) {
	std.log(ERROR, 2, args...)
}

func Fatal(args ...any) {
	std.log(FATAL, 2, args...)
}

func Fatalf(format string, args ...any) {
	std.logf(FATAL, 2, format, args...)
}

func Fatalln(args ...any) {
	std.log(FATAL, 2, args...)
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
	std.log(w.level, 4, msg)
	return len(p), nil
}
