package core

import (
	"strings"
	"testing"
)

func mustRegisterRoute(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("register route: %v", err)
	}
}

func newTestApp() *App {
	return New(DefaultConfig(), AppDependencies{})
}

func assertCoreError(t *testing.T, err error, operation string, fragments ...string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	prefix := "core " + operation
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("error prefix = %q, want prefix %q", got, prefix)
	}
	for _, fragment := range fragments {
		if !strings.Contains(got, fragment) {
			t.Fatalf("error = %q, want fragment %q", got, fragment)
		}
	}
}
