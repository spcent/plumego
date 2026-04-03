package gateway

import (
	"errors"
	"strings"
	"testing"
)

func TestProxyErrorMessage(t *testing.T) {
	inner := errors.New("connection refused")
	pe := NewProxyError("http://backend:8080", inner, 2)

	msg := pe.Error()
	if !strings.Contains(msg, "http://backend:8080") {
		t.Errorf("Error() missing backend URL: %q", msg)
	}
	if !strings.Contains(msg, "2") {
		t.Errorf("Error() missing attempt number: %q", msg)
	}
	if !strings.Contains(msg, "connection refused") {
		t.Errorf("Error() missing inner error: %q", msg)
	}
}

func TestProxyErrorUnwrap(t *testing.T) {
	inner := errors.New("timeout")
	pe := NewProxyError("http://backend:8080", inner, 0)

	if !errors.Is(pe, inner) {
		t.Error("errors.Is should find the wrapped error via Unwrap")
	}
}

func TestProxyErrorIs(t *testing.T) {
	pe := NewProxyError("http://backend:8080", ErrBackendTimeout, 1)

	if !errors.Is(pe, ErrBackendTimeout) {
		t.Error("errors.Is should match ErrBackendTimeout")
	}
	if errors.Is(pe, ErrNoHealthyBackends) {
		t.Error("should not match ErrNoHealthyBackends")
	}
}

func TestProxyErrorFields(t *testing.T) {
	pe := NewProxyError("http://x:1234", ErrBackendRefused, 3)
	if pe.Backend != "http://x:1234" {
		t.Errorf("Backend = %q", pe.Backend)
	}
	if pe.Attempt != 3 {
		t.Errorf("Attempt = %d, want 3", pe.Attempt)
	}
	if !errors.Is(pe.Err, ErrBackendRefused) {
		t.Errorf("Err = %v", pe.Err)
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrNoBackends,
		ErrNoHealthyBackends,
		ErrBackendTimeout,
		ErrBackendRefused,
		ErrWebSocketUpgrade,
		ErrInvalidConfig,
		ErrTooManyRetries,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Error("sentinel error is nil")
		}
		if err.Error() == "" {
			t.Errorf("sentinel error has empty message: %T", err)
		}
	}
}
