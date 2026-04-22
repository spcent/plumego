package domain

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

func EvaluateWorkerStatus(snapshot WorkerSnapshot, now time.Time, policy StatusPolicy) (WorkerStatus, string) {
	if snapshot.Identity.WorkerID == "" {
		return WorkerStatusUnknown, "missing_worker_id"
	}

	if !snapshot.Pod.DeletedAt.IsZero() {
		return WorkerStatusOffline, "pod_deleted"
	}
	switch snapshot.Pod.Phase {
	case PodPhaseFailed:
		return WorkerStatusOffline, "pod_failed"
	case PodPhaseSucceeded:
		return WorkerStatusOffline, "pod_succeeded"
	}

	lastObserved := latestTime(snapshot.Runtime.LastHeartbeatAt, snapshot.Runtime.LastSeenAt)
	if lastObserved.IsZero() {
		return WorkerStatusUnknown, "awaiting_first_heartbeat"
	}

	if !snapshot.Runtime.ProcessAlive {
		return WorkerStatusOffline, "process_not_alive"
	}

	age := now.Sub(lastObserved)
	if policy.OfflineAfter > 0 && age > policy.OfflineAfter {
		return WorkerStatusOffline, "heartbeat_expired"
	}
	if policy.StaleAfter > 0 && age > policy.StaleAfter {
		return WorkerStatusDegraded, "heartbeat_stale"
	}
	if snapshot.Runtime.LastError != "" {
		return WorkerStatusDegraded, "worker_reported_error"
	}
	if hasStuckTask(snapshot.ActiveTasks, now, policy.StageStuckAfter) {
		return WorkerStatusDegraded, "task_stage_stuck"
	}
	if !snapshot.Runtime.AcceptingTasks && len(snapshot.ActiveTasks) == 0 {
		return WorkerStatusDegraded, "not_accepting_tasks"
	}
	if snapshot.Runtime.AcceptingTasks {
		return WorkerStatusOnline, "ready"
	}
	if len(snapshot.ActiveTasks) > 0 {
		return WorkerStatusOnline, "busy"
	}

	return WorkerStatusUnknown, "insufficient_signal"
}

