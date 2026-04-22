package metrics

import (
	"fmt"
	"sort"
)

const Namespace = "workerfleet"

const (
	MetricWorkersTotal                 = "workerfleet_workers"
	MetricPodsTotal                    = "workerfleet_pods"
	MetricActiveCases                  = "workerfleet_active_cases"
	MetricWorkerActiveCases            = "workerfleet_worker_active_cases"
	MetricWorkerAcceptingTasks         = "workerfleet_worker_accepting_tasks"
	MetricWorkerHeartbeatAgeSeconds    = "workerfleet_worker_heartbeat_age_seconds"
	MetricNodeActiveCases              = "workerfleet_node_active_cases"
	MetricCaseStartedTotal             = "workerfleet_case_started_total"
	MetricCaseFinishedTotal            = "workerfleet_case_finished_total"
	MetricCaseCompletedTotal           = "workerfleet_case_completed_total"
	MetricCaseFailedTotal              = "workerfleet_case_failed_total"
	MetricCasePhaseTransitionsTotal    = "workerfleet_case_phase_transitions_total"
	MetricWorkerStatusTransitionsTotal = "workerfleet_worker_status_transitions_total"
	MetricAlertsTotal                  = "workerfleet_alerts_total"
	MetricAlertsFiring                 = "workerfleet_alerts_firing"
	MetricCasePhaseDurationSeconds     = "workerfleet_case_phase_duration_seconds"
	MetricCaseTotalDurationSeconds     = "workerfleet_case_total_duration_seconds"
	MetricCaseDurationSeconds          = "workerfleet_case_duration_seconds"
	MetricCaseStepCompletedTotal       = "workerfleet_case_step_completed_total"
	MetricCaseStepDurationSeconds      = "workerfleet_case_step_duration_seconds"
	MetricCaseStepStuckCases           = "workerfleet_case_step_stuck_cases"
	MetricCaseStepOldestActiveAge      = "workerfleet_case_step_oldest_active_age_seconds"
	MetricWorkerReportApplySeconds     = "workerfleet_worker_report_apply_duration_seconds"
	MetricKubeInventorySyncSeconds     = "workerfleet_kube_inventory_sync_duration_seconds"
)

const (
	LabelNamespace  = "namespace"
	LabelNode       = "node"
	LabelPod        = "pod"
	LabelStatus     = "status"
	LabelPhase      = "phase"
	LabelTaskType   = "task_type"
	LabelExecPlanID = "exec_plan_id"
	LabelStep       = "step"
	LabelAlertType  = "alert_type"
	LabelSeverity   = "severity"
	LabelFromPhase  = "from_phase"
	LabelToPhase    = "to_phase"
	LabelFromStatus = "from_status"
	LabelToStatus   = "to_status"
	LabelOperation  = "operation"
	LabelResult     = "result"
	LabelErrorClass = "error_class"
)

var forbiddenLabels = map[string]struct{}{
	"task_id":       {},
	"case_id":       {},
	"worker_id":     {},
	"pod_name":      {},
	"pod_uid":       {},
	"raw_error":     {},
	"error_message": {},
}

type MetricKind string

const (
	MetricKindGauge     MetricKind = "gauge"
	MetricKindCounter   MetricKind = "counter"
	MetricKindHistogram MetricKind = "histogram"
)

type MetricSpec struct {
	Name        string
	Kind        MetricKind
	Description string
	Labels      []string
}

