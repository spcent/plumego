package metrics

import (
	"sync"
	"time"

	"workerfleet/internal/domain"
)

const unknownLabel = "unknown"

type Observer struct {
	collector *Collector
	mu        sync.Mutex
	seen      map[domain.WorkerID]struct{}
}

func NewObserver(collector *Collector) *Observer {
	return &Observer{collector: collector, seen: make(map[domain.WorkerID]struct{})}
}

func (o *Observer) ObserveWorkerSnapshot(previous domain.WorkerSnapshot, current domain.WorkerSnapshot) {
	if o == nil || o.collector == nil {
		return
	}
	firstObservation := o.markObserved(current.Identity.WorkerID)
	o.observeWorkerStatus(previous, current)
	o.observeAcceptingTasks(previous, current)
	o.observeActiveTasks(previous, current)
	if firstObservation {
		o.observeUnchangedBaseline(previous, current)
	}
}

func (o *Observer) ObserveWorkerReportApplied(operation string, duration time.Duration) {
	if o == nil || o.collector == nil {
		return
	}
	o.observeHistogram(MetricWorkerReportApplySeconds, map[string]string{
		LabelOperation: safeLabel(operation),
	}, duration.Seconds())
}

func (o *Observer) ObserveAlerts(records []domain.AlertRecord) {
	if o == nil || o.collector == nil {
		return
	}
	for _, record := range records {
		status := safeLabel(string(record.Status))
		alertType := safeLabel(string(record.AlertType))
		severity := safeLabel(record.Severity)
		o.addCounter(MetricAlertsTotal, map[string]string{
			LabelAlertType: alertType,
			LabelSeverity:  severity,
			LabelStatus:    status,
		}, 1)

		labels := map[string]string{
			LabelAlertType: alertType,
			LabelSeverity:  severity,
		}
		switch record.Status {
		case domain.AlertStatusFiring:
			o.addGauge(MetricAlertsFiring, labels, 1)
		case domain.AlertStatusResolved:
			o.addGauge(MetricAlertsFiring, labels, -1)
		}
	}
}

func (o *Observer) ObserveInventorySync(snapshots []domain.WorkerSnapshot, operation string, result string, duration time.Duration) {
	if o == nil || o.collector == nil {
		return
	}
	o.observeHistogram(MetricKubeInventorySyncSeconds, map[string]string{
		LabelOperation: safeLabel(operation),
		LabelResult:    safeLabel(result),
	}, duration.Seconds())
	if result != "success" {
		return
	}

	counts := make(map[string]GaugeSample)
	for _, snapshot := range snapshots {
		labels := podLabels(snapshot)
		key := canonicalLabels(labels)
		sample := counts[key]
		if sample.Labels == nil {
			sample.Labels = labels
		}
		sample.Value++
		counts[key] = sample
	}
	samples := make([]GaugeSample, 0, len(counts))
	for _, sample := range counts {
		samples = append(samples, sample)
	}
	_ = o.collector.SetGaugeSeries(MetricPodsTotal, samples)
}

func (o *Observer) observeWorkerStatus(previous domain.WorkerSnapshot, current domain.WorkerSnapshot) {
	if previous.Status == current.Status && canonicalLabels(workerStatusLabels(previous)) == canonicalLabels(workerStatusLabels(current)) {
		return
	}
	if previous.Status != "" {
		o.addGauge(MetricWorkersTotal, workerStatusLabels(previous), -1)
	}
	if current.Status != "" {
		o.addGauge(MetricWorkersTotal, workerStatusLabels(current), 1)
	}
	if previous.Status != "" && current.Status != "" && previous.Status != current.Status {
		o.addCounter(MetricWorkerStatusTransitionsTotal, map[string]string{
			LabelNamespace:  safeLabel(current.Identity.Namespace),
			LabelNode:       safeLabel(current.Identity.NodeName),
			LabelFromStatus: safeLabel(string(previous.Status)),
			LabelToStatus:   safeLabel(string(current.Status)),
		}, 1)
	}
}

func (o *Observer) observeAcceptingTasks(previous domain.WorkerSnapshot, current domain.WorkerSnapshot) {
	if previous.Runtime.AcceptingTasks == current.Runtime.AcceptingTasks && canonicalLabels(identityLabels(previous.Identity)) == canonicalLabels(identityLabels(current.Identity)) {
		return
	}
	if previous.Runtime.AcceptingTasks {
		o.addGauge(MetricWorkerAcceptingTasks, identityLabels(previous.Identity), -1)
	}
	if current.Runtime.AcceptingTasks {
		o.addGauge(MetricWorkerAcceptingTasks, identityLabels(current.Identity), 1)
	}
}

