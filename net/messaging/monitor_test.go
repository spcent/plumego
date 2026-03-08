package messaging

import (
	"errors"
	"testing"
	"time"
)

func TestChannelMonitor_InitialHealthy(t *testing.T) {
	sms := &mockSMS{}
	email := &mockEmail{}
	m := NewChannelMonitor(sms, email, nil)

	statuses := m.Status()
	if len(statuses) != 2 {
		t.Fatalf("channels=%d, want 2", len(statuses))
	}
	for _, s := range statuses {
		if s.State != ChannelHealthy {
			t.Fatalf("channel %s: state=%s, want healthy", s.Channel, s.State)
		}
	}
}

func TestChannelMonitor_RecordSuccess(t *testing.T) {
	m := NewChannelMonitor(&mockSMS{}, nil, nil)
	m.RecordSuccess(ChannelSMS, 50*time.Millisecond)

	if m.ChannelHealth(ChannelSMS) != ChannelHealthy {
		t.Fatal("expected healthy after success")
	}
}

func TestChannelMonitor_RecordFailure(t *testing.T) {
	m := NewChannelMonitor(&mockSMS{}, nil, nil)
	m.RecordFailure(ChannelSMS, errors.New("timeout"))

	if m.ChannelHealth(ChannelSMS) != ChannelDegraded {
		t.Fatal("expected degraded after failure")
	}
}

func TestChannelMonitor_RecoveryAfterFailure(t *testing.T) {
	m := NewChannelMonitor(&mockSMS{}, nil, nil)
	m.RecordFailure(ChannelSMS, errors.New("timeout"))
	m.RecordSuccess(ChannelSMS, 10*time.Millisecond)

	if m.ChannelHealth(ChannelSMS) != ChannelHealthy {
		t.Fatal("expected healthy after recovery")
	}
}

func TestChannelMonitor_UnknownChannel(t *testing.T) {
	m := NewChannelMonitor(nil, nil, nil)
	if m.ChannelHealth(ChannelSMS) != ChannelUnhealthy {
		t.Fatal("expected unhealthy for unconfigured channel")
	}
}

func TestChannelMonitor_StatusContainsProvider(t *testing.T) {
	m := NewChannelMonitor(&mockSMS{}, nil, nil)
	for _, s := range m.Status() {
		if s.Channel == ChannelSMS && s.Provider != "mock-sms" {
			t.Fatalf("provider=%s, want mock-sms", s.Provider)
		}
	}
}
