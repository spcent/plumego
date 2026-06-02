package handler

import (
	"context"
	"testing"
)

func TestOperationRegistry_RegisterCancelUnregister(t *testing.T) {
	reg := NewOperationRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	id := reg.Register(OperationInfo{
		Driver:   "redis",
		Kind:     "command",
		ConnID:   "conn1",
		Resource: "db0",
		Summary:  "PING",
	}, cancel)
	if id == "" {
		t.Fatal("operation id is empty")
	}
	if reg.Count() != 1 {
		t.Fatalf("count=%d, want 1", reg.Count())
	}
	if !reg.Cancel(id) {
		t.Fatal("cancel returned false")
	}
	select {
	case <-ctx.Done():
	default:
		t.Fatal("context was not cancelled")
	}
	reg.Unregister(id)
	if reg.Count() != 0 {
		t.Fatalf("count=%d, want 0", reg.Count())
	}
}

func TestOperationRegistry_ListActive(t *testing.T) {
	reg := NewOperationRegistry()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	reg.Register(OperationInfo{OperationID: "op1", Driver: "mongodb", Kind: "query"}, cancel)
	active := reg.ListActive()
	if len(active) != 1 {
		t.Fatalf("active len=%d, want 1", len(active))
	}
	if active[0].OperationID != "op1" {
		t.Fatalf("operation id=%q, want op1", active[0].OperationID)
	}
	if reg.Cancel("missing") {
		t.Fatal("cancel missing returned true")
	}
}
