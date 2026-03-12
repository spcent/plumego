package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spcent/plumego/x/mq/store"
	"github.com/spcent/plumego/x/pubsub"
	"github.com/spcent/plumego/x/scheduler"
)

// --- mock providers ---

type mockSMS struct {
	mu    sync.Mutex
	calls []SMSMessage
	err   error
	seq   atomic.Int64
}

func (m *mockSMS) Name() string { return "mock-sms" }
func (m *mockSMS) Send(_ context.Context, msg SMSMessage) (*SMSResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, msg)
	if m.err != nil {
		return nil, m.err
	}
	return &SMSResult{ProviderID: fmt.Sprintf("sms-%d", m.seq.Add(1))}, nil
}

type mockEmail struct {
	mu    sync.Mutex
	calls []EmailMessage
	err   error
	seq   atomic.Int64
}

func (m *mockEmail) Name() string { return "mock-email" }
func (m *mockEmail) Send(_ context.Context, msg EmailMessage) (*EmailResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, msg)
	if m.err != nil {
		return nil, m.err
	}
	return &EmailResult{MessageID: fmt.Sprintf("email-%d", m.seq.Add(1))}, nil
}

// --- helpers ---

func newTestService(sms SMSProvider, email EmailProvider) *Service {
	return New(Config{
		TaskStore:         store.NewMemory(store.DefaultMemConfig()),
		SMS:               sms,
		Email:             email,
		WorkerConcurrency: 2,
	})
}

func waitForSent(svc *Service, want int64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if svc.totalSent.Load() >= want {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func waitForFailed(svc *Service, want int64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if svc.totalFailed.Load() >= want {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// --- tests ---

func TestValidation_SMS(t *testing.T) {
	sms := &mockSMS{}
	svc := newTestService(sms, nil)
	ctx := context.Background()

	tests := []struct {
		name string
		req  SendRequest
		err  error
	}{
		{
			name: "missing id",
			req:  SendRequest{Channel: ChannelSMS, To: "+1234567890", Body: "hi"},
			err:  ErrMissingID,
		},
		{
			name: "missing to",
			req:  SendRequest{ID: "1", Channel: ChannelSMS, Body: "hi"},
			err:  ErrMissingRecipient,
		},
		{
			name: "invalid phone",
			req:  SendRequest{ID: "2", Channel: ChannelSMS, To: "not-a-phone", Body: "hi"},
			err:  ErrInvalidPhone,
		},
		{
			name: "missing body and template",
			req:  SendRequest{ID: "3", Channel: ChannelSMS, To: "+1234567890"},
			err:  ErrMissingBody,
		},
		{
			name: "valid sms",
			req:  SendRequest{ID: "4", Channel: ChannelSMS, To: "+1234567890", Body: "hello"},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Send(ctx, tt.req)
			if tt.err == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.err)
			}
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected error %v, got %v", tt.err, err)
			}
		})
	}
}

func TestValidation_Email(t *testing.T) {
	email := &mockEmail{}
	svc := newTestService(nil, email)
	ctx := context.Background()

	tests := []struct {
		name string
		req  SendRequest
		err  error
	}{
		{
			name: "invalid email",
			req:  SendRequest{ID: "1", Channel: ChannelEmail, To: "bad", Subject: "hi", Body: "x"},
			err:  ErrInvalidEmail,
		},
		{
			name: "missing subject",
			req:  SendRequest{ID: "2", Channel: ChannelEmail, To: "a@b.com", Body: "x"},
			err:  ErrMissingSubject,
		},
		{
			name: "missing body",
			req:  SendRequest{ID: "3", Channel: ChannelEmail, To: "a@b.com", Subject: "s"},
			err:  ErrMissingBody,
		},
		{
			name: "valid email",
			req:  SendRequest{ID: "4", Channel: ChannelEmail, To: "a@b.com", Subject: "s", Body: "body"},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Send(ctx, tt.req)
			if tt.err == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.err)
			}
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected error %v, got %v", tt.err, err)
			}
		})
	}
}

