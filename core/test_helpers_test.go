package core

import "testing"

func mustRegisterRoute(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("register route: %v", err)
	}
}
