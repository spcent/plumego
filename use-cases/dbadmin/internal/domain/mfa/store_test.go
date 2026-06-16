package mfa

import (
	"errors"
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	t.Cleanup(func() { _ = kv.Close() })
	return NewStore(kv)
}

func TestStore_StartEnrollment(t *testing.T) {
	s := newTestStore(t)
	secret, uri, err := s.StartEnrollment("alice")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	if secret == "" {
		t.Fatal("StartEnrollment returned empty secret")
	}
	if uri == "" {
		t.Fatal("StartEnrollment returned empty otpauth URI")
	}
	if s.IsEnabled("alice") {
		t.Fatal("MFA should not be enabled immediately after StartEnrollment")
	}
	enr, err := s.Get("alice")
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if enr.Secret != secret {
		t.Fatalf("stored secret = %q, want %q", enr.Secret, secret)
	}
}

func TestStore_ConfirmEnrollment_success(t *testing.T) {
	s := newTestStore(t)
	secret, _, err := s.StartEnrollment("bob")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	code, err := GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	if err := s.ConfirmEnrollment("bob", code); err != nil {
		t.Fatalf("ConfirmEnrollment error = %v", err)
	}
	if !s.IsEnabled("bob") {
		t.Fatal("MFA should be enabled after ConfirmEnrollment with a valid code")
	}
}

func TestStore_ConfirmEnrollment_wrongCode(t *testing.T) {
	s := newTestStore(t)
	if _, _, err := s.StartEnrollment("carol"); err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	if err := s.ConfirmEnrollment("carol", "000000"); err == nil {
		t.Fatal("ConfirmEnrollment should fail with an incorrect code")
	}
	if s.IsEnabled("carol") {
		t.Fatal("MFA must not be enabled after a failed confirmation")
	}
}

func TestStore_ConfirmEnrollment_noEnrollment(t *testing.T) {
	s := newTestStore(t)
	if err := s.ConfirmEnrollment("nobody", "123456"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ConfirmEnrollment error = %v, want ErrNotFound", err)
	}
}

func TestStore_Disable(t *testing.T) {
	s := newTestStore(t)
	secret, _, err := s.StartEnrollment("dave")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	code, _ := GenerateCode(secret, time.Now())
	if err := s.ConfirmEnrollment("dave", code); err != nil {
		t.Fatalf("ConfirmEnrollment error = %v", err)
	}
	if !s.IsEnabled("dave") {
		t.Fatal("expected MFA enabled before Disable")
	}
	if err := s.Disable("dave"); err != nil {
		t.Fatalf("Disable error = %v", err)
	}
	if s.IsEnabled("dave") {
		t.Fatal("MFA should be disabled after Disable")
	}
}

func TestStore_Disable_noEnrollment_isNoop(t *testing.T) {
	s := newTestStore(t)
	if err := s.Disable("ghost"); err != nil {
		t.Fatalf("Disable on a user with no enrollment should not error, got %v", err)
	}
}

func TestStore_VerifyCode(t *testing.T) {
	s := newTestStore(t)
	secret, _, err := s.StartEnrollment("erin")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	code, _ := GenerateCode(secret, time.Now())

	// Not yet enabled: VerifyCode must fail closed even with a correct code.
	if s.VerifyCode("erin", code) {
		t.Fatal("VerifyCode must reject codes before enrollment is confirmed")
	}

	if err := s.ConfirmEnrollment("erin", code); err != nil {
		t.Fatalf("ConfirmEnrollment error = %v", err)
	}

	freshCode, _ := GenerateCode(secret, time.Now())
	if !s.VerifyCode("erin", freshCode) {
		t.Fatal("VerifyCode should accept a valid code after enrollment is enabled")
	}
	if s.VerifyCode("erin", "000000") {
		t.Fatal("VerifyCode should reject an incorrect code")
	}
}

func TestStore_VerifyCode_unknownUser(t *testing.T) {
	s := newTestStore(t)
	if s.VerifyCode("unknown", "123456") {
		t.Fatal("VerifyCode must fail closed for an unknown user")
	}
}

func TestStore_Challenge_roundTrip(t *testing.T) {
	s := newTestStore(t)
	token, err := s.CreateChallenge("frank")
	if err != nil {
		t.Fatalf("CreateChallenge error = %v", err)
	}
	if token == "" {
		t.Fatal("CreateChallenge returned empty token")
	}
	username, err := s.ConsumeChallenge(token)
	if err != nil {
		t.Fatalf("ConsumeChallenge error = %v", err)
	}
	if username != "frank" {
		t.Fatalf("ConsumeChallenge username = %q, want %q", username, "frank")
	}
}

func TestStore_Challenge_singleUse(t *testing.T) {
	s := newTestStore(t)
	token, err := s.CreateChallenge("gina")
	if err != nil {
		t.Fatalf("CreateChallenge error = %v", err)
	}
	if _, err := s.ConsumeChallenge(token); err != nil {
		t.Fatalf("first ConsumeChallenge error = %v", err)
	}
	if _, err := s.ConsumeChallenge(token); !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("second ConsumeChallenge error = %v, want ErrChallengeNotFound", err)
	}
}

func TestStore_Challenge_unknownToken(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.ConsumeChallenge("does-not-exist"); !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("ConsumeChallenge error = %v, want ErrChallengeNotFound", err)
	}
}
