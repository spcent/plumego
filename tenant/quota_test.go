package tenant

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryQuotaManager_Allow(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 10,
			TokensPerMinute:   100,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()

	// First request should be allowed
	result, err := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Requests: 1,
		Tokens:   10,
		Now:      time.Now().UTC(),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}
	if result.RemainingRequests != 9 {
		t.Errorf("expected 9 remaining requests, got %d", result.RemainingRequests)
	}
	if result.RemainingTokens != 90 {
		t.Errorf("expected 90 remaining tokens, got %d", result.RemainingTokens)
	}
}

func TestInMemoryQuotaManager_Exceed_Requests(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 2,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()
	now := time.Now().UTC()

	// First request - allowed
	result1, _ := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Requests: 1,
		Now:      now,
	})
	if !result1.Allowed {
		t.Errorf("first request should be allowed")
	}

	// Second request - allowed
	result2, _ := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Requests: 1,
		Now:      now,
	})
	if !result2.Allowed {
		t.Errorf("second request should be allowed")
	}

	// Third request - should be denied
	result3, err := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Requests: 1,
		Now:      now,
	})
	if err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	if result3.Allowed {
		t.Errorf("third request should be denied")
	}
	if result3.RetryAfter <= 0 {
		t.Errorf("RetryAfter should be positive, got %v", result3.RetryAfter)
	}
}

func TestInMemoryQuotaManager_Exceed_Tokens(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			TokensPerMinute: 50,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()
	now := time.Now().UTC()

	// Request with 30 tokens - allowed
	result1, _ := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Tokens: 30,
		Now:    now,
	})
	if !result1.Allowed {
		t.Errorf("first request should be allowed")
	}
	if result1.RemainingTokens != 20 {
		t.Errorf("expected 20 remaining tokens, got %d", result1.RemainingTokens)
	}

	// Request with 25 tokens - should be denied
	result2, err := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Tokens: 25,
		Now:    now,
	})
	if err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	if result2.Allowed {
		t.Errorf("second request should be denied (tokens exceeded)")
	}
}

func TestInMemoryQuotaManager_WindowReset(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 2,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()

	// First window
	now1 := time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC)
	quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{Requests: 1, Now: now1})
	quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{Requests: 1, Now: now1})

	// Third request should fail (quota exceeded in window 1)
	result, err := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{Requests: 1, Now: now1})
	if err != ErrQuotaExceeded {
		t.Errorf("expected quota exceeded in first window")
	}
	if result.Allowed {
		t.Errorf("request should be denied")
	}

	// New window (next minute)
	now2 := time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC)
	result2, err2 := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{Requests: 1, Now: now2})
	if err2 != nil {
		t.Errorf("expected no error in new window, got %v", err2)
	}
	if !result2.Allowed {
		t.Errorf("request should be allowed in new window")
	}
	if result2.RemainingRequests != 1 {
		t.Errorf("expected 1 remaining in new window, got %d", result2.RemainingRequests)
	}
}

func TestInMemoryQuotaManager_Unlimited(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "unlimited-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 0, // 0 = unlimited
			TokensPerMinute:   0, // 0 = unlimited
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()
	now := time.Now().UTC()

	// Make many requests - all should be allowed
	for i := 0; i < 100; i++ {
		result, err := quotaMgr.Allow(ctx, "unlimited-tenant", QuotaRequest{
			Requests: 1,
			Tokens:   1000,
			Now:      now,
		})
		if err != nil {
			t.Errorf("iteration %d: unexpected error: %v", i, err)
		}
		if !result.Allowed {
			t.Errorf("iteration %d: should be allowed (unlimited)", i)
		}
	}
}

