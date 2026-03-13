package tenant

import (
	"context"
	"testing"
	"time"
)

func TestWindowQuotaManager_FallbackMinute(t *testing.T) {
	cfg := NewInMemoryConfigManager()
	cfg.SetTenantConfig(Config{
		TenantID: "t-1",
		Quota: QuotaConfig{
			RequestsPerMinute: 1,
		},
	})

	store := NewInMemoryQuotaStore()
	manager := NewWindowQuotaManager(cfg, store)
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	res, err := manager.Allow(context.Background(), "t-1", QuotaRequest{Requests: 1, Now: now})
	if err != nil || !res.Allowed {
		t.Fatalf("expected first request allowed, got err=%v res=%+v", err, res)
	}

	res, err = manager.Allow(context.Background(), "t-1", QuotaRequest{Requests: 1, Now: now})
	if err != ErrQuotaExceeded || res.Allowed {
		t.Fatalf("expected quota exceeded, got err=%v res=%+v", err, res)
	}
}

func TestWindowQuotaManager_MultiWindowRollback(t *testing.T) {
	cfg := NewInMemoryConfigManager()
	cfg.SetTenantConfig(Config{
		TenantID: "t-1",
		Quota: QuotaConfig{
			Limits: []QuotaLimit{
				{Window: QuotaWindowMinute, Requests: 2},
				{Window: QuotaWindowDay, Requests: 1},
			},
		},
	})

	store := NewInMemoryQuotaStore()
	manager := NewWindowQuotaManager(cfg, store)
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	res, err := manager.Allow(context.Background(), "t-1", QuotaRequest{Requests: 1, Now: now})
	if err != nil || !res.Allowed {
		t.Fatalf("expected first request allowed, got err=%v res=%+v", err, res)
	}

	res, err = manager.Allow(context.Background(), "t-1", QuotaRequest{Requests: 1, Now: now})
	if err != ErrQuotaExceeded || res.Allowed {
		t.Fatalf("expected quota exceeded, got err=%v res=%+v", err, res)
	}

	windowStart := quotaWindowStart(now, QuotaWindowMinute)
	usage, ok := store.Usage("t-1", QuotaWindowMinute, windowStart)
	if !ok {
		t.Fatalf("expected usage entry for minute window")
	}
	if usage.Requests != 1 {
		t.Fatalf("expected minute usage to stay at 1 after rollback, got %d", usage.Requests)
	}
}
