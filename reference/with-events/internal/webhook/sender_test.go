package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spcent/plumego/reference/with-events/internal/order"
)

func TestSenderSuccessFirstAttempt(t *testing.T) {
	var attempts atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content-type = %q, want application/json", got)
		}
		var event order.OrderCreated
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode event: %v", err)
		}
		if event.OrderID != "ord_123" {
			t.Errorf("order id = %q, want ord_123", event.OrderID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	sender := NewSender(target.URL, WithSleep(noSleep))
	if err := sender.Send(context.Background(), testEvent()); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
}

func TestSenderRetriesThenSucceeds(t *testing.T) {
	var attempts atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			http.Error(w, "temporary failure", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	var delays []time.Duration
	sender := NewSender(target.URL, WithSleep(func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	}))
	if err := sender.Send(context.Background(), testEvent()); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got := attempts.Load(); got != 3 {
		t.Fatalf("attempts = %d, want 3", got)
	}
	want := []time.Duration{time.Second, 2 * time.Second}
	if len(delays) != len(want) {
		t.Fatalf("delays = %v, want %v", delays, want)
	}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delays = %v, want %v", delays, want)
		}
	}
}

func TestSenderReturnsErrorAfterMaxRetries(t *testing.T) {
	var attempts atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "down", http.StatusInternalServerError)
	}))
	defer target.Close()

	sender := NewSender(target.URL, WithSleep(noSleep))
	if err := sender.Send(context.Background(), testEvent()); err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	if got := attempts.Load(); got != MaxRetries {
		t.Fatalf("attempts = %d, want %d", got, MaxRetries)
	}
}

func TestSenderEmptyTargetIsNoop(t *testing.T) {
	var calls atomic.Int64
	sender := NewSender("", WithHTTPClient(roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})))

	if err := sender.Send(context.Background(), testEvent()); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("http calls = %d, want 0", got)
	}
}

func testEvent() order.OrderCreated {
	return order.OrderCreated{
		ID:         "evt_123",
		OrderID:    "ord_123",
		CustomerID: "cus_123",
		TotalCents: 4200,
		CreatedAt:  time.Unix(1700000000, 0).UTC(),
	}
}

func noSleep(context.Context, time.Duration) error {
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}
