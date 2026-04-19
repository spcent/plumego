package store

import "time"

const DefaultRetention = 7 * 24 * time.Hour

type RetentionResult struct {
	TaskHistoryPruned  int
	WorkerEventsPruned int
	AlertsPruned       int
}

type RetentionStore interface {
	ApplyRetention(now time.Time, retention time.Duration) RetentionResult
}

func (s *MemoryStore) ApplyRetention(now time.Time, retention time.Duration) RetentionResult {
	if retention <= 0 {
		retention = DefaultRetention
	}
	cutoff := now.Add(-retention)

	s.mu.Lock()
	defer s.mu.Unlock()

	result := RetentionResult{}
	for taskID, records := range s.taskHistory {
		filtered := records[:0]
		for _, record := range records {
			if latestTaskRecordTime(record).Before(cutoff) {
				result.TaskHistoryPruned++
				continue
			}
			filtered = append(filtered, record)
		}
		if len(filtered) == 0 {
			delete(s.taskHistory, taskID)
			continue
		}
		s.taskHistory[taskID] = filtered
	}

	filteredEvents := s.workerEvents[:0]
	for _, event := range s.workerEvents {
		if event.OccurredAt.Before(cutoff) {
			result.WorkerEventsPruned++
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}
	s.workerEvents = filteredEvents

	filteredAlerts := s.alerts[:0]
	for _, alert := range s.alerts {
		if latestAlertTime(alert).Before(cutoff) {
			result.AlertsPruned++
			continue
		}
		filteredAlerts = append(filteredAlerts, alert)
	}
	s.alerts = filteredAlerts

	return result
}

func latestTaskRecordTime(record TaskHistoryRecord) time.Time {
	if !record.EndedAt.IsZero() {
		return record.EndedAt
	}
	if !record.LastUpdatedAt.IsZero() {
		return record.LastUpdatedAt
	}
	return record.StartedAt
}

func latestAlertTime(alert AlertRecord) time.Time {
	if !alert.ResolvedAt.IsZero() {
		return alert.ResolvedAt
	}
	return alert.TriggeredAt
}

var _ RetentionStore = (*MemoryStore)(nil)
