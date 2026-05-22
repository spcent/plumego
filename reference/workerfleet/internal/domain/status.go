package domain

import (
	"fmt"
	"time"
)

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

type AlertPolicy struct {
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

func DefaultAlertPolicy() AlertPolicy {
	policy := DefaultStatusPolicy()
	return policy.AlertPolicy()
}

func (p StatusPolicy) AlertPolicy() AlertPolicy {
	return AlertPolicy{
		StageStuckAfter:       p.StageStuckAfter,
		RestartBurstThreshold: p.RestartBurstThreshold,
	}
}

func (p StatusPolicy) Validate() error {
	if p.StaleAfter <= 0 {
		return fmt.Errorf("stale_after must be greater than zero")
	}
	if p.OfflineAfter <= 0 {
		return fmt.Errorf("offline_after must be greater than zero")
	}
	if p.OfflineAfter <= p.StaleAfter {
		return fmt.Errorf("offline_after must be greater than stale_after")
	}
	if p.StageStuckAfter <= 0 {
		return fmt.Errorf("stage_stuck_after must be greater than zero")
	}
	if p.RestartBurstThreshold <= 0 {
		return fmt.Errorf("restart_burst_threshold must be greater than zero")
	}
	return nil
}

func (p AlertPolicy) Validate() error {
	if p.StageStuckAfter <= 0 {
		return fmt.Errorf("stage_stuck_after must be greater than zero")
	}
	if p.RestartBurstThreshold <= 0 {
		return fmt.Errorf("restart_burst_threshold must be greater than zero")
	}
	return nil
}
