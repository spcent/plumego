package audit

import (
	"context"
	"fmt"
	"testing"
)

func TestRecordAndListNewestFirst(t *testing.T) {
	rec := NewRecorder(100)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := rec.Record(ctx, Entry{TenantID: "t", ActorID: "u", Action: fmt.Sprintf("act-%d", i)}); err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	got, err := rec.List(ctx, "t", 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Action != "act-2" || got[2].Action != "act-0" {
		t.Fatalf("not newest first: %s … %s", got[0].Action, got[2].Action)
	}
	for _, e := range got {
		if e.ID == "" || e.At.IsZero() {
			t.Fatalf("entry missing ID or timestamp: %+v", e)
		}
	}
}

func TestListLimit(t *testing.T) {
	rec := NewRecorder(100)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_ = rec.Record(ctx, Entry{TenantID: "t", Action: fmt.Sprintf("a%d", i)})
	}
	got, _ := rec.List(ctx, "t", 2)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Action != "a4" {
		t.Fatalf("first = %s, want a4", got[0].Action)
	}
}

func TestRingDropsOldest(t *testing.T) {
	rec := NewRecorder(2)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		_ = rec.Record(ctx, Entry{TenantID: "t", Action: fmt.Sprintf("a%d", i)})
	}
	got, _ := rec.List(ctx, "t", 0)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (ring cap)", len(got))
	}
	if got[0].Action != "a3" || got[1].Action != "a2" {
		t.Fatalf("ring kept wrong entries: %s, %s", got[0].Action, got[1].Action)
	}
}

func TestTenantIsolation(t *testing.T) {
	rec := NewRecorder(10)
	ctx := context.Background()
	_ = rec.Record(ctx, Entry{TenantID: "a", Action: "secret"})
	got, _ := rec.List(ctx, "b", 0)
	if len(got) != 0 {
		t.Fatalf("tenant b sees %d entries from tenant a", len(got))
	}
}
