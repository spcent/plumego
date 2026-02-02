package health

import (
	"time"
)

// HealthHistoryEntry represents a single entry in health state history.
type HealthHistoryEntry struct {
	Timestamp  time.Time     `json:"timestamp"`
	State      HealthState   `json:"state"`
	Message    string        `json:"message"`
	Components []string      `json:"components,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

// HealthHistoryQuery represents a query for filtering health history.
type HealthHistoryQuery struct {
	StartTime *time.Time   `json:"start_time,omitempty"`
	EndTime   *time.Time   `json:"end_time,omitempty"`
	State     *HealthState `json:"state,omitempty"`
	Component string       `json:"component,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
}

// HealthHistoryQueryResult represents the result of a health history query.
type HealthHistoryQueryResult struct {
	Entries []HealthHistoryEntry `json:"entries"`
	Total   int                  `json:"total"`
	Limit   int                  `json:"limit"`
	Offset  int                  `json:"offset"`
	HasMore bool                 `json:"has_more"`
}

// GetHealthHistory returns the health history.
func (hm *healthManager) GetHealthHistory() []HealthHistoryEntry {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]HealthHistoryEntry, len(hm.history))
	copy(result, hm.history)

	return result
}

// QueryHealthHistory queries health history with filtering and pagination.
func (hm *healthManager) QueryHealthHistory(query HealthHistoryQuery) HealthHistoryQueryResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Filter entries based on query criteria
	var filtered []HealthHistoryEntry
	for _, entry := range hm.history {
		// Time range filter
		if query.StartTime != nil && entry.Timestamp.Before(*query.StartTime) {
			continue
		}
		if query.EndTime != nil && entry.Timestamp.After(*query.EndTime) {
			continue
		}

		// State filter
		if query.State != nil && entry.State != *query.State {
			continue
		}

		// Component filter
		if query.Component != "" {
			found := false
			for _, comp := range entry.Components {
				if comp == query.Component {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	// Calculate total before pagination
	total := len(filtered)

	// Apply pagination
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	// Ensure we don't exceed bounds
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	entries := make([]HealthHistoryEntry, 0)
	if offset < len(filtered) {
		entries = filtered[offset:end]
	}

	hasMore := end < len(filtered)

	return HealthHistoryQueryResult{
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}

// GetHealthHistoryStats returns statistics about the health history.
func (hm *healthManager) GetHealthHistoryStats() map[string]any {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.history) == 0 {
		return map[string]any{
			"total_entries":    0,
			"oldest_entry":     nil,
			"newest_entry":     nil,
			"entries_by_state": map[string]int{},
			"retention_config": hm.config,
		}
	}

	// Count entries by state
	stateCounts := make(map[string]int)
	for _, entry := range hm.history {
		stateCounts[string(entry.State)]++
	}

	oldestEntry := hm.history[0]
	newestEntry := hm.history[len(hm.history)-1]

	return map[string]any{
		"total_entries":    len(hm.history),
		"oldest_entry":     oldestEntry,
		"newest_entry":     newestEntry,
		"entries_by_state": stateCounts,
		"retention_config": hm.config,
		"time_span":        newestEntry.Timestamp.Sub(oldestEntry.Timestamp),
	}
}

// applyRetentionPolicy applies both time-based and count-based retention.
func (hm *healthManager) applyRetentionPolicy() {
	if len(hm.history) == 0 {
		return
	}

	// Apply time-based retention
	if hm.config.HistoryRetention > 0 {
		cutoffTime := time.Now().Add(-hm.config.HistoryRetention)
		keepIndex := 0
		for i, entry := range hm.history {
			if entry.Timestamp.After(cutoffTime) {
				keepIndex = i
				break
			}
			keepIndex = i + 1
		}
		if keepIndex < len(hm.history) {
			hm.history = hm.history[keepIndex:]
		}
	}

	// Apply count-based retention
	if hm.config.MaxHistoryEntries > 0 && len(hm.history) > hm.config.MaxHistoryEntries {
		hm.history = hm.history[len(hm.history)-hm.config.MaxHistoryEntries:]
	}
}

// cleanupHistory performs cleanup of old health history entries.
func (hm *healthManager) cleanupHistory() {
	hm.applyRetentionPolicy()
}

// ForceCleanup forces an immediate cleanup of health history.
func (hm *healthManager) ForceCleanup() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.cleanupHistory()
}

// getOverallState determines the overall health state from component healths.
func getOverallState(healths map[string]*ComponentHealth) HealthState {
	hasUnhealthy := false
	hasDegraded := false

	for _, health := range healths {
		switch health.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}