func ReconcileActiveTasks(previous []ActiveTask, next []TaskReport, now time.Time, workerID WorkerID) ([]ActiveTask, []DomainEvent) {
	previousByID := make(map[TaskID]ActiveTask, len(previous))
	for _, task := range previous {
		if task.TaskID == "" {
			continue
		}
		previousByID[task.TaskID] = cloneActiveTask(task)
	}

	seen := make(map[TaskID]struct{}, len(next))
	nextByID := make(map[TaskID]TaskReport, len(next))
	events := make([]DomainEvent, 0, len(next)+len(previous))

	for _, report := range next {
		if report.TaskID == "" {
			events = append(events, DomainEvent{
				Type:       EventStateConflictDetected,
				OccurredAt: now,
				WorkerID:   workerID,
				Reason:     "missing_task_id",
			})
			continue
		}
		if _, exists := seen[report.TaskID]; exists {
			events = append(events, DomainEvent{
				Type:       EventStateConflictDetected,
				OccurredAt: now,
				WorkerID:   workerID,
				TaskID:     report.TaskID,
				Reason:     "duplicate_task_id_in_report",
			})
		}
		seen[report.TaskID] = struct{}{}
		nextByID[report.TaskID] = report
	}

	taskIDs := make([]TaskID, 0, len(nextByID))
	for taskID := range nextByID {
		taskIDs = append(taskIDs, taskID)
	}
	sort.Slice(taskIDs, func(i, j int) bool { return taskIDs[i] < taskIDs[j] })

	normalized := make([]ActiveTask, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		report := nextByID[taskID]
		previousTask, existed := previousByID[taskID]
		task := normalizeTaskReport(report, previousTask, now)
		normalized = append(normalized, task)

		if !existed {
			events = append(events, DomainEvent{
				Type:       EventTaskStarted,
				OccurredAt: task.UpdatedAt,
				WorkerID:   workerID,
				TaskID:     task.TaskID,
				Reason:     "task_added_to_active_set",
			})
			events = append(events, reconcileTaskStepEvents(ActiveTask{}, task, task.UpdatedAt, workerID)...)
			continue
		}
		if previousTask.Phase != task.Phase || previousTask.PhaseName != task.PhaseName {
			events = append(events, DomainEvent{
				Type:       EventTaskPhaseChanged,
				OccurredAt: task.UpdatedAt,
				WorkerID:   workerID,
				TaskID:     task.TaskID,
				Reason:     "task_phase_changed",
				Attributes: map[string]string{
					"from_phase":      string(previousTask.Phase),
					"to_phase":        string(task.Phase),
					"from_phase_name": previousTask.PhaseName,
					"to_phase_name":   task.PhaseName,
				},
			})
		}
		events = append(events, reconcileTaskStepEvents(previousTask, task, task.UpdatedAt, workerID)...)
	}

	for _, previousTask := range previous {
		if _, stillActive := nextByID[previousTask.TaskID]; stillActive {
			continue
		}
		if hasCaseStep(previousTask.CurrentStep) {
			events = append(events, taskStepFinishedEvent(previousTask, previousTask.CurrentStep, now, workerID, "task_removed_from_active_set"))
		}
		events = append(events, DomainEvent{
			Type:       EventTaskFinished,
			OccurredAt: now,
			WorkerID:   workerID,
			TaskID:     previousTask.TaskID,
			Reason:     "task_removed_from_active_set",
			Attributes: map[string]string{
				"final_phase":      string(previousTask.Phase),
				"final_phase_name": previousTask.PhaseName,
			},
		})
	}

	return normalized, events
}

func MergeWorkerReport(snapshot WorkerSnapshot, report WorkerReport, now time.Time, policy StatusPolicy) (WorkerSnapshot, []DomainEvent) {
	merged := snapshot
	events := make([]DomainEvent, 0, len(report.ActiveTasks)+6)
	observedAt := report.ObservedAt
	if observedAt.IsZero() {
		observedAt = now
	}

	wasRegistered := snapshot.Identity.WorkerID != ""
	merged.Identity = mergeIdentity(snapshot.Identity, report.Identity)
	if !wasRegistered && merged.Identity.WorkerID != "" {
		events = append(events, DomainEvent{
			Type:       EventWorkerRegistered,
			OccurredAt: observedAt,
			WorkerID:   merged.Identity.WorkerID,
			Reason:     "first_worker_report",
		})
	}

	previousAccepting := snapshot.Runtime.AcceptingTasks
	merged.Runtime.ProcessAlive = report.ProcessAlive
	merged.Runtime.AcceptingTasks = report.AcceptingTasks
	merged.Runtime.LastSeenAt = observedAt
	merged.Runtime.LastHeartbeatAt = observedAt
	merged.Runtime.LastError = strings.TrimSpace(report.LastError)
	if report.ProcessAlive && report.AcceptingTasks {
		merged.Runtime.LastReadyAt = observedAt
	}

	tasks, taskEvents := ReconcileActiveTasks(snapshot.ActiveTasks, report.ActiveTasks, observedAt, merged.Identity.WorkerID)
	merged.ActiveTasks = tasks
	merged.ActiveTaskCount = len(tasks)
	events = append(events, taskEvents...)
	events = append(events, DomainEvent{
		Type:       EventWorkerHeartbeat,
		OccurredAt: observedAt,
		WorkerID:   merged.Identity.WorkerID,
		Reason:     "worker_report_received",
	})

	if previousAccepting != merged.Runtime.AcceptingTasks {
		eventType := EventWorkerNotReady
		reason := "worker_stopped_accepting_tasks"
		if merged.Runtime.AcceptingTasks {
			eventType = EventWorkerReady
			reason = "worker_accepting_tasks"
		}
		events = append(events, DomainEvent{
			Type:       eventType,
			OccurredAt: observedAt,
			WorkerID:   merged.Identity.WorkerID,
			Reason:     reason,
		})
	}

	merged, statusEvents := applyStatus(snapshot, merged, now, policy)
	events = append(events, statusEvents...)

	return merged, events
}

