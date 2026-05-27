package domain

import "time"

type PodPhase string

const (
	PodPhaseUnknown   PodPhase = "unknown"
	PodPhasePending   PodPhase = "pending"
	PodPhaseRunning   PodPhase = "running"
	PodPhaseSucceeded PodPhase = "succeeded"
	PodPhaseFailed    PodPhase = "failed"
)

type PodSnapshot struct {
	Phase        PodPhase
	PodIP        string
	HostIP       string
	StartedAt    time.Time
	RestartCount int32
	DeletedAt    time.Time
}
