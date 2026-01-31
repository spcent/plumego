package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type flakyChecker struct {
	name      string
	failCount int32
	attempts  int32
}

func (fc *flakyChecker) Name() string {
	return fc.name
}

func (fc *flakyChecker) Check(ctx context.Context) error {
	atomic.AddInt32(&fc.attempts, 1)
	if atomic.LoadInt32(&fc.attempts) <= fc.failCount {
		return errors.New("flaky failure")
	}
	return nil
}

type delayChecker struct {
	name  string
	delay time.Duration
}

func (dc *delayChecker) Name() string {
	return dc.name
}

func (dc *delayChecker) Check(ctx context.Context) error {
	if dc.delay <= 0 {
		return nil
	}
	select {
	case <-time.After(dc.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestCheckComponentRetry(t *testing.T) {
	config := HealthCheckConfig{
		RetryCount: 2,
		RetryDelay: 0,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	checker := &flakyChecker{name: "flaky", failCount: 1}
	if err := manager.RegisterComponent(checker); err != nil {
		t.Fatalf("failed to register component: %v", err)
	}

	if err := manager.CheckComponent(context.Background(), "flaky"); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}

	if attempts := atomic.LoadInt32(&checker.attempts); attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}

	health, ok := manager.GetComponentHealth("flaky")
	if !ok {
		t.Fatalf("expected component health")
	}
	attemptsVal, ok := health.Details["check_attempts"].(int)
	if !ok || attemptsVal != 2 {
		t.Fatalf("expected check_attempts to be 2, got %v", health.Details["check_attempts"])
	}
}

func TestCheckComponentTimeout(t *testing.T) {
	config := HealthCheckConfig{
		Timeout: 30 * time.Millisecond,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	checker := &delayChecker{name: "slow", delay: 100 * time.Millisecond}
	if err := manager.RegisterComponent(checker); err != nil {
		t.Fatalf("failed to register component: %v", err)
	}

	err = manager.CheckComponent(context.Background(), "slow")
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected deadline error, got %v", err)
	}

	health, ok := manager.GetComponentHealth("slow")
	if !ok {
		t.Fatalf("expected component health")
	}
	if health.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy status, got %v", health.Status)
	}
	if attemptsVal, ok := health.Details["check_attempts"].(int); !ok || attemptsVal != 1 {
		t.Fatalf("expected check_attempts to be 1, got %v", health.Details["check_attempts"])
	}
}
