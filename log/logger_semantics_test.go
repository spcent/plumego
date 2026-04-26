package log

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDefaultLoggerRespectVerbosity(t *testing.T) {
	t.Run("respects-verbosity", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(LoggerConfig{
			Format:           LoggerFormatText,
			Output:           &buf,
			Level:            DEBUG,
			Verbosity:        0,
			RespectVerbosity: true,
		})
		logger.Debug("debug-suppressed")
		if got := buf.String(); got != "" {
			t.Fatalf("expected debug to be suppressed, got %q", got)
		}
	})

	t.Run("ignores-verbosity-when-disabled", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(LoggerConfig{
			Format:           LoggerFormatText,
			Output:           &buf,
			Level:            DEBUG,
			Verbosity:        0,
			RespectVerbosity: false,
		})
		logger.Debug("debug-ok")
		if got := buf.String(); !strings.Contains(got, "debug-ok") {
			t.Fatalf("expected debug to log, got %q", got)
		}
	})
}

func TestDerivedLoggerPreservesRespectVerbosityAcrossFormats(t *testing.T) {
	formats := []LoggerFormat{LoggerFormatText, LoggerFormatJSON}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(LoggerConfig{
				Format:           format,
				Output:           &buf,
				Level:            DEBUG,
				Verbosity:        0,
				RespectVerbosity: true,
			})

			child := logger.WithFields(Fields{"scope": "child"})
			child.Debug("debug-suppressed")
			if got := buf.String(); got != "" {
				t.Fatalf("expected derived logger to preserve verbosity gating, got %q", got)
			}

			child.Info("info-ok")
			got := buf.String()
			if !strings.Contains(got, "info-ok") {
				t.Fatalf("expected derived logger info to log, got %q", got)
			}
			if !strings.Contains(got, "child") {
				t.Fatalf("expected derived logger fields to be preserved, got %q", got)
			}
		})
	}
}

func TestDefaultLoggerErrorOutput(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Format:      LoggerFormatText,
		Output:      &out,
		ErrorOutput: &errOut,
		Level:       INFO,
	})

	logger.Info("info-msg")
	logger.Error("error-msg")

	if got := out.String(); !strings.Contains(got, "info-msg") {
		t.Fatalf("expected info log in output, got %q", got)
	}
	if got := out.String(); strings.Contains(got, "error-msg") {
		t.Fatalf("expected error log to skip output, got %q", got)
	}
	if got := errOut.String(); !strings.Contains(got, "error-msg") {
		t.Fatalf("expected error log in error output, got %q", got)
	}
}

func TestTextStructuredLoggerReportsExternalCaller(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Format: LoggerFormatText,
		Output: &buf,
		Level:  INFO,
	})

	callStructuredInfoForDepth(logger)

	got := buf.String()
	if !strings.Contains(got, "callsite_helper_test.go:") {
		t.Fatalf("expected structured text log to report external caller, got %q", got)
	}
	if strings.Contains(got, "logger.go:") || strings.Contains(got, "glog.go:") {
		t.Fatalf("expected structured text log to skip internal wrappers, got %q", got)
	}
}

func TestLoggerFatalExits(t *testing.T) {
	if os.Getenv("PLUMEGO_FATAL_CHILD") == "1" {
		format := LoggerFormat(os.Getenv("PLUMEGO_FATAL_FORMAT"))
		logger := NewLogger(LoggerConfig{
			Format:      format,
			Output:      io.Discard,
			ErrorOutput: io.Discard,
			Level:       DEBUG,
		})
		logger.Fatal("fatal-exit")
		return
	}

	formats := []LoggerFormat{LoggerFormatText, LoggerFormatJSON, LoggerFormatDiscard}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestLoggerFatalExits", "-test.v")
			cmd.Env = append(os.Environ(),
				"PLUMEGO_FATAL_CHILD=1",
				"PLUMEGO_FATAL_FORMAT="+string(format),
			)
			err := cmd.Run()
			exitErr, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected exit error, got %v", err)
			}
			if code := exitErr.ExitCode(); code != 1 {
				t.Fatalf("expected exit code 1, got %d", code)
			}
		})
	}
}