func TestValidation_InvalidChannel(t *testing.T) {
	svc := newTestService(nil, nil)
	err := svc.Send(context.Background(), SendRequest{
		ID:      "1",
		Channel: "fax",
		To:      "+1234567890",
		Body:    "hi",
	})
	if !errors.Is(err, ErrInvalidChannel) {
		t.Fatalf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestSendAndWorker_SMS(t *testing.T) {
	sms := &mockSMS{}
	svc := newTestService(sms, nil)
	ctx := context.Background()

	svc.Start(ctx)
	defer svc.Stop(ctx)

	err := svc.Send(ctx, SendRequest{
		ID:      "msg-1",
		Channel: ChannelSMS,
		To:      "+1234567890",
		Body:    "Hello from test",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if !waitForSent(svc, 1, 5*time.Second) {
		t.Fatal("timed out waiting for SMS to be sent")
	}

	sms.mu.Lock()
	defer sms.mu.Unlock()
	if len(sms.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(sms.calls))
	}
	if sms.calls[0].To != "+1234567890" {
		t.Fatalf("expected to=+1234567890, got %s", sms.calls[0].To)
	}
	if sms.calls[0].Body != "Hello from test" {
		t.Fatalf("expected body='Hello from test', got %s", sms.calls[0].Body)
	}
}

func TestSendAndWorker_Email(t *testing.T) {
	email := &mockEmail{}
	svc := newTestService(nil, email)
	ctx := context.Background()

	svc.Start(ctx)
	defer svc.Stop(ctx)

	err := svc.Send(ctx, SendRequest{
		ID:      "msg-2",
		Channel: ChannelEmail,
		To:      "user@example.com",
		Subject: "Test Subject",
		Body:    "Email body",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if !waitForSent(svc, 1, 5*time.Second) {
		t.Fatal("timed out waiting for email to be sent")
	}

	email.mu.Lock()
	defer email.mu.Unlock()
	if len(email.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(email.calls))
	}
	if email.calls[0].Subject != "Test Subject" {
		t.Fatalf("expected subject='Test Subject', got %s", email.calls[0].Subject)
	}
}

func TestSendBatch(t *testing.T) {
	sms := &mockSMS{}
	email := &mockEmail{}
	svc := newTestService(sms, email)
	ctx := context.Background()

	result := svc.SendBatch(ctx, BatchRequest{
		Requests: []SendRequest{
			{ID: "b1", Channel: ChannelSMS, To: "+1234567890", Body: "sms1"},
			{ID: "b2", Channel: ChannelEmail, To: "a@b.com", Subject: "s", Body: "email1"},
			{ID: "b3", Channel: "fax", To: "x", Body: "nope"},
		},
	})
	if result.Total != 3 {
		t.Fatalf("total=%d, want 3", result.Total)
	}
	if result.Accepted != 2 {
		t.Fatalf("accepted=%d, want 2", result.Accepted)
	}
	if result.Rejected != 1 {
		t.Fatalf("rejected=%d, want 1", result.Rejected)
	}
}

func TestStats(t *testing.T) {
	svc := newTestService(&mockSMS{}, nil)
	ctx := context.Background()

	stats, err := svc.Stats(ctx)
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if stats.Queued != 0 {
		t.Fatalf("queued=%d, want 0", stats.Queued)
	}

	// Enqueue one
	svc.Send(ctx, SendRequest{
		ID: "s1", Channel: ChannelSMS, To: "+1234567890", Body: "hi",
	})

	stats, _ = svc.Stats(ctx)
	if stats.Queued != 1 {
		t.Fatalf("queued=%d, want 1", stats.Queued)
	}
}

func TestTemplate(t *testing.T) {
	sms := &mockSMS{}
	svc := newTestService(sms, nil)
	ctx := context.Background()

	svc.Templates().Register("otp", "Your code is {{code}}. Expires in {{minutes}} minutes.")

	svc.Start(ctx)
	defer svc.Stop(ctx)

	err := svc.Send(ctx, SendRequest{
		ID:       "t1",
		Channel:  ChannelSMS,
		To:       "+1234567890",
		Template: "otp",
		Params:   map[string]string{"code": "123456", "minutes": "5"},
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if !waitForSent(svc, 1, 5*time.Second) {
		t.Fatal("timed out")
	}

	sms.mu.Lock()
	defer sms.mu.Unlock()
	if len(sms.calls) == 0 {
		t.Fatal("no calls")
	}
	want := "Your code is 123456. Expires in 5 minutes."
	if sms.calls[0].Body != want {
		t.Fatalf("body=%q, want %q", sms.calls[0].Body, want)
	}
}

func TestDeduplication(t *testing.T) {
	svc := newTestService(&mockSMS{}, nil)
	ctx := context.Background()

	req := SendRequest{
		ID:        "dup-1",
		Channel:   ChannelSMS,
		To:        "+1234567890",
		Body:      "hello",
		DedupeKey: "unique-key",
	}
	err := svc.Send(ctx, req)
	if err != nil {
		t.Fatalf("first send: %v", err)
	}

	// Same ID should fail (duplicate task)
	req.ID = "dup-1" // same ID
	err = svc.Send(ctx, req)
	if err == nil {
		t.Fatal("expected duplicate error, got nil")
	}
}

func TestScheduledSend(t *testing.T) {
	svc := newTestService(&mockSMS{}, nil)
	ctx := context.Background()

	future := time.Now().Add(1 * time.Hour)
	err := svc.Send(ctx, SendRequest{
		ID:          "sched-1",
		Channel:     ChannelSMS,
		To:          "+1234567890",
		Body:        "hello",
		ScheduledAt: &future,
	})
	if err != nil {
		t.Fatalf("scheduled send: %v", err)
	}

	stats, _ := svc.Stats(ctx)
	if stats.Queued != 1 {
		t.Fatalf("queued=%d, want 1", stats.Queued)
	}
}

func TestServiceStart_SchedulerRegistrationError(t *testing.T) {
	sch := scheduler.New()
	if _, err := sch.AddCron("messaging.dlq-retry", "* * * * *", func(context.Context) error { return nil }); err != nil {
		t.Fatalf("seed scheduler job: %v", err)
	}

	svc := New(Config{
		TaskStore:  store.NewMemory(store.DefaultMemConfig()),
		SMS:        &mockSMS{},
		Scheduler:  sch,
		Receipts:   NewMemReceiptStore(100),
		ConsumerID: "test-consumer",
	})
	err := svc.Start(context.Background())
	if err == nil {
		t.Fatal("expected scheduler registration error")
	}
	if !errors.Is(err, scheduler.ErrJobExists) {
		t.Fatalf("expected ErrJobExists, got %v", err)
	}
}

func TestDeliverTask_RenderFailureUpdatesReceiptAndPublishesFailedEvent(t *testing.T) {
	bus := pubsub.New()
	defer bus.Close()
	sub, err := bus.Subscribe("messaging.result", pubsub.SubOptions{BufferSize: 16})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	receipts := NewMemReceiptStore(100)
	svc := New(Config{
		TaskStore:         store.NewMemory(store.DefaultMemConfig()),
		SMS:               &mockSMS{},
		Bus:               bus,
		Receipts:          receipts,
		WorkerConcurrency: 1,
	})
	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer svc.Stop(ctx)

	err = svc.Send(ctx, SendRequest{
		ID:         "render-fail-1",
		Channel:    ChannelSMS,
		To:         "+1234567890",
		Template:   "missing-template",
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if !waitForFailed(svc, 1, 5*time.Second) {
		t.Fatal("timed out waiting for failed delivery")
	}
	receipt, ok := receipts.Get("render-fail-1")
	if !ok {
		t.Fatal("receipt not found")
	}
	if receipt.Status != "failed" {
		t.Fatalf("receipt status=%s, want failed", receipt.Status)
	}
	if receipt.Error == "" {
		t.Fatal("expected receipt error")
	}
	if svc.Monitor().ChannelHealth(ChannelSMS) != ChannelHealthy {
		t.Fatalf("expected monitor to stay healthy on render failure, got %s", svc.Monitor().ChannelHealth(ChannelSMS))
	}

	select {
	case msg := <-sub.C():
		raw, ok := msg.Data.(string)
		if !ok {
			t.Fatalf("event data type=%T, want string", msg.Data)
		}
		var result SendResult
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			t.Fatalf("decode result: %v", err)
		}
		if result.RequestID != "render-fail-1" {
			t.Fatalf("result request_id=%s, want render-fail-1", result.RequestID)
		}
		if result.Status != "failed" {
			t.Fatalf("result status=%s, want failed", result.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for failed event")
	}
}

func TestDeliverTask_ProviderFailurePublishesFailedEvent(t *testing.T) {
	bus := pubsub.New()
	defer bus.Close()
	sub, err := bus.Subscribe("messaging.result", pubsub.SubOptions{BufferSize: 16})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	receipts := NewMemReceiptStore(100)
	svc := New(Config{
		TaskStore:         store.NewMemory(store.DefaultMemConfig()),
		SMS:               &mockSMS{err: errors.New("gateway down")},
		Bus:               bus,
		Receipts:          receipts,
		WorkerConcurrency: 1,
	})
	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer svc.Stop(ctx)

	err = svc.Send(ctx, SendRequest{
		ID:         "provider-fail-1",
		Channel:    ChannelSMS,
		To:         "+1234567890",
		Body:       "hello",
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if !waitForFailed(svc, 1, 5*time.Second) {
		t.Fatal("timed out waiting for failed delivery")
	}
	if svc.Monitor().ChannelHealth(ChannelSMS) != ChannelDegraded {
		t.Fatalf("expected degraded monitor state, got %s", svc.Monitor().ChannelHealth(ChannelSMS))
	}
	receipt, ok := receipts.Get("provider-fail-1")
	if !ok {
		t.Fatal("receipt not found")
	}
	if receipt.Status != "failed" {
		t.Fatalf("receipt status=%s, want failed", receipt.Status)
	}

	select {
	case msg := <-sub.C():
		raw, ok := msg.Data.(string)
		if !ok {
			t.Fatalf("event data type=%T, want string", msg.Data)
		}
		var result SendResult
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			t.Fatalf("decode result: %v", err)
		}
		if result.RequestID != "provider-fail-1" {
			t.Fatalf("result request_id=%s, want provider-fail-1", result.RequestID)
		}
		if result.Status != "failed" {
			t.Fatalf("result status=%s, want failed", result.Status)
		}
		if result.Error == "" {
			t.Fatal("expected non-empty error in failed result")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for failed event")
	}
}
