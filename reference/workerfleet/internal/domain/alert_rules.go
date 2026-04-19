package domain

import "time"

func EvaluateSnapshotAlerts(snapshot WorkerSnapshot, now time.Time, policy StatusPolicy) []AlertRecord {
	candidates := make([]AlertRecord, 0, 4)

	switch snapshot.Status {
	case WorkerStatusOffline:
		candidates = append(candidates, newWorkerAlert(snapshot, AlertWorkerOffline, "error", "worker is offline", now))
	case WorkerStatusDegraded:
		candidates = append(candidates, newWorkerAlert(snapshot, AlertWorkerDegraded, "warning", "worker is degraded", now))
	}

	if snapshot.StatusReason == "heartbeat_stale" {
		candidates = append(candidates, newWorkerAlert(snapshot, AlertWorkerNoHeartbeat, "warning", "worker heartbeat is stale", now))
	}
	if snapshot.Status == WorkerStatusDegraded && !snapshot.Runtime.AcceptingTasks && len(snapshot.ActiveTasks) == 0 {
		candidates = append(candidates, newWorkerAlert(snapshot, AlertWorkerNotAccepting, "warning", "worker is not accepting tasks", now))
	}
	if hasStuckTask(snapshot.ActiveTasks, now, policy.StageStuckAfter) {
		candidates = append(candidates, newWorkerAlert(snapshot, AlertWorkerStageStuck, "error", "worker has a stuck task phase", now))
	}
	if policy.RestartBurstThreshold > 0 && snapshot.Pod.RestartCount >= policy.RestartBurstThreshold {
		candidates = append(candidates, newWorkerAlert(snapshot, AlertPodRestartBurst, "warning", "pod restart count exceeded threshold", now))
	}
	if !snapshot.Pod.DeletedAt.IsZero() || snapshot.Pod.Phase == PodPhaseFailed {
		candidates = append(candidates, newWorkerAlert(snapshot, AlertPodMissing, "error", "pod is unavailable", now))
	}

	return candidates
}

func EvaluateTaskConflictAlerts(snapshots []WorkerSnapshot, now time.Time) []AlertRecord {
	type owner struct {
		workerID WorkerID
		task     ActiveTask
	}

	owners := make(map[TaskID][]owner)
	for _, snapshot := range snapshots {
		for _, task := range snapshot.ActiveTasks {
			owners[task.TaskID] = append(owners[task.TaskID], owner{
				workerID: snapshot.Identity.WorkerID,
				task:     task,
			})
		}
	}

	alerts := make([]AlertRecord, 0, len(owners))
	for taskID, currentOwners := range owners {
		if len(currentOwners) < 2 {
			continue
		}
		alerts = append(alerts, AlertRecord{
			AlertID:   alertID(AlertTaskConflict, "", taskID, now),
			TaskID:    taskID,
			AlertType: AlertTaskConflict,
			Status:    AlertStatusFiring,
			Severity:  "error",
			DedupeKey: dedupeKey(AlertTaskConflict, "", taskID),
			Message:   "task is active on multiple workers",
			Details: map[string]string{
				"owner_count": intString(len(currentOwners)),
			},
			TriggeredAt: now,
		})
	}
	return alerts
}

func newWorkerAlert(snapshot WorkerSnapshot, alertType AlertType, severity string, message string, now time.Time) AlertRecord {
	return AlertRecord{
		AlertID:   alertID(alertType, snapshot.Identity.WorkerID, "", now),
		WorkerID:  snapshot.Identity.WorkerID,
		AlertType: alertType,
		Status:    AlertStatusFiring,
		Severity:  severity,
		DedupeKey: dedupeKey(alertType, snapshot.Identity.WorkerID, ""),
		Message:   message,
		Details: map[string]string{
			"status":        string(snapshot.Status),
			"status_reason": snapshot.StatusReason,
		},
		TriggeredAt: now,
	}
}