func (o *Observer) observeActiveTasks(previous domain.WorkerSnapshot, current domain.WorkerSnapshot) {
	previousTasks := activeTaskMap(previous.ActiveTasks)
	currentTasks := activeTaskMap(current.ActiveTasks)

	for taskID, previousTask := range previousTasks {
		currentTask, found := currentTasks[taskID]
		if !found {
			o.addActiveTaskGauge(previous.Identity, previousTask, -1)
			o.observeTaskFinished(current, previousTask)
			continue
		}
		if activeTaskSeriesChanged(previous.Identity, previousTask, current.Identity, currentTask) {
			o.addActiveTaskGauge(previous.Identity, previousTask, -1)
			o.addActiveTaskGauge(current.Identity, currentTask, 1)
		}
		if previousTask.Phase != currentTask.Phase || previousTask.PhaseName != currentTask.PhaseName {
			o.observeTaskPhaseTransition(current.Identity, previousTask, currentTask)
		}
		delete(currentTasks, taskID)
	}

	for _, currentTask := range currentTasks {
		o.addActiveTaskGauge(current.Identity, currentTask, 1)
		o.addCounter(MetricCaseStartedTotal, map[string]string{
			LabelNamespace: safeLabel(current.Identity.Namespace),
			LabelNode:      safeLabel(current.Identity.NodeName),
			LabelTaskType:  taskTypeLabel(currentTask.TaskType),
		}, 1)
	}
}

func (o *Observer) observeUnchangedBaseline(previous domain.WorkerSnapshot, current domain.WorkerSnapshot) {
	if current.Status != "" && previous.Status == current.Status && canonicalLabels(workerStatusLabels(previous)) == canonicalLabels(workerStatusLabels(current)) {
		o.addGauge(MetricWorkersTotal, workerStatusLabels(current), 1)
	}
	if current.Runtime.AcceptingTasks && previous.Runtime.AcceptingTasks == current.Runtime.AcceptingTasks && canonicalLabels(identityLabels(previous.Identity)) == canonicalLabels(identityLabels(current.Identity)) {
		o.addGauge(MetricWorkerAcceptingTasks, identityLabels(current.Identity), 1)
	}

	previousTasks := activeTaskMap(previous.ActiveTasks)
	for _, currentTask := range activeTaskMap(current.ActiveTasks) {
		previousTask, found := previousTasks[currentTask.TaskID]
		if !found {
			continue
		}
		if !activeTaskSeriesChanged(previous.Identity, previousTask, current.Identity, currentTask) {
			o.addActiveTaskGauge(current.Identity, currentTask, 1)
		}
	}
}

func (o *Observer) addActiveTaskGauge(identity domain.WorkerIdentity, task domain.ActiveTask, delta float64) {
	o.addGauge(MetricActiveCases, activeCaseLabels(identity, task), delta)
	o.addGauge(MetricNodeActiveCases, nodeActiveCaseLabels(identity, task), delta)
}

func (o *Observer) observeTaskPhaseTransition(identity domain.WorkerIdentity, previous domain.ActiveTask, current domain.ActiveTask) {
	o.addCounter(MetricCasePhaseTransitionsTotal, map[string]string{
		LabelNamespace: safeLabel(identity.Namespace),
		LabelNode:      safeLabel(identity.NodeName),
		LabelTaskType:  taskTypeLabel(current.TaskType),
		LabelFromPhase: safeLabel(string(previous.Phase)),
		LabelToPhase:   safeLabel(string(current.Phase)),
	}, 1)
	if previous.UpdatedAt.IsZero() || current.UpdatedAt.IsZero() {
		return
	}
	duration := current.UpdatedAt.Sub(previous.UpdatedAt)
	if duration <= 0 {
		return
	}
	o.observeHistogram(MetricCasePhaseDurationSeconds, map[string]string{
		LabelNamespace: safeLabel(identity.Namespace),
		LabelNode:      safeLabel(identity.NodeName),
		LabelTaskType:  taskTypeLabel(previous.TaskType),
		LabelPhase:     safeLabel(string(previous.Phase)),
	}, duration.Seconds())
}

