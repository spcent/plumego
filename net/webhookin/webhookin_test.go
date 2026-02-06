package webhookin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestDeduper(t *testing.T) {
	d := NewDeduper(10 * time.Millisecond)
	if d.SeenBefore("id") {
		t.Fatalf("first call should not be seen before")
	}
	if !d.SeenBefore("id") {
		t.Fatalf("second call should be detected as duplicate")
	}

	time.Sleep(20 * time.Millisecond)
	if d.SeenBefore("id") {
		t.Fatalf("entry should expire after ttl")
	}
}

func TestVerifyGitHub(t *testing.T) {
	body := []byte(`{"ok":true}`)
	secret := "topsecret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("X-Hub-Signature-256", sig)

	got, err := VerifyGitHub(req, secret, int64(len(body))+10)
	if err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("body mismatch")
	}

	badReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	badReq.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	if _, err := VerifyGitHub(badReq, secret, 1024); err == nil {
		t.Fatalf("expected signature error")
	}
}

func TestVerifyStripe(t *testing.T) {
	body := []byte("payload")
	secret := "endpoint"
	now := time.Now()
	timestamp := now.Unix()

	mac := hmac.New(sha256.New, []byte(secret))
	tsStr := strconv.FormatInt(timestamp, 10)
	mac.Write([]byte(tsStr + "." + string(body)))
	signature := hex.EncodeToString(mac.Sum(nil))

	header := http.Header{}
	header.Set("Stripe-Signature", "t="+tsStr+",v1="+signature)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header = header

	payload, err := VerifyStripe(req, secret, StripeVerifyOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("expected valid stripe signature, got %v", err)
	}
	if !bytes.Equal(payload, body) {
		t.Fatalf("unexpected payload")
	}

	staleReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	staleReq.Header = header
	if _, err := VerifyStripe(staleReq, secret, StripeVerifyOptions{Now: func() time.Time { return now.Add(10 * time.Minute) }}); err == nil {
		t.Fatalf("expected tolerance failure")
	}
}

func TestParseStripeHeaderErrors(t *testing.T) {
	if _, err := parseStripeSigHeader("v1=abc"); err == nil {
		t.Fatalf("expected error without timestamp")
	}
	if _, err := parseStripeSigHeader("t=abc,v1=abc"); err == nil {
		t.Fatalf("expected parse int error")
	}
}