func MergePodSnapshot(snapshot WorkerSnapshot, pod PodSnapshot, now time.Time, policy StatusPolicy) (WorkerSnapshot, []DomainEvent) {
	merged := snapshot
	events := make([]DomainEvent, 0, 4)

	if pod.RestartCount > snapshot.Pod.RestartCount {
		events = append(events, DomainEvent{
			Type:       EventPodRestarted,
			OccurredAt: now,
			WorkerID:   snapshot.Identity.WorkerID,
			Reason:     "pod_restart_count_increased",
			Attributes: map[string]string{
				"from_restart_count": int32String(snapshot.Pod.RestartCount),
				"to_restart_count":   int32String(pod.RestartCount),
			},
		})
	}
	if snapshot.Pod.DeletedAt.IsZero() && !pod.DeletedAt.IsZero() {
		events = append(events, DomainEvent{
			Type:       EventPodDisappeared,
			OccurredAt: pod.DeletedAt,
			WorkerID:   snapshot.Identity.WorkerID,
			Reason:     "pod_deleted",
		})
	}

	merged.Pod = pod
	if pod.RestartCount > merged.Runtime.RestartCount {
		merged.Runtime.RestartCount = pod.RestartCount
	}

	merged, statusEvents := applyStatus(snapshot, merged, now, policy)
	events = append(events, statusEvents...)

	return merged, events
}

func applyStatus(previous WorkerSnapshot, current WorkerSnapshot, now time.Time, policy StatusPolicy) (WorkerSnapshot, []DomainEvent) {
	status, reason := EvaluateWorkerStatus(current, now, policy)
	current.Status = status
	current.StatusReason = reason

	if current.ActiveTaskCount == 0 && len(current.ActiveTasks) > 0 {
		current.ActiveTaskCount = len(current.ActiveTasks)
	}

	if previous.Status == status {
		if current.LastStatusChangedAt.IsZero() {
			current.LastStatusChangedAt = previous.LastStatusChangedAt
		}
		return current, nil
	}

	current.LastStatusChangedAt = now
	if previous.Status == "" {
		return current, nil
	}

	eventType := EventWorkerDegraded
	switch status {
	case WorkerStatusOnline:
		eventType = EventWorkerOnline
	case WorkerStatusOffline:
		eventType = EventWorkerOffline
	case WorkerStatusDegraded:
		eventType = EventWorkerDegraded
	default:
		return current, nil
	}

	return current, []DomainEvent{{
		Type:       eventType,
		OccurredAt: now,
		WorkerID:   current.Identity.WorkerID,
		Reason:     reason,
		Attributes: map[string]string{
			"from_status": string(previous.Status),
			"to_status":   string(status),
		},
	}}
}

func normalizeTaskReport(report TaskReport, previous ActiveTask, observedAt time.Time) ActiveTask {
	task := ActiveTask{
		TaskID:      report.TaskID,
		ExecPlanID:  report.ExecPlanID,
		TaskType:    strings.TrimSpace(report.TaskType),
		Phase:       report.Phase,
		PhaseName:   strings.TrimSpace(report.PhaseName),
		CurrentStep: normalizeCaseStepReport(report.CurrentStep, previous.CurrentStep, observedAt),
		StartedAt:   report.StartedAt,
		UpdatedAt:   report.UpdatedAt,
		Metadata:    cloneMetadata(report.Metadata),
	}
	if task.Phase == "" {
		task.Phase = TaskPhaseUnknown
	}
	if task.PhaseName == "" {
		task.PhaseName = string(task.Phase)
	}
	if task.StartedAt.IsZero() {
		task.StartedAt = previous.StartedAt
	}
	if task.StartedAt.IsZero() {
		task.StartedAt = observedAt
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = observedAt
	}
	if task.TaskType == "" {
		task.TaskType = previous.TaskType
	}
	if task.ExecPlanID == "" {
		task.ExecPlanID = previous.ExecPlanID
	}
	if len(task.Metadata) == 0 && len(previous.Metadata) > 0 {
		task.Metadata = cloneMetadata(previous.Metadata)
	}
	return task
}

