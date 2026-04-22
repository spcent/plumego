package store

import "time"

const DefaultRetention = 7 * 24 * time.Hour

type RetentionResult struct {
	TaskHistoryPruned     int
	CaseStepHistoryPruned int
	WorkerEventsPruned    int
	AlertsPruned          int
}

type RetentionStore interface {
	ApplyRetention(now time.Time, retention time.Duration) RetentionResult
}
