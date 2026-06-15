package commands

import "testing"

func TestModuleFromFilePath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// stable roots — single-level
		{"middleware/timeout/timeout.go", "middleware"},
		{"core/app.go", "core"},
		{"router/router.go", "router"},
		// x/* — two-level (existing behaviour)
		{"x/tenant/resolver.go", "x/tenant"},
		{"x/observability/tracer/tracer.go", "x/observability"},
		// reference/* — two-level (new behaviour)
		{"reference/standard-service/internal/app/app_test.go", "reference/standard-service"},
		{"reference/with-ai/main.go", "reference/with-ai"},
		// use-cases/* — two-level (new behaviour)
		{"use-cases/workerfleet/main.go", "use-cases/workerfleet"},
		{"use-cases/mini-saas-api/cmd/api/main.go", "use-cases/mini-saas-api"},
		// other top-level paths
		{"specs/gate-profiles.yaml", "specs"},
		{"docs/reference/style-guide.md", "docs"},
		{"cmd/plumego/commands/agents_validate.go", "cmd"},
		// empty / root
		{"", ""},
	}
	for _, tt := range tests {
		got := moduleFromFilePath(tt.path)
		if got != tt.want {
			t.Errorf("moduleFromFilePath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
