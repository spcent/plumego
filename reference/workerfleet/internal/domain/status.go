package domain

import "time"

type WorkerID string
type TaskID string
type PodUID string

type WorkerStatus string

const (
	WorkerStatusUnknown  WorkerStatus = "unknown"
	WorkerStatusOnline   WorkerStatus = "online"
	WorkerStatusDegraded WorkerStatus = "degraded"
	WorkerStatusOffline  WorkerStatus = "offline"
)

type StatusPolicy struct {
	StaleAfter            time.Duration
	OfflineAfter          time.Duration
	StageStuckAfter       time.Duration
	RestartBurstThreshold int32
}

func DefaultStatusPolicy() StatusPolicy {
	return StatusPolicy{
		StaleAfter:            45 * time.Second,
		OfflineAfter:          90 * time.Second,
		StageStuckAfter:       10 * time.Minute,
		RestartBurstThreshold: 3,
	}
}