func TestInMemoryQuotaManager_TokensOnly(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "tokens-only",
		Quota: QuotaConfig{
			RequestsPerMinute: 0,   // unlimited
			TokensPerMinute:   100, // limited
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()
	now := time.Now().UTC()

	// Make 10 requests with 10 tokens each - should all be allowed
	for i := 0; i < 10; i++ {
		result, err := quotaMgr.Allow(ctx, "tokens-only", QuotaRequest{
			Requests: 1,
			Tokens:   10,
			Now:      now,
		})
		if err != nil {
			t.Errorf("iteration %d: unexpected error: %v", i, err)
		}
		if !result.Allowed {
			t.Errorf("iteration %d: should be allowed", i)
		}
	}

	// Next request should fail (tokens exceeded)
	result, err := quotaMgr.Allow(ctx, "tokens-only", QuotaRequest{
		Requests: 1,
		Tokens:   10,
		Now:      now,
	})
	if err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	if result.Allowed {
		t.Errorf("should be denied (tokens exceeded)")
	}
}

func TestInMemoryQuotaManager_RequestsOnly(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "requests-only",
		Quota: QuotaConfig{
			RequestsPerMinute: 5, // limited
			TokensPerMinute:   0, // unlimited
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()
	now := time.Now().UTC()

	// Make 5 requests with varying tokens - should all be allowed
	for i := 0; i < 5; i++ {
		result, err := quotaMgr.Allow(ctx, "requests-only", QuotaRequest{
			Requests: 1,
			Tokens:   i * 1000, // varying tokens
			Now:      now,
		})
		if err != nil {
			t.Errorf("iteration %d: unexpected error: %v", i, err)
		}
		if !result.Allowed {
			t.Errorf("iteration %d: should be allowed", i)
		}
	}

	// 6th request should fail (requests exceeded)
	result, err := quotaMgr.Allow(ctx, "requests-only", QuotaRequest{
		Requests: 1,
		Tokens:   1000,
		Now:      now,
	})
	if err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	if result.Allowed {
		t.Errorf("should be denied (requests exceeded)")
	}
}

func TestInMemoryQuotaManager_RetryAfter(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "retry-test",
		Quota: QuotaConfig{
			RequestsPerMinute: 1,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()

	// Use current time for realistic RetryAfter calculation
	now := time.Now().UTC()

	// First request - allowed
	quotaMgr.Allow(ctx, "retry-test", QuotaRequest{Requests: 1, Now: now})

	// Second request - denied with RetryAfter
	result, err := quotaMgr.Allow(ctx, "retry-test", QuotaRequest{Requests: 1, Now: now})
	if err != ErrQuotaExceeded {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	if result.Allowed {
		t.Errorf("should be denied")
	}

	// RetryAfter should be positive and less than 60 seconds (within current minute)
	if result.RetryAfter <= 0 {
		t.Errorf("expected positive RetryAfter, got %v", result.RetryAfter)
	}
	if result.RetryAfter > 60*time.Second {
		t.Errorf("expected RetryAfter <= 60s, got %v", result.RetryAfter)
	}
}

func TestInMemoryQuotaManager_Concurrent(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "concurrent-test",
		Quota: QuotaConfig{
			RequestsPerMinute: 100,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()
	now := time.Now().UTC()

	var wg sync.WaitGroup
	allowedCount := int64(0)
	deniedCount := int64(0)
	var mu sync.Mutex

	// Concurrent requests
	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, _ := quotaMgr.Allow(ctx, "concurrent-test", QuotaRequest{
				Requests: 1,
				Now:      now,
			})
			mu.Lock()
			if result.Allowed {
				allowedCount++
			} else {
				deniedCount++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Should allow exactly 100 and deny 50
	if allowedCount != 100 {
		t.Errorf("expected 100 allowed, got %d", allowedCount)
	}
	if deniedCount != 50 {
		t.Errorf("expected 50 denied, got %d", deniedCount)
	}
}

func TestInMemoryQuotaManager_TenantNotFound(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()

	_, err := quotaMgr.Allow(ctx, "non-existent", QuotaRequest{
		Requests: 1,
		Now:      time.Now().UTC(),
	})

	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestInMemoryQuotaManager_NilProvider(t *testing.T) {
	quotaMgr := NewInMemoryQuotaManager(nil)
	ctx := context.Background()

	// Should allow all requests when provider is nil
	result, err := quotaMgr.Allow(ctx, "any-tenant", QuotaRequest{
		Requests: 1,
		Tokens:   100,
	})

	if err != nil {
		t.Errorf("expected no error with nil provider, got %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed with nil provider")
	}
}

func TestInMemoryQuotaManager_NilManager(t *testing.T) {
	var quotaMgr *InMemoryQuotaManager
	ctx := context.Background()

	// Should allow all requests when manager is nil
	result, err := quotaMgr.Allow(ctx, "any-tenant", QuotaRequest{
		Requests: 1,
	})

	if err != nil {
		t.Errorf("expected no error with nil manager, got %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed with nil manager")
	}
}

func TestInMemoryQuotaManager_DefaultRequestCount(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 10,
		},
	})

	quotaMgr := NewInMemoryQuotaManager(mgr)
	ctx := context.Background()

	// Request without specifying Requests (should default to 1)
	result, err := quotaMgr.Allow(ctx, "test-tenant", QuotaRequest{
		Requests: 0, // Will be set to 1
		Now:      time.Now().UTC(),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Errorf("should be allowed")
	}
	if result.RemainingRequests != 9 {
		t.Errorf("expected 9 remaining, got %d (default should count as 1)", result.RemainingRequests)
	}
}

func TestRemainingHelper(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		used     int
		expected int
	}{
		{"unlimited", 0, 10, -1},
		{"some remaining", 100, 30, 70},
		{"none remaining", 50, 50, 0},
		{"over limit", 10, 15, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := remaining(tt.limit, tt.used)
			if result != tt.expected {
				t.Errorf("remaining(%d, %d) = %d, want %d", tt.limit, tt.used, result, tt.expected)
			}
		})
	}
}
