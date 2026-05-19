package migrate

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewNilRunnerReturnsError(t *testing.T) {
	_, err := New(nil)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestMigratorDelegatesOperations(t *testing.T) {
	appliedAt := time.Now()
	runner := &fakeRunner{
		status: []MigrationStatus{
			{Version: 1, Name: "create_widgets", State: "applied", AppliedAt: &appliedAt},
		},
	}
	migrator, err := New(runner)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := migrator.Up(t.Context()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := migrator.Down(t.Context()); err != nil {
		t.Fatalf("Down() error = %v", err)
	}
	status, err := migrator.Status(t.Context())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if len(status) != 1 || status[0].Name != "create_widgets" {
		t.Fatalf("status = %#v", status)
	}
	if err := migrator.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if !runner.up || !runner.down || !runner.closed {
		t.Fatalf("runner calls: up=%v down=%v closed=%v", runner.up, runner.down, runner.closed)
	}
}

func TestMigratorWrapsRunnerError(t *testing.T) {
	want := errors.New("boom")
	migrator, err := New(&fakeRunner{upErr: want})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	err = migrator.Up(t.Context())
	if !errors.Is(err, ErrMigrationFailed) || !errors.Is(err, want) {
		t.Fatalf("Up() error = %v, want ErrMigrationFailed wrapping boom", err)
	}
}

type fakeRunner struct {
	up      bool
	down    bool
	closed  bool
	upErr   error
	status  []MigrationStatus
	statusE error
}

func (r *fakeRunner) Up(context.Context) error {
	r.up = true
	return r.upErr
}

func (r *fakeRunner) Down(context.Context) error {
	r.down = true
	return nil
}

func (r *fakeRunner) Status(context.Context) ([]MigrationStatus, error) {
	return r.status, r.statusE
}

func (r *fakeRunner) Close() error {
	r.closed = true
	return nil
}
