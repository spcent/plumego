package mongo

import (
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func (doc WorkerSnapshotDoc) Domain() domain.WorkerSnapshot {
	tasks := make([]domain.ActiveTask, 0, len(doc.ActiveTasks))
	for _, task := range doc.ActiveTasks {
		tasks = append(tasks, task.Domain())
	}
	return domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:      domain.WorkerID(doc.ID),
			Namespace:     doc.Namespace,
			PodName:       doc.PodName,
			PodUID:        domain.PodUID(doc.PodUID),
			NodeName:      doc.NodeName,
			ContainerName: doc.ContainerName,
			Image:         doc.Image,
			Version:       doc.Version,
		},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:    doc.ProcessAlive,
			AcceptingTasks:  doc.AcceptingTasks,
			LastSeenAt:      doc.LastSeenAt,
			LastReadyAt:     doc.LastReadyAt,
			LastHeartbeatAt: doc.LastHeartbeatAt,
			LastError:       doc.LastError,
			RestartCount:    doc.WorkerRestartCount,
		},
		Pod: domain.PodSnapshot{
			Phase:        domain.PodPhase(doc.PodPhase),
			PodIP:        doc.PodIP,
			HostIP:       doc.HostIP,
			StartedAt:    doc.PodStartedAt,
			RestartCount: doc.PodRestartCount,
			DeletedAt:    doc.PodDeletedAt,
		},
		Status:              domain.WorkerStatus(doc.WorkerStatus),
		StatusReason:        doc.StatusReason,
		LastStatusChangedAt: doc.LastStatusChangedAt,
		ActiveTasks:         tasks,
		ActiveTaskCount:     len(tasks),
	}
}

func (doc ActiveTaskEmbeddedDoc) Domain() domain.ActiveTask {
	return domain.ActiveTask{
		TaskID:    domain.TaskID(doc.TaskID),
		TaskType:  doc.TaskType,
		Phase:     domain.TaskPhase(doc.Phase),
		PhaseName: doc.PhaseName,
		StartedAt: doc.StartedAt,
		UpdatedAt: doc.UpdatedAt,
		Metadata:  cloneStringMap(doc.Metadata),
	}
}

func (doc WorkerActiveTaskDoc) Domain() platformstore.CurrentTaskRecord {
	return platformstore.CurrentTaskRecord{
		WorkerID: domain.WorkerID(doc.WorkerID),
		Task: domain.ActiveTask{
			TaskID:    domain.TaskID(doc.TaskID),
			TaskType:  doc.TaskType,
			Phase:     domain.TaskPhase(doc.Phase),
			PhaseName: doc.PhaseName,
			StartedAt: doc.StartedAt,
			UpdatedAt: doc.UpdatedAt,
			Metadata:  cloneStringMap(doc.Metadata),
		},
	}
}
