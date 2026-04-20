package mongo

import (
	"time"

	"workerfleet/internal/domain"
)

type ActiveTaskEmbeddedDoc struct {
	TaskID     string            `bson:"task_id"`
	TaskType   string            `bson:"task_type,omitempty"`
	Phase      string            `bson:"phase,omitempty"`
	PhaseName  string            `bson:"phase_name,omitempty"`
	StartedAt  time.Time         `bson:"started_at,omitempty"`
	UpdatedAt  time.Time         `bson:"updated_at,omitempty"`
	Metadata   map[string]string `bson:"metadata,omitempty"`
	StoredAt   time.Time         `bson:"stored_at,omitempty"`
	SchemaVers int               `bson:"schema_version"`
}

type WorkerSnapshotDoc struct {
	ID                  string                  `bson:"_id"`
	Namespace           string                  `bson:"namespace,omitempty"`
	PodName             string                  `bson:"pod_name,omitempty"`
	PodUID              string                  `bson:"pod_uid,omitempty"`
	NodeName            string                  `bson:"node_name,omitempty"`
	ContainerName       string                  `bson:"container_name,omitempty"`
	Image               string                  `bson:"image,omitempty"`
	Version             string                  `bson:"version,omitempty"`
	WorkerStatus        string                  `bson:"worker_status"`
	StatusReason        string                  `bson:"status_reason,omitempty"`
	ProcessAlive        bool                    `bson:"process_alive"`
	AcceptingTasks      bool                    `bson:"accepting_tasks"`
	LastSeenAt          time.Time               `bson:"last_seen_at,omitempty"`
	LastReadyAt         time.Time               `bson:"last_ready_at,omitempty"`
	LastHeartbeatAt     time.Time               `bson:"last_heartbeat_at,omitempty"`
	LastError           string                  `bson:"last_error,omitempty"`
	WorkerRestartCount  int32                   `bson:"worker_restart_count,omitempty"`
	PodPhase            string                  `bson:"pod_phase,omitempty"`
	PodIP               string                  `bson:"pod_ip,omitempty"`
	HostIP              string                  `bson:"host_ip,omitempty"`
	PodStartedAt        time.Time               `bson:"pod_started_at,omitempty"`
	PodRestartCount     int32                   `bson:"restart_count,omitempty"`
	PodDeletedAt        time.Time               `bson:"pod_deleted_at,omitempty"`
	ActiveTaskCount     int                     `bson:"active_task_count"`
	LastStatusChangedAt time.Time               `bson:"last_status_changed_at,omitempty"`
	UpdatedAt           time.Time               `bson:"updated_at,omitempty"`
	ActiveTasks         []ActiveTaskEmbeddedDoc `bson:"active_tasks,omitempty"`
	SchemaVersion       int                     `bson:"schema_version"`
}

func WorkerSnapshotDocFromDomain(snapshot domain.WorkerSnapshot, updatedAt time.Time) WorkerSnapshotDoc {
	activeTasks := make([]ActiveTaskEmbeddedDoc, 0, len(snapshot.ActiveTasks))
	for _, task := range snapshot.ActiveTasks {
		activeTasks = append(activeTasks, ActiveTaskEmbeddedDocFromDomain(task, updatedAt))
	}

	return WorkerSnapshotDoc{
		ID:                  string(snapshot.Identity.WorkerID),
		Namespace:           snapshot.Identity.Namespace,
		PodName:             snapshot.Identity.PodName,
		PodUID:              string(snapshot.Identity.PodUID),
		NodeName:            snapshot.Identity.NodeName,
		ContainerName:       snapshot.Identity.ContainerName,
		Image:               snapshot.Identity.Image,
		Version:             snapshot.Identity.Version,
		WorkerStatus:        string(snapshot.Status),
		StatusReason:        snapshot.StatusReason,
		ProcessAlive:        snapshot.Runtime.ProcessAlive,
		AcceptingTasks:      snapshot.Runtime.AcceptingTasks,
		LastSeenAt:          snapshot.Runtime.LastSeenAt,
		LastReadyAt:         snapshot.Runtime.LastReadyAt,
		LastHeartbeatAt:     snapshot.Runtime.LastHeartbeatAt,
		LastError:           snapshot.Runtime.LastError,
		WorkerRestartCount:  snapshot.Runtime.RestartCount,
		PodPhase:            string(snapshot.Pod.Phase),
		PodIP:               snapshot.Pod.PodIP,
		HostIP:              snapshot.Pod.HostIP,
		PodStartedAt:        snapshot.Pod.StartedAt,
		PodRestartCount:     snapshot.Pod.RestartCount,
		PodDeletedAt:        snapshot.Pod.DeletedAt,
		ActiveTaskCount:     len(snapshot.ActiveTasks),
		LastStatusChangedAt: snapshot.LastStatusChangedAt,
		UpdatedAt:           updatedAt,
		ActiveTasks:         activeTasks,
		SchemaVersion:       SchemaVersion,
	}
}

func ActiveTaskEmbeddedDocFromDomain(task domain.ActiveTask, storedAt time.Time) ActiveTaskEmbeddedDoc {
	return ActiveTaskEmbeddedDoc{
		TaskID:     string(task.TaskID),
		TaskType:   task.TaskType,
		Phase:      string(task.Phase),
		PhaseName:  task.PhaseName,
		StartedAt:  task.StartedAt,
		UpdatedAt:  task.UpdatedAt,
		Metadata:   cloneStringMap(task.Metadata),
		StoredAt:   storedAt,
		SchemaVers: SchemaVersion,
	}
}
