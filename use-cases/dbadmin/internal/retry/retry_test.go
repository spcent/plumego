package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySuccess(t *testing.T) {
	cfg := Config{
		MaxAttempts:  3,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}

	attempts := 0
	err := Do(context.Background(), cfg, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("connection timeout")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryExhausted(t *testing.T) {
	cfg := Config{
		MaxAttempts:  2,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}

	attempts := 0
	err := Do(context.Background(), cfg, func() error {
		attempts++
		return errors.New("connection timeout")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if attempts != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryNonRetryable(t *testing.T) {
	cfg := Config{
		MaxAttempts:  3,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}

	attempts := 0
	err := Do(context.Background(), cfg, func() error {
		attempts++
		return errors.New("invalid credentials")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestRetryWithResult(t *testing.T) {
	cfg := Config{
		MaxAttempts:  3,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}

	attempts := 0
	result, err := WithResult(context.Background(), cfg, func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("connection reset")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	cfg := Config{
		MaxAttempts:  3,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, cfg, func() error {
		return errors.New("connection timeout")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestIsTransient(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"timeout", errors.New("connection timeout"), true},
		{"reset", errors.New("connection reset by peer"), true},
		{"refused", errors.New("connection refused"), true},
		{"unreachable", errors.New("network is unreachable"), true},
		{"credentials", errors.New("invalid credentials"), false},
		{"not found", errors.New("resource not found"), false},
		{"permission", errors.New("permission denied"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTransient(tt.err)
			if result != tt.expected {
				t.Errorf("IsTransient(%q) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
