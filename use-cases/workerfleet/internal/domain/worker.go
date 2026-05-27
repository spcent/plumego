package domain

import "time"

type WorkerIdentity struct {
	WorkerID      WorkerID
	Namespace     string
	PodName       string
	PodUID        PodUID
	NodeName      string
	ContainerName string
	Image         string
	Version       string
}

type WorkerRuntime struct {
	ProcessAlive    bool
	AcceptingTasks  bool
	LastSeenAt      time.Time
	LastReadyAt     time.Time
	LastHeartbeatAt time.Time
	LastError       string
	RestartCount    int32
}

type WorkerSnapshot struct {
	Identity            WorkerIdentity
	Runtime             WorkerRuntime
	Pod                 PodSnapshot
	Status              WorkerStatus
	StatusReason        string
	LastStatusChangedAt time.Time
	ActiveTasks         []ActiveTask
	ActiveTaskCount     int
}

type WorkerReport struct {
	Identity       WorkerIdentity
	ProcessAlive   bool
	AcceptingTasks bool
	ObservedAt     time.Time
	LastError      string
	ActiveTasks    []TaskReport
}