func normalizeCaseStepReport(report CaseStepRuntime, previous CaseStepRuntime, observedAt time.Time) CaseStepRuntime {
	step := strings.TrimSpace(report.Step)
	if step == "" {
		return previous
	}

	status := normalizeCaseStepStatus(report.Status)
	if status == "" {
		if previous.Step == step && previous.Status != "" {
			status = previous.Status
		} else {
			status = CaseStepStatusRunning
		}
	}

	current := CaseStepRuntime{
		Step:       step,
		StepName:   strings.TrimSpace(report.StepName),
		Status:     status,
		StartedAt:  report.StartedAt,
		UpdatedAt:  report.UpdatedAt,
		FinishedAt: report.FinishedAt,
		Attempt:    report.Attempt,
		ErrorClass: strings.TrimSpace(report.ErrorClass),
	}
	if current.StepName == "" {
		current.StepName = current.Step
	}
	if previous.Step == current.Step {
		if current.StartedAt.IsZero() {
			current.StartedAt = previous.StartedAt
		}
		if current.Attempt == 0 {
			current.Attempt = previous.Attempt
		}
		if current.ErrorClass == "" {
			current.ErrorClass = previous.ErrorClass
		}
	}
	if current.StartedAt.IsZero() {
		current.StartedAt = observedAt
	}
	if current.UpdatedAt.IsZero() {
		current.UpdatedAt = observedAt
	}
	if current.Attempt == 0 {
		current.Attempt = 1
	}
	if isTerminalCaseStepStatus(current.Status) && current.FinishedAt.IsZero() {
		current.FinishedAt = current.UpdatedAt
	}
	return current
}

func normalizeCaseStepStatus(status CaseStepStatus) CaseStepStatus {
	switch status {
	case "":
		return ""
	case CaseStepStatusUnknown, CaseStepStatusRunning, CaseStepStatusSucceeded, CaseStepStatusFailed, CaseStepStatusCanceled, CaseStepStatusSkipped:
		return status
	default:
		return CaseStepStatusUnknown
	}
}

func reconcileTaskStepEvents(previous ActiveTask, current ActiveTask, occurredAt time.Time, workerID WorkerID) []DomainEvent {
	if !hasCaseStep(current.CurrentStep) {
		return nil
	}
	if !hasCaseStep(previous.CurrentStep) {
		return []DomainEvent{taskStepChangedEvent(previous, current, occurredAt, workerID, "task_step_reported")}
	}
	if previous.CurrentStep.Step != current.CurrentStep.Step {
		return []DomainEvent{
			taskStepFinishedEvent(previous, previous.CurrentStep, occurredAt, workerID, "task_step_transitioned"),
			taskStepChangedEvent(previous, current, occurredAt, workerID, "task_step_transitioned"),
		}
	}
	if isTerminalCaseStepStatus(current.CurrentStep.Status) && previous.CurrentStep.Status != current.CurrentStep.Status {
		return []DomainEvent{taskStepFinishedEvent(current, current.CurrentStep, occurredAt, workerID, "task_step_finished")}
	}
	if isTerminalCaseStepStatus(current.CurrentStep.Status) && previous.CurrentStep.FinishedAt.IsZero() && !current.CurrentStep.FinishedAt.IsZero() {
		return []DomainEvent{taskStepFinishedEvent(current, current.CurrentStep, occurredAt, workerID, "task_step_finished")}
	}
	return nil
}

