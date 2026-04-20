package metrics

import "testing"

func TestCatalogRejectsForbiddenLabels(t *testing.T) {
	for _, spec := range Catalog() {
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
	for _, label := range []string{"task_id", "case_id", "worker_id", "pod_name"} {
		if err := ValidateLabels(map[string]string{label: "value"}); err == nil {
			t.Fatalf("ValidateLabels(%q) succeeded, want error", label)
		}
	}
}

func TestCatalogNamesAreStableAndNamespaced(t *testing.T) {
	seen := map[string]struct{}{}
	for _, spec := range Catalog() {
		if spec.Name == "" {
			t.Fatalf("empty metric name")
		}
		if len(spec.Name) <= len(Namespace) || spec.Name[:len(Namespace)] != Namespace {
			t.Fatalf("metric %q does not use namespace %q", spec.Name, Namespace)
		}
		if _, ok := seen[spec.Name]; ok {
			t.Fatalf("duplicate metric name %q", spec.Name)
		}
		seen[spec.Name] = struct{}{}
	}
}
