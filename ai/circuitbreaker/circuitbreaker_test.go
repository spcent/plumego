package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second)

	// Should start closed
	if cb.State() != StateClosed {
		t.Errorf("Initial state = %v, want StateClosed", cb.State())
	}

	// First 2 failures keep it closed
	for i := 0; i < 2; i++ {
		err := cb.Execute(func() error {
			return errors.New("failure")
		})
		if err == nil {
			t.Error("Should return error")
		}
		if cb.State() != StateClosed {
			t.Errorf("State after %d failures = %v, want StateClosed", i+1, cb.State())
		}
	}

	// 3rd failure opens circuit
	cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.State() != StateOpen {
		t.Errorf("State after max failures = %v, want StateOpen", cb.State())
	}

	// Further requests fail fast
	err := cb.Execute(func() error {
		t.Error("Function should not be called when circuit is open")
		return nil
	})

	if err != ErrCircuitBreakerOpen {
		t.Errorf("Error = %v, want ErrCircuitBreakerOpen", err)
	}
}

func TestCircuitBreaker_OpenToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

	// Trip circuit
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return errors.New("fail") })

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Next request should transition to half-open
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Request in half-open should succeed, got error: %v", err)
	}

	// After success in half-open, should transition to closed
	if cb.State() != StateClosed {
		t.Errorf("State after successful half-open = %v, want StateClosed", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

	// Trip circuit
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return errors.New("fail") })

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Next request fails in half-open
	cb.Execute(func() error {
		return errors.New("fail")
	})

	// Should reopen circuit immediately
	if cb.State() != StateOpen {
		t.Errorf("State after failed half-open = %v, want StateOpen", cb.State())
	}
}

func TestCircuitBreaker_Success(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second)

	called := false
	err := cb.Execute(func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if !called {
		t.Error("Function should be called")
	}

	if cb.State() != StateClosed {
		t.Errorf("State = %v, want StateClosed", cb.State())
	}
}

func TestCircuitBreaker_Timeout(t *testing.T) {
	cb := NewCircuitBreakerWithConfig(Config{
		Name:         "test",
		MaxFailures:  1,
		Timeout:      50 * time.Millisecond,
		ResetTimeout: 1 * time.Second,
	})

	err := cb.Execute(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err != ErrTimeout {
		t.Errorf("Error = %v, want ErrTimeout", err)
	}

	if cb.State() != StateOpen {
		t.Errorf("State after timeout = %v, want StateOpen", cb.State())
	}
}

func TestCircuitBreaker_ExecuteWithContext(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second)

	ctx := context.Background()
	called := false

	err := cb.ExecuteWithContext(ctx, func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("ExecuteWithContext() error = %v, want nil", err)
	}

	if !called {
		t.Error("Function should be called")
	}
}

func TestCircuitBreaker_ExecuteWithContext_Cancelled(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := cb.ExecuteWithContext(ctx, func(ctx context.Context) error {
		return ctx.Err()
	})

	if err != context.Canceled {
		t.Errorf("Error = %v, want context.Canceled", err)
	}

	if cb.State() != StateOpen {
		t.Errorf("State = %v, want StateOpen", cb.State())
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker("test-breaker", 3, 1*time.Second)

	// Record some failures
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return errors.New("fail") })

	stats := cb.Stats()

	if stats.Name != "test-breaker" {
		t.Errorf("Name = %v, want 'test-breaker'", stats.Name)
	}

	if stats.State != StateClosed {
		t.Errorf("State = %v, want StateClosed", stats.State)
	}

	if stats.Failures != 2 {
		t.Errorf("Failures = %d, want 2", stats.Failures)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 1*time.Second)

	// Trip circuit
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return errors.New("fail") })

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("State after reset = %v, want StateClosed", cb.State())
	}

	stats := cb.Stats()
	if stats.Failures != 0 {
		t.Errorf("Failures after reset = %d, want 0", stats.Failures)
	}
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	var stateChanges []string
	var mu sync.Mutex

	cb := NewCircuitBreakerWithConfig(Config{
		Name:         "test",
		MaxFailures:  2,
		ResetTimeout: 1 * time.Second,
		OnStateChange: func(from, to State) {
			mu.Lock()
			defer mu.Unlock()
			stateChanges = append(stateChanges, from.String()+"->"+to.String())
		},
	})

	// Trip circuit
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return errors.New("fail") })

	// Wait for callback
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(stateChanges) != 1 || stateChanges[0] != "closed->open" {
		t.Errorf("State changes = %v, want ['closed->open']", stateChanges)
	}
	mu.Unlock()
}

func TestCircuitBreaker_Concurrent(t *testing.T) {
	cb := NewCircuitBreaker("test", 50, 1*time.Second)

	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex

	// 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			err := cb.Execute(func() error {
				if i%2 == 0 {
					return nil
				}
				return errors.New("fail")
			})

			mu.Lock()
			if err == nil {
				successCount++
			} else {
				failureCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successCount != 50 {
		t.Errorf("Success count = %d, want 50", successCount)
	}

	// Some failures should trigger circuit breaker
	if failureCount != 50 {
		t.Logf("Failure count = %d (some may have been blocked by circuit breaker)", failureCount)
	}
}

func TestNoOpCircuitBreaker(t *testing.T) {
	cb := &NoOpCircuitBreaker{}

	// Should always allow
	for i := 0; i < 100; i++ {
		err := cb.Execute(func() error {
			if i%2 == 0 {
				return errors.New("fail")
			}
			return nil
		})

		// NoOp just passes through the error
		if i%2 == 0 && err == nil {
			t.Error("Should return error")
		}
		if i%2 == 1 && err != nil {
			t.Error("Should not return error")
		}
	}

	// State should always be closed
	if cb.State() != StateClosed {
		t.Errorf("State = %v, want StateClosed", cb.State())
	}
}

func BenchmarkCircuitBreaker_Closed(b *testing.B) {
	cb := NewCircuitBreaker("test", 1000, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(func() error {
			return nil
		})
	}
}

func BenchmarkCircuitBreaker_Open(b *testing.B) {
	cb := NewCircuitBreaker("test", 1, 1*time.Second)

	// Trip circuit
	cb.Execute(func() error { return errors.New("fail") })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(func() error {
			b.Fatal("Should not be called")
			return nil
		})
	}
}

func BenchmarkCircuitBreaker_Concurrent(b *testing.B) {
	cb := NewCircuitBreaker("test", 10000, 1*time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() error {
				return nil
			})
		}
	})
}