func taskStepChangedEvent(previous ActiveTask, current ActiveTask, occurredAt time.Time, workerID WorkerID, reason string) DomainEvent {
	return DomainEvent{
		Type:       EventTaskStepChanged,
		OccurredAt: occurredAt,
		WorkerID:   workerID,
		TaskID:     current.TaskID,
		Reason:     reason,
		Attributes: map[string]string{
			"exec_plan_id":       string(current.ExecPlanID),
			"from_step":          previous.CurrentStep.Step,
			"to_step":            current.CurrentStep.Step,
			"from_step_name":     previous.CurrentStep.StepName,
			"to_step_name":       current.CurrentStep.StepName,
			"from_step_status":   string(previous.CurrentStep.Status),
			"to_step_status":     string(current.CurrentStep.Status),
			"to_step_attempt":    strconv.Itoa(current.CurrentStep.Attempt),
			"to_step_started_at": current.CurrentStep.StartedAt.Format(time.RFC3339Nano),
		},
	}
}

func taskStepFinishedEvent(task ActiveTask, step CaseStepRuntime, occurredAt time.Time, workerID WorkerID, reason string) DomainEvent {
	finishedAt := nonZeroTime(step.FinishedAt, occurredAt)
	return DomainEvent{
		Type:       EventTaskStepFinished,
		OccurredAt: occurredAt,
		WorkerID:   workerID,
		TaskID:     task.TaskID,
		Reason:     reason,
		Attributes: map[string]string{
			"exec_plan_id":  string(task.ExecPlanID),
			"step":          step.Step,
			"step_name":     step.StepName,
			"step_status":   string(step.Status),
			"result":        string(step.Status),
			"error_class":   step.ErrorClass,
			"step_attempt":  strconv.Itoa(step.Attempt),
			"step_started":  step.StartedAt.Format(time.RFC3339Nano),
			"step_finished": finishedAt.Format(time.RFC3339Nano),
		},
	}
}

func hasCaseStep(step CaseStepRuntime) bool {
	return step.Step != ""
}

func isTerminalCaseStepStatus(status CaseStepStatus) bool {
	switch status {
	case CaseStepStatusSucceeded, CaseStepStatusFailed, CaseStepStatusCanceled, CaseStepStatusSkipped:
		return true
	default:
		return false
	}
}

func mergeIdentity(previous WorkerIdentity, next WorkerIdentity) WorkerIdentity {
	merged := previous
	if next.WorkerID != "" {
		merged.WorkerID = next.WorkerID
	}
	if next.Namespace != "" {
		merged.Namespace = next.Namespace
	}
	if next.PodName != "" {
		merged.PodName = next.PodName
	}
	if next.PodUID != "" {
		merged.PodUID = next.PodUID
	}
	if next.NodeName != "" {
		merged.NodeName = next.NodeName
	}
	if next.ContainerName != "" {
		merged.ContainerName = next.ContainerName
	}
	if next.Image != "" {
		merged.Image = next.Image
	}
	if next.Version != "" {
		merged.Version = next.Version
	}
	return merged
}

func hasStuckTask(tasks []ActiveTask, now time.Time, threshold time.Duration) bool {
	if threshold <= 0 {
		return false
	}
	for _, task := range tasks {
		if task.UpdatedAt.IsZero() {
			continue
		}
		if now.Sub(task.UpdatedAt) > threshold {
			return true
		}
	}
	return false
}

func latestTime(times ...time.Time) time.Time {
	var latest time.Time
	for _, candidate := range times {
		if candidate.After(latest) {
			latest = candidate
		}
	}
	return latest
}

func cloneMetadata(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneActiveTask(input ActiveTask) ActiveTask {
	input.Metadata = cloneMetadata(input.Metadata)
	return input
}

func int32String(value int32) string {
	return strconv.FormatInt(int64(value), 10)
}
