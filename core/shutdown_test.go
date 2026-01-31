package core

import (
	"context"
	"testing"
)

func TestShutdownHooksRunOnce(t *testing.T) {
	app := New()

	var calls []string
	h1 := func(ctx context.Context) error {
		calls = append(calls, "h1")
		return nil
	}
	h2 := func(ctx context.Context) error {
		calls = append(calls, "h2")
		return nil
	}

	if err := app.OnShutdown(h1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := app.OnShutdown(h2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	app.runShutdownHooks(context.Background())
	app.runShutdownHooks(context.Background())

	if len(calls) != 2 {
		t.Fatalf("expected 2 hook calls, got %d", len(calls))
	}
	if calls[0] != "h2" || calls[1] != "h1" {
		t.Fatalf("expected reverse order execution, got %v", calls)
	}
}