func Catalog() []MetricSpec {
	return []MetricSpec{
		{Name: MetricWorkersTotal, Kind: MetricKindGauge, Description: "Current workers by status, namespace, and node.", Labels: []string{LabelStatus, LabelNamespace, LabelNode}},
		{Name: MetricPodsTotal, Kind: MetricKindGauge, Description: "Current pods by phase, namespace, and node.", Labels: []string{LabelPhase, LabelNamespace, LabelNode}},
		{Name: MetricActiveCases, Kind: MetricKindGauge, Description: "Current active cases by namespace, node, task type, and phase.", Labels: []string{LabelNamespace, LabelNode, LabelTaskType, LabelPhase}},
		{Name: MetricWorkerActiveCases, Kind: MetricKindGauge, Description: "Current active cases by namespace, node, and pod.", Labels: []string{LabelNamespace, LabelNode, LabelPod}},
		{Name: MetricWorkerAcceptingTasks, Kind: MetricKindGauge, Description: "Current workers accepting tasks by namespace and node.", Labels: []string{LabelNamespace, LabelNode}},
		{Name: MetricWorkerHeartbeatAgeSeconds, Kind: MetricKindGauge, Description: "Worker heartbeat age in seconds by namespace, node, pod, and worker status.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelStatus}},
		{Name: MetricNodeActiveCases, Kind: MetricKindGauge, Description: "Current active case count by node, task type, and phase.", Labels: []string{LabelNode, LabelTaskType, LabelPhase}},
		{Name: MetricCaseStartedTotal, Kind: MetricKindCounter, Description: "Total cases started.", Labels: []string{LabelNamespace, LabelNode, LabelTaskType}},
		{Name: MetricCaseFinishedTotal, Kind: MetricKindCounter, Description: "Total cases finished by status.", Labels: []string{LabelNamespace, LabelNode, LabelTaskType, LabelStatus}},
		{Name: MetricCaseCompletedTotal, Kind: MetricKindCounter, Description: "Total cases completed by pod, exec plan, task type, and result.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelResult}},
		{Name: MetricCaseFailedTotal, Kind: MetricKindCounter, Description: "Total failed cases by pod, exec plan, task type, and error class.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelErrorClass}},
		{Name: MetricCasePhaseTransitionsTotal, Kind: MetricKindCounter, Description: "Total case phase transitions.", Labels: []string{LabelNamespace, LabelNode, LabelTaskType, LabelFromPhase, LabelToPhase}},
		{Name: MetricWorkerStatusTransitionsTotal, Kind: MetricKindCounter, Description: "Total worker status transitions.", Labels: []string{LabelNamespace, LabelNode, LabelFromStatus, LabelToStatus}},
		{Name: MetricAlertsTotal, Kind: MetricKindCounter, Description: "Total alert records emitted.", Labels: []string{LabelAlertType, LabelSeverity, LabelStatus}},
		{Name: MetricAlertsFiring, Kind: MetricKindGauge, Description: "Current firing alerts by alert type and severity.", Labels: []string{LabelAlertType, LabelSeverity}},
		{Name: MetricCasePhaseDurationSeconds, Kind: MetricKindHistogram, Description: "Case phase duration in seconds.", Labels: []string{LabelNamespace, LabelNode, LabelTaskType, LabelPhase}},
		{Name: MetricCaseTotalDurationSeconds, Kind: MetricKindHistogram, Description: "Case total duration in seconds.", Labels: []string{LabelNamespace, LabelNode, LabelTaskType, LabelStatus}},
		{Name: MetricCaseDurationSeconds, Kind: MetricKindHistogram, Description: "Case total duration in seconds by pod, exec plan, task type, and result.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelResult}},
		{Name: MetricCaseStepCompletedTotal, Kind: MetricKindCounter, Description: "Total case steps completed by pod, exec plan, task type, step, and result.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelStep, LabelResult}},
		{Name: MetricCaseStepDurationSeconds, Kind: MetricKindHistogram, Description: "Case step duration in seconds by pod, exec plan, task type, step, and result.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelStep, LabelResult}},
		{Name: MetricCaseStepStuckCases, Kind: MetricKindGauge, Description: "Current stuck active cases by pod, exec plan, task type, step, and severity.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelStep, LabelSeverity}},
		{Name: MetricCaseStepOldestActiveAge, Kind: MetricKindGauge, Description: "Oldest active case step age in seconds by pod, exec plan, task type, and step.", Labels: []string{LabelNamespace, LabelNode, LabelPod, LabelExecPlanID, LabelTaskType, LabelStep}},
		{Name: MetricWorkerReportApplySeconds, Kind: MetricKindHistogram, Description: "Worker report apply duration in seconds.", Labels: []string{LabelOperation}},
		{Name: MetricKubeInventorySyncSeconds, Kind: MetricKindHistogram, Description: "Kubernetes inventory sync duration in seconds.", Labels: []string{LabelOperation, LabelResult}},
	}
}

func ValidateLabels(labels map[string]string) error {
	for label := range labels {
		if _, forbidden := forbiddenLabels[label]; forbidden {
			return fmt.Errorf("label %q is forbidden for default workerfleet Prometheus series", label)
		}
	}
	return nil
}

func LabelNames(spec MetricSpec) []string {
	out := make([]string, len(spec.Labels))
	copy(out, spec.Labels)
	sort.Strings(out)
	return out
}
