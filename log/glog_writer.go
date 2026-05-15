package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// logBufferPool is used to recycle log buffers to reduce memory allocations
var logBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 1024) // Preallocate 1KB buffer
	},
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

	l.writeMu.Lock()
	defer l.writeMu.Unlock()

	// Snapshot writer state under writeMu so Close cannot close files between
	// selecting a writer and writing through it.
	l.mu.RLock()
	writer := l.getLogWriter(level)
	shouldBacktrace := l.shouldLogBacktrace(file, line)
	l.mu.RUnlock()

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

	// Handle fatal errors
	if level == FATAL {
		l.Flush()
		os.Exit(1)
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
