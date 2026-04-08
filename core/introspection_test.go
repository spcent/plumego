package core

import "testing"

func TestPreparationStateDefaultsToMutable(t *testing.T) {
	app := New(DefaultConfig(), AppDependencies{})
	if app.preparationState != PreparationStateMutable {
		t.Fatalf("preparationState = %q, want %q", app.preparationState, PreparationStateMutable)
	}
}

func TestPreparationStateCanBeObservedInternally(t *testing.T) {
	app := New(DefaultConfig(), AppDependencies{})
	app.preparationState = PreparationStateServerPrepared

	if app.preparationState != PreparationStateServerPrepared {
		t.Fatalf("preparationState = %q, want %q", app.preparationState, PreparationStateServerPrepared)
	}
}
