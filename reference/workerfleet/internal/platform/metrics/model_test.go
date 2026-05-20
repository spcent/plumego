package metrics

import (
	"reflect"
	"testing"
)

func TestCatalogRejectsForbiddenLabels(t *testing.T) {
	for _, spec := range Catalog() {
		if err := ValidateSpec(spec); err != nil {
			t.Fatalf("ValidateSpec(%s) = %v", spec.Name, err)
		}
		labels := make(map[string]string, len(spec.Labels))
		for _, label := range spec.Labels {
			labels[label] = "value"
		}
		if err := ValidateLabels(labels); err != nil {
			t.Fatalf("%s uses forbidden label: %v", spec.Name, err)
		}
	}
}

func TestValidateLabelsRejectsHighCardinalityLabels(t *testing.T) {
	for _, label := range []string{"task_id", "case_id", "worker_id", "pod_name", "pod_uid", "raw_error", "error_message"} {
		if err := ValidateLabels(map[string]string{label: "value"}); err == nil {
			t.Fatalf("ValidateLabels(%q) succeeded, want error", label)
		}
	}
}

func TestCatalogNamesAreStableAndNamespaced(t *testing.T) {
	seen := map[string]struct{}{}
	for _, spec := range Catalog() {
		if err := ValidateSpec(spec); err != nil {
			t.Fatalf("ValidateSpec(%s) = %v", spec.Name, err)
		}
		if _, ok := seen[spec.Name]; ok {
			t.Fatalf("duplicate metric name %q", spec.Name)
		}
		seen[spec.Name] = struct{}{}
	}
}

func TestStableMetricCatalogFreezesLabelContract(t *testing.T) {
	want := map[string][]string{
		MetricWorkersTotal:                 {LabelNamespace, LabelNode, LabelStatus},
		MetricPodsTotal:                    {LabelNamespace, LabelNode, LabelPhase},
		MetricActiveCases:                  {LabelNamespace, LabelNode, LabelPhase, LabelTaskType},
		MetricWorkerActiveCases:            {LabelNamespace, LabelNode, LabelPod},
		MetricWorkerAcceptingTasks:         {LabelNamespace, LabelNode},
		MetricWorkerHeartbeatAgeSeconds:    {LabelNamespace, LabelNode, LabelPod, LabelStatus},
		MetricNodeActiveCases:              {LabelNode, LabelPhase, LabelTaskType},
		MetricCaseStartedTotal:             {LabelNamespace, LabelNode, LabelTaskType},
		MetricCaseFinishedTotal:            {LabelNamespace, LabelNode, LabelStatus, LabelTaskType},
		MetricCasePhaseTransitionsTotal:    {LabelFromPhase, LabelNamespace, LabelNode, LabelTaskType, LabelToPhase},
		MetricWorkerStatusTransitionsTotal: {LabelFromStatus, LabelNamespace, LabelNode, LabelToStatus},
		MetricAlertsTotal:                  {LabelAlertType, LabelSeverity, LabelStatus},
		MetricAlertsFiring:                 {LabelAlertType, LabelSeverity},
		MetricCasePhaseDurationSeconds:     {LabelNamespace, LabelNode, LabelPhase, LabelTaskType},
		MetricCaseTotalDurationSeconds:     {LabelNamespace, LabelNode, LabelStatus, LabelTaskType},
		MetricWorkerReportApplySeconds:     {LabelOperation},
		MetricKubeInventorySyncSeconds:     {LabelOperation, LabelResult},
		MetricRuntimeErrorsTotal:           {LabelErrorClass, LabelOperation},
	}

	gotSpecs := StableCatalog()
	if len(gotSpecs) != len(want) {
		t.Fatalf("StableCatalog() count = %d, want %d", len(gotSpecs), len(want))
	}

	for _, spec := range gotSpecs {
		wantLabels, ok := want[spec.Name]
		if !ok {
			t.Fatalf("unexpected stable metric %q", spec.Name)
		}
		if spec.Stability != MetricStabilityStable {
			t.Fatalf("%s stability = %q, want %q", spec.Name, spec.Stability, MetricStabilityStable)
		}
		if !reflect.DeepEqual(LabelNames(spec), wantLabels) {
			t.Fatalf("%s labels = %v, want %v", spec.Name, LabelNames(spec), wantLabels)
		}
	}
}

func TestExperimentalMetricsAreExplicitlyFlagged(t *testing.T) {
	if len(ExperimentalCatalog()) == 0 {
		t.Fatalf("ExperimentalCatalog() is empty, want explicitly flagged experimental metrics")
	}

	for _, spec := range Catalog() {
		labels := LabelNames(spec)
		hasExecPlanID := containsLabel(labels, LabelExecPlanID)
		hasStep := containsLabel(labels, LabelStep)
		if (hasExecPlanID || hasStep) && spec.Stability != MetricStabilityExperimental {
			t.Fatalf("%s uses experimental labels but stability = %q", spec.Name, spec.Stability)
		}
		if spec.Stability == MetricStabilityExperimental && containsForbiddenLabel(spec.Labels) {
			t.Fatalf("%s experimental labels include forbidden label", spec.Name)
		}
	}
}

func containsLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}

func containsForbiddenLabel(labels []string) bool {
	for _, label := range labels {
		if err := validateLabelName(label); err != nil {
			return true
		}
	}
	return false
}
