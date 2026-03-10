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

// HistoryStats summarises collected health history.
type HistoryStats struct {
	TotalEntries   int                 `json:"total_entries"`
	OldestEntry    *HealthHistoryEntry `json:"oldest_entry,omitempty"`
	NewestEntry    *HealthHistoryEntry `json:"newest_entry,omitempty"`
	EntriesByState map[string]int      `json:"entries_by_state"`
	TimeSpan       time.Duration       `json:"time_span,omitempty"`
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
		if query.StartTime != nil && entry.Timestamp.Before(*query.StartTime) {
			continue
		}
		if query.EndTime != nil && entry.Timestamp.After(*query.EndTime) {
			continue
		}
		if query.State != nil && entry.State != *query.State {
			continue
		}
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

	total := len(filtered)

	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	entries := make([]HealthHistoryEntry, 0)
	if offset < len(filtered) {
		entries = filtered[offset:end]
	}

	return HealthHistoryQueryResult{
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: end < len(filtered),
	}
}

// GetHealthHistoryStats returns statistics about the health history.
func (hm *healthManager) GetHealthHistoryStats() HistoryStats {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	stateCounts := make(map[string]int)

	if len(hm.history) == 0 {
		return HistoryStats{
			TotalEntries:   0,
			EntriesByState: stateCounts,
		}
	}

	for _, entry := range hm.history {
		stateCounts[string(entry.State)]++
	}

	oldest := hm.history[0]
	newest := hm.history[len(hm.history)-1]

	return HistoryStats{
		TotalEntries:   len(hm.history),
		OldestEntry:    &oldest,
		NewestEntry:    &newest,
		EntriesByState: stateCounts,
		TimeSpan:       newest.Timestamp.Sub(oldest.Timestamp),
	}
}

// applyRetentionPolicy applies both time-based and count-based retention.
func (hm *healthManager) applyRetentionPolicy() {
	if len(hm.history) == 0 {
		return
	}

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
	state, _ := calculateOverallStatus(healths)
	return state
}
