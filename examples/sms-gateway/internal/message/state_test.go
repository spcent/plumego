package message

import (
	"context"
	"testing"
)

type hookCounter struct {
	before  int
	after   int
	invalid int
}

func (h *hookCounter) Before(ctx context.Context, from Status, to Status, msg Message) error {
	h.before++
	return nil
}

func (h *hookCounter) After(ctx context.Context, from Status, to Status, msg Message) error {
	h.after++
	return nil
}

func (h *hookCounter) OnInvalid(ctx context.Context, from Status, to Status, msg Message, err error) {
	h.invalid++
}

func TestApplyTransitionAllowed(t *testing.T) {
	msg := Message{Status: StatusQueued, Version: 1}
	h := &hookCounter{}

	if err := ApplyTransition(context.Background(), &msg, StatusSending, h); err != nil {
		t.Fatalf("transition: %v", err)
	}
	if msg.Status != StatusSending {
		t.Fatalf("expected sending, got %s", msg.Status)
	}
	if msg.Version != 2 {
		t.Fatalf("expected version 2, got %d", msg.Version)
	}
	if h.before != 1 || h.after != 1 {
		t.Fatalf("expected hooks called, before=%d after=%d", h.before, h.after)
	}
}

func TestApplyTransitionInvalid(t *testing.T) {
	msg := Message{Status: StatusQueued}
	h := &hookCounter{}

	if err := ApplyTransition(context.Background(), &msg, StatusDelivered, h); err == nil {
		t.Fatalf("expected error")
	}
	if h.invalid != 1 {
		t.Fatalf("expected invalid hook called")
	}
}
