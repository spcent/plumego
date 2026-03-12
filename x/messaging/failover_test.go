package messaging

import (
	"context"
	"errors"
	"testing"
)

type failingSMS struct{}

func (f *failingSMS) Name() string { return "failing-sms" }
func (f *failingSMS) Send(_ context.Context, _ SMSMessage) (*SMSResult, error) {
	return nil, errors.New("gateway down")
}

type failingEmail struct{}

func (f *failingEmail) Name() string { return "failing-email" }
func (f *failingEmail) Send(_ context.Context, _ EmailMessage) (*EmailResult, error) {
	return nil, errors.New("smtp error")
}

func TestFailoverSMS_FallsBackToSecond(t *testing.T) {
	primary := &failingSMS{}
	backup := &mockSMS{}

	fo := NewFailoverSMSProvider(primary, backup)
	if fo.Name() != "failover(failing-sms)" {
		t.Fatalf("name=%s", fo.Name())
	}

	result, err := fo.Send(context.Background(), SMSMessage{To: "+1", Body: "hi"})
	if err != nil {
		t.Fatalf("expected fallback success, got: %v", err)
	}
	if result.ProviderID == "" {
		t.Fatal("expected provider ID from backup")
	}

	backup.mu.Lock()
	if len(backup.calls) != 1 {
		t.Fatalf("backup calls=%d, want 1", len(backup.calls))
	}
	backup.mu.Unlock()
}

func TestFailoverSMS_AllFail(t *testing.T) {
	fo := NewFailoverSMSProvider(&failingSMS{}, &failingSMS{})
	_, err := fo.Send(context.Background(), SMSMessage{To: "+1", Body: "hi"})
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
	if !errors.Is(err, ErrProviderFailure) {
		t.Fatalf("expected ErrProviderFailure, got %v", err)
	}
}

func TestFailoverSMS_Empty(t *testing.T) {
	fo := NewFailoverSMSProvider()
	if fo.Name() != "failover(empty)" {
		t.Fatalf("name=%s", fo.Name())
	}
	_, err := fo.Send(context.Background(), SMSMessage{})
	if err == nil {
		t.Fatal("expected error with no providers")
	}
}

func TestFailoverEmail_FallsBackToSecond(t *testing.T) {
	primary := &failingEmail{}
	backup := &mockEmail{}

	fo := NewFailoverEmailProvider(primary, backup)

	result, err := fo.Send(context.Background(), EmailMessage{To: "a@b.com", Subject: "s", Body: "b"})
	if err != nil {
		t.Fatalf("expected fallback success, got: %v", err)
	}
	if result.MessageID == "" {
		t.Fatal("expected message ID from backup")
	}
}

func TestFailoverEmail_AllFail(t *testing.T) {
	fo := NewFailoverEmailProvider(&failingEmail{}, &failingEmail{})
	_, err := fo.Send(context.Background(), EmailMessage{To: "a@b.com"})
	if !errors.Is(err, ErrProviderFailure) {
		t.Fatalf("expected ErrProviderFailure, got %v", err)
	}
}

func TestFailoverSMS_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	fo := NewFailoverSMSProvider(&failingSMS{}, &mockSMS{})
	_, err := fo.Send(ctx, SMSMessage{To: "+1", Body: "hi"})
	if err == nil {
		t.Fatal("expected context error")
	}
}