func (o *Observer) observeTaskFinished(current domain.WorkerSnapshot, task domain.ActiveTask) {
	finishedAt := snapshotObservedAt(current)
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	status := finishedStatus(task.Phase)
	o.addCounter(MetricCaseFinishedTotal, map[string]string{
		LabelNamespace: safeLabel(current.Identity.Namespace),
		LabelNode:      safeLabel(current.Identity.NodeName),
		LabelTaskType:  taskTypeLabel(task.TaskType),
		LabelStatus:    status,
	}, 1)
	if !task.UpdatedAt.IsZero() {
		duration := finishedAt.Sub(task.UpdatedAt)
		if duration > 0 {
			o.observeHistogram(MetricCasePhaseDurationSeconds, map[string]string{
				LabelNamespace: safeLabel(current.Identity.Namespace),
				LabelNode:      safeLabel(current.Identity.NodeName),
				LabelTaskType:  taskTypeLabel(task.TaskType),
				LabelPhase:     safeLabel(string(task.Phase)),
			}, duration.Seconds())
		}
	}
	if !task.StartedAt.IsZero() {
		duration := finishedAt.Sub(task.StartedAt)
		if duration > 0 {
			o.observeHistogram(MetricCaseTotalDurationSeconds, map[string]string{
				LabelNamespace: safeLabel(current.Identity.Namespace),
				LabelNode:      safeLabel(current.Identity.NodeName),
				LabelTaskType:  taskTypeLabel(task.TaskType),
				LabelStatus:    status,
			}, duration.Seconds())
		}
	}
}

func (o *Observer) addGauge(name string, labels map[string]string, delta float64) {
	_ = o.collector.AddGauge(name, labels, delta)
}

func (o *Observer) addCounter(name string, labels map[string]string, delta float64) {
	_ = o.collector.AddCounter(name, labels, delta)
}

func (o *Observer) observeHistogram(name string, labels map[string]string, value float64) {
	_ = o.collector.ObserveHistogram(name, labels, value)
}

func (o *Observer) markObserved(workerID domain.WorkerID) bool {
	if workerID == "" {
		return false
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.seen[workerID]; ok {
		return false
	}
	o.seen[workerID] = struct{}{}
	return true
}

func activeTaskMap(tasks []domain.ActiveTask) map[domain.TaskID]domain.ActiveTask {
	out := make(map[domain.TaskID]domain.ActiveTask, len(tasks))
	for _, task := range tasks {
		if task.TaskID == "" {
			continue
		}
		out[task.TaskID] = task
	}
	return out
}

func activeTaskSeriesChanged(previousIdentity domain.WorkerIdentity, previousTask domain.ActiveTask, currentIdentity domain.WorkerIdentity, currentTask domain.ActiveTask) bool {
	return canonicalLabels(activeCaseLabels(previousIdentity, previousTask)) != canonicalLabels(activeCaseLabels(currentIdentity, currentTask)) ||
		canonicalLabels(nodeActiveCaseLabels(previousIdentity, previousTask)) != canonicalLabels(nodeActiveCaseLabels(currentIdentity, currentTask))
}

func workerStatusLabels(snapshot domain.WorkerSnapshot) map[string]string {
	labels := identityLabels(snapshot.Identity)
	labels[LabelStatus] = safeLabel(string(snapshot.Status))
	return labels
}

func podLabels(snapshot domain.WorkerSnapshot) map[string]string {
	labels := identityLabels(snapshot.Identity)
	labels[LabelPhase] = safeLabel(string(snapshot.Pod.Phase))
	return labels
}

func identityLabels(identity domain.WorkerIdentity) map[string]string {
	return map[string]string{
		LabelNamespace: safeLabel(identity.Namespace),
		LabelNode:      safeLabel(identity.NodeName),
	}
}

func activeCaseLabels(identity domain.WorkerIdentity, task domain.ActiveTask) map[string]string {
	return map[string]string{
		LabelNamespace: safeLabel(identity.Namespace),
		LabelNode:      safeLabel(identity.NodeName),
		LabelTaskType:  taskTypeLabel(task.TaskType),
		LabelPhase:     safeLabel(string(task.Phase)),
	}
}

func nodeActiveCaseLabels(identity domain.WorkerIdentity, task domain.ActiveTask) map[string]string {
	return map[string]string{
		LabelNode:     safeLabel(identity.NodeName),
		LabelTaskType: taskTypeLabel(task.TaskType),
		LabelPhase:    safeLabel(string(task.Phase)),
	}
}

func snapshotObservedAt(snapshot domain.WorkerSnapshot) time.Time {
	if snapshot.Runtime.LastHeartbeatAt.After(snapshot.Runtime.LastSeenAt) {
		return snapshot.Runtime.LastHeartbeatAt
	}
	return snapshot.Runtime.LastSeenAt
}

func finishedStatus(phase domain.TaskPhase) string {
	switch phase {
	case domain.TaskPhaseSucceeded, domain.TaskPhaseFailed, domain.TaskPhaseCanceled:
		return safeLabel(string(phase))
	default:
		return "finished"
	}
}

func taskTypeLabel(value string) string {
	if value == "" {
		return unknownLabel
	}
	return value
}

func safeLabel(value string) string {
	if value == "" {
		return unknownLabel
	}
	return value
}
