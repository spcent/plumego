package domain

import "time"

type EventType string

const (
	EventWorkerRegistered      EventType = "worker_registered"
	EventWorkerHeartbeat       EventType = "worker_heartbeat"
	EventWorkerOnline          EventType = "worker_online"
	EventWorkerDegraded        EventType = "worker_degraded"
	EventWorkerOffline         EventType = "worker_offline"
	EventWorkerReady           EventType = "worker_ready"
	EventWorkerNotReady        EventType = "worker_not_ready"
	EventTaskStarted           EventType = "task_started"
	EventTaskPhaseChanged      EventType = "task_phase_changed"
	EventTaskFinished          EventType = "task_finished"
	EventPodRestarted          EventType = "pod_restarted"
	EventPodDisappeared        EventType = "pod_disappeared"
	EventStateConflictDetected EventType = "state_conflict_detected"
)

type DomainEvent struct {
	Type       EventType
	OccurredAt time.Time
	WorkerID   WorkerID
	TaskID     TaskID
	Reason     string
	Attributes map[string]string
}
