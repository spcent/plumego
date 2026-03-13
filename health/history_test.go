package health

import (
	"testing"
	"time"
)

func TestHealthHistoryQuery(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Add some test history entries by simulating checks
	// We need to use the internal implementation to add history
	// For testing, we'll create a mock implementation
	now := time.Now()

	// Manually add history entries for testing
	hm := manager.(*healthManager)
	hm.mu.Lock()
	hm.history = []HealthHistoryEntry{
		{
			Timestamp:  now.Add(-2 * time.Hour),
			State:      StatusHealthy,
			Message:    "System healthy",
			Components: []string{"database", "cache"},
			Duration:   100 * time.Millisecond,
		},
		{
			Timestamp:  now.Add(-1 * time.Hour),
			State:      StatusDegraded,
			Message:    "High latency",
			Components: []string{"database"},
			Duration:   2 * time.Second,
		},
		{
			Timestamp:  now,
			State:      StatusHealthy,
			Message:    "System recovered",
			Components: []string{"database", "cache", "api"},
			Duration:   150 * time.Millisecond,
		},
	}
	hm.mu.Unlock()

	// Test querying all history
	result := manager.QueryHealthHistory(HealthHistoryQuery{})
	if len(result.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entries))
	}

	// Test filtering by time range
	startTime := now.Add(-90 * time.Minute)
	endTime := now.Add(30 * time.Minute)
	result = manager.QueryHealthHistory(HealthHistoryQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
	})
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries in time range, got %d", len(result.Entries))
	}

	// Test filtering by state
	healthyState := StatusHealthy
	result = manager.QueryHealthHistory(HealthHistoryQuery{
		State: &healthyState,
	})
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 healthy entries, got %d", len(result.Entries))
	}

	// Test filtering by component
	result = manager.QueryHealthHistory(HealthHistoryQuery{
		Component: "cache",
	})
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries with cache component, got %d", len(result.Entries))
	}

	// Test pagination
	result = manager.QueryHealthHistory(HealthHistoryQuery{
		Limit:  2,
		Offset: 0,
	})
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries with limit 2, got %d", len(result.Entries))
	}
	if !result.HasMore {
		t.Fatalf("expected HasMore to be true")
	}
}

func TestDataRetention(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Configure short retention for testing
	newConfig := HealthCheckConfig{
		MaxHistoryEntries:  3,
		HistoryRetention:   1 * time.Hour,
		AutoCleanupEnabled: false, // Disable auto cleanup for testing
	}
	manager.SetConfig(newConfig)

	// Add test entries
	hm := manager.(*healthManager)
	hm.mu.Lock()
	now := time.Now()
	hm.history = []HealthHistoryEntry{
		{Timestamp: now.Add(-3 * time.Hour), State: StatusHealthy, Message: "Old entry"},
		{Timestamp: now.Add(-2 * time.Hour), State: StatusHealthy, Message: "Recent entry 1"},
		{Timestamp: now.Add(-30 * time.Minute), State: StatusHealthy, Message: "Recent entry 2"},
		{Timestamp: now.Add(-10 * time.Minute), State: StatusHealthy, Message: "Recent entry 3"},
		{Timestamp: now, State: StatusHealthy, Message: "Current entry"},
	}
	hm.mu.Unlock()

	// Force cleanup
	manager.ForceCleanup()

	// Verify cleanup worked
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Should only have 3 most recent entries within retention period
	if len(hm.history) != 3 {
		t.Fatalf("expected 3 entries after cleanup, got %d", len(hm.history))
	}

	// Verify that old entries are removed
	for _, entry := range hm.history {
		if entry.Timestamp.Before(now.Add(-1 * time.Hour)) {
			t.Fatalf("found entry older than retention period: %v", entry.Timestamp)
		}
	}
}

func TestHealthHistoryConfigManagement(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Test initial default config
	cfg := manager.GetConfig()
	if cfg.MaxHistoryEntries != 100 {
		t.Fatalf("expected default MaxHistoryEntries to be 100, got %d", cfg.MaxHistoryEntries)
	}
	if cfg.HistoryRetention != 24*time.Hour {
		t.Fatalf("expected default HistoryRetention to be 24h, got %v", cfg.HistoryRetention)
	}
	if cfg.AutoCleanupEnabled {
		t.Fatalf("expected AutoCleanupEnabled to be false")
	}

	// Update config
	newConfig := HealthCheckConfig{
		MaxHistoryEntries:  50,
		HistoryRetention:   2 * time.Hour,
		AutoCleanupEnabled: true,
		CleanupInterval:    30 * time.Minute,
	}
	manager.SetConfig(newConfig)

	// Verify config was updated
	updatedConfig := manager.GetConfig()
	if updatedConfig.MaxHistoryEntries != 50 {
		t.Fatalf("expected MaxHistoryEntries to be 50, got %d", updatedConfig.MaxHistoryEntries)
	}
	if updatedConfig.HistoryRetention != 2*time.Hour {
		t.Fatalf("expected HistoryRetention to be 2h, got %v", updatedConfig.HistoryRetention)
	}
	if !updatedConfig.AutoCleanupEnabled {
		t.Fatalf("expected AutoCleanupEnabled to be true")
	}
}
