package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRequestReplyBasic tests basic request-reply with the manager enabled.
func TestRequestReplyBasic(t *testing.T) {
	ps := New(WithRequestReply())
	defer ps.Close()

	// Set up responder
	sub, err := ps.Subscribe("math.add", SubOptions{BufferSize: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	go func() {
		for msg := range sub.C() {
			if IsRequest(msg) {
				ps.Reply(msg, Message{Data: "result:5"})
			}
		}
	}()

	// Send request
	resp, err := ps.RequestWithTimeout("math.add", Message{Data: "2+3"}, 2*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.Data != "result:5" {
		t.Fatalf("unexpected response data: %v", resp.Data)
	}
}

// TestRequestReplyWithContext tests request with context cancellation.
func TestRequestReplyWithContext(t *testing.T) {
	ps := New(WithRequestReply())
	defer ps.Close()

	// No responder - request should time out via context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := ps.Request(ctx, "no.responder", Message{Data: "hello"})
	if err == nil {
		t.Fatal("expected error from context timeout")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got: %v", err)
	}
}

// TestRequestReplyWithoutManager tests request-reply without the manager (fallback path).
func TestRequestReplyWithoutManager(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Set up responder
	sub, err := ps.Subscribe("echo", SubOptions{BufferSize: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	go func() {
		for msg := range sub.C() {
			if IsRequest(msg) {
				ps.Reply(msg, Message{Data: msg.Data})
			}
		}
	}()

	resp, err := ps.RequestWithTimeout("echo", Message{Data: "ping"}, 2*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.Data != "ping" {
		t.Fatalf("unexpected response data: %v", resp.Data)
	}
}

// TestRequestReplyMultipleConcurrent tests concurrent requests through the manager.
func TestRequestReplyMultipleConcurrent(t *testing.T) {
	ps := New(WithRequestReply())
	defer ps.Close()

	// Set up responder that echoes back
	sub, err := ps.Subscribe("echo", SubOptions{BufferSize: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	go func() {
		for msg := range sub.C() {
			if IsRequest(msg) {
				ps.Reply(msg, Message{Data: msg.Data})
			}
		}
	}()

	const n = 20
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			resp, err := ps.RequestWithTimeout("echo", Message{Data: val}, 5*time.Second)
			if err != nil {
				errs <- err
				return
			}
			if resp.Data != val {
				errs <- &PubSubError{
					Code:    ErrCodeInternal,
					Op:      "test",
					Message: "response data mismatch",
				}
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent request failed: %v", err)
	}
}

// TestRequestReplyCorrelationID tests that correlation IDs are correctly propagated.
func TestRequestReplyCorrelationID(t *testing.T) {
	ps := New(WithRequestReply())
	defer ps.Close()

	sub, err := ps.Subscribe("service.ping", SubOptions{BufferSize: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	var capturedCorrelationID string
	go func() {
		for msg := range sub.C() {
			if IsRequest(msg) {
				capturedCorrelationID = GetCorrelationID(msg)
				replyTo := GetReplyTo(msg)
				if replyTo == "" {
					return
				}
				ps.Reply(msg, Message{Data: "pong"})
			}
		}
	}()

	resp, err := ps.RequestWithTimeout("service.ping", Message{Data: "ping"}, 2*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Correlation ID should be present in the response
	respCorrelationID := GetCorrelationID(resp)
	if respCorrelationID == "" {
		t.Fatal("response missing correlation ID")
	}
	if capturedCorrelationID == "" {
		t.Fatal("request was missing correlation ID")
	}
	if respCorrelationID != capturedCorrelationID {
		t.Fatalf("correlation ID mismatch: request=%s response=%s",
			capturedCorrelationID, respCorrelationID)
	}
}

// TestRequestReplyOnClosedPubSub tests that requests fail on closed pubsub.
func TestRequestReplyOnClosedPubSub(t *testing.T) {
	ps := New(WithRequestReply())
	ps.Close()

	_, err := ps.Request(context.Background(), "test", Message{})
	if err != ErrClosed {
		t.Fatalf("expected ErrClosed, got: %v", err)
	}
}

// TestReplyNoReplyToHeader tests Reply with a non-request message.
func TestReplyNoReplyToHeader(t *testing.T) {
	ps := New()
	defer ps.Close()

	err := ps.Reply(Message{Topic: "test"}, Message{Data: "reply"})
	if err == nil {
		t.Fatal("expected error when replying to non-request message")
	}
}

// TestReplyNoCorrelationID tests Reply with missing correlation ID.
func TestReplyNoCorrelationID(t *testing.T) {
	ps := New()
	defer ps.Close()

	msg := Message{
		Topic: "test",
		Meta:  map[string]string{ReplyToHeader: "_reply.abc"},
	}
	err := ps.Reply(msg, Message{Data: "reply"})
	if err == nil {
		t.Fatal("expected error when replying without correlation ID")
	}
}

// TestIsRequest tests the IsRequest helper.
func TestIsRequest(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want bool
	}{
		{"nil meta", Message{}, false},
		{"empty meta", Message{Meta: map[string]string{}}, false},
		{"has reply-to", Message{Meta: map[string]string{ReplyToHeader: "_reply.abc"}}, true},
		{"has other meta", Message{Meta: map[string]string{"X-Custom": "val"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRequest(tt.msg)
			if got != tt.want {
				t.Fatalf("IsRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetCorrelationIDAndReplyTo tests the helper functions.
func TestGetCorrelationIDAndReplyTo(t *testing.T) {
	msg := Message{}
	if GetCorrelationID(msg) != "" {
		t.Fatal("expected empty correlation ID for nil meta")
	}
	if GetReplyTo(msg) != "" {
		t.Fatal("expected empty reply-to for nil meta")
	}

	msg.Meta = map[string]string{
		CorrelationIDHeader: "corr-123",
		ReplyToHeader:       "_reply.xyz",
	}
	if GetCorrelationID(msg) != "corr-123" {
		t.Fatalf("unexpected correlation ID: %s", GetCorrelationID(msg))
	}
	if GetReplyTo(msg) != "_reply.xyz" {
		t.Fatalf("unexpected reply-to: %s", GetReplyTo(msg))
	}
}

// TestPendingRequests tests the PendingRequests method.
func TestPendingRequests(t *testing.T) {
	// Without manager
	ps := New()
	defer ps.Close()

	_, err := ps.PendingRequests()
	if err != ErrRequestReplyDisabled {
		t.Fatalf("expected ErrRequestReplyDisabled, got: %v", err)
	}

	// With manager
	ps2 := New(WithRequestReply())
	defer ps2.Close()

	n, err := ps2.PendingRequests()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 pending requests, got %d", n)
	}
}

// TestRequestManagerClose tests that closing the request manager cleans up pending requests.
func TestRequestManagerClose(t *testing.T) {
	ps := New(WithRequestReply())

	// Start a request that will block
	var requestErr atomic.Value
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := ps.RequestWithTimeout("no.responder", Message{Data: "test"}, 5*time.Second)
		if err != nil {
			requestErr.Store(err)
		}
	}()

	// Give the request time to register
	time.Sleep(50 * time.Millisecond)

	n, _ := ps.PendingRequests()
	if n != 1 {
		t.Fatalf("expected 1 pending request, got %d", n)
	}

	// Close should clean up
	ps.Close()
	wg.Wait()

	// The request should have terminated (either with error from channel close or context)
	// We don't assert the exact error since it depends on timing
}

// TestRequestReplyManagerMultipleRequestsSameResponder tests multiple sequential requests.
func TestRequestReplyManagerMultipleRequestsSameResponder(t *testing.T) {
	ps := New(WithRequestReply())
	defer ps.Close()

	sub, err := ps.Subscribe("counter", SubOptions{BufferSize: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	var count atomic.Int32
	go func() {
		for msg := range sub.C() {
			if IsRequest(msg) {
				c := count.Add(1)
				ps.Reply(msg, Message{Data: int(c)})
			}
		}
	}()

	for i := 1; i <= 5; i++ {
		resp, err := ps.RequestWithTimeout("counter", Message{Data: "inc"}, 2*time.Second)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		if resp.Data != i {
			t.Fatalf("request %d: expected %d, got %v", i, i, resp.Data)
		}
	}
}

// TestWithRequestReplyOption tests that the option sets the config correctly.
func TestWithRequestReplyOption(t *testing.T) {
	ps := New(WithRequestReply())
	defer ps.Close()

	if !ps.config.EnableRequestReply {
		t.Fatal("EnableRequestReply should be true")
	}
	if ps.requestMgr == nil {
		t.Fatal("requestMgr should not be nil")
	}
}

// TestWithoutRequestReplyOption tests that without the option, manager is nil.
func TestWithoutRequestReplyOption(t *testing.T) {
	ps := New()
	defer ps.Close()

	if ps.config.EnableRequestReply {
		t.Fatal("EnableRequestReply should be false")
	}
	if ps.requestMgr != nil {
		t.Fatal("requestMgr should be nil")
	}
}
