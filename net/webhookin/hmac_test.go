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

func TestVerifyHMACBasic(t *testing.T) {
	secret := []byte("secret")
	body := []byte("payload")

	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBuffer(body))
	req.Header.Set("X-Signature", signature)

	result, err := VerifyHMAC(req, HMACConfig{
		Secret:   secret,
		Header:   "X-Signature",
		Encoding: EncodingHex,
	})
	if err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
	if !bytes.Equal(result.Body, body) {
		t.Fatalf("body mismatch")
	}
}

func TestVerifyHMACReplayProtection(t *testing.T) {
	secret := []byte("secret")
	body := []byte("payload")
	now := time.Now().UTC()
	nonceStore := NewMemoryNonceStore(2 * time.Minute)

	signedPayload := []byte(strconvUnix(now) + ".nonce-1." + string(body))
	mac := hmac.New(sha256.New, secret)
	mac.Write(signedPayload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBuffer(body))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", strconvUnix(now))
	req.Header.Set("X-Nonce", "nonce-1")

	cfg := HMACConfig{
		Secret:   secret,
		Header:   "X-Signature",
		Encoding: EncodingHex,
		Replay: HMACReplayConfig{
			TimestampHeader: "X-Timestamp",
			NonceHeader:     "X-Nonce",
			Tolerance:       2 * time.Minute,
			Now:             func() time.Time { return now },
			NonceStore:      nonceStore,
		},
	}

	if _, err := VerifyHMAC(req, cfg); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}

	replayReq := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBuffer(body))
	replayReq.Header.Set("X-Signature", signature)
	replayReq.Header.Set("X-Timestamp", strconvUnix(now))
	replayReq.Header.Set("X-Nonce", "nonce-1")

	if _, err := VerifyHMAC(replayReq, cfg); err == nil || ErrorCodeOf(err) != CodeReplayDetected {
		t.Fatalf("expected replay detected error, got %v", err)
	}
}

func TestVerifyHMACIPAllowlist(t *testing.T) {
	secret := []byte("secret")
	body := []byte("payload")

	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBuffer(body))
	req.Header.Set("X-Signature", signature)
	req.RemoteAddr = "203.0.113.10:1234"

	allow, err := NewIPAllowlist([]string{"203.0.113.0/24"})
	if err != nil {
		t.Fatalf("allowlist: %v", err)
	}

	if _, err := VerifyHMAC(req, HMACConfig{
		Secret:      secret,
		Header:      "X-Signature",
		Encoding:    EncodingHex,
		IPAllowlist: allow,
	}); err != nil {
		t.Fatalf("expected ip allowed, got %v", err)
	}

	denyReq := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBuffer(body))
	denyReq.Header.Set("X-Signature", signature)
	denyReq.RemoteAddr = "198.51.100.1:2345"

	if _, err := VerifyHMAC(denyReq, HMACConfig{
		Secret:      secret,
		Header:      "X-Signature",
		Encoding:    EncodingHex,
		IPAllowlist: allow,
	}); err == nil || ErrorCodeOf(err) != CodeIPDenied {
		t.Fatalf("expected ip denied error, got %v", err)
	}
}

func TestVerifyHMACTimestampExpired(t *testing.T) {
	secret := []byte("secret")
	body := []byte("payload")
	now := time.Now().UTC()
	old := now.Add(-10 * time.Minute)

	signedPayload := []byte(strconvUnix(old) + "." + string(body))
	mac := hmac.New(sha256.New, secret)
	mac.Write(signedPayload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBuffer(body))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", strconvUnix(old))

	_, err := VerifyHMAC(req, HMACConfig{
		Secret:   secret,
		Header:   "X-Signature",
		Encoding: EncodingHex,
		Replay: HMACReplayConfig{
			TimestampHeader: "X-Timestamp",
			Tolerance:       2 * time.Minute,
			Now:             func() time.Time { return now },
		},
		SignedFormat: PayloadTimestampBody,
	})
	if err == nil || ErrorCodeOf(err) != CodeTimestampExpired {
		t.Fatalf("expected expired timestamp, got %v", err)
	}
}

func strconvUnix(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}
