package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecureCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://example.test/api/auth/login", nil)
	if secureCookie(req) {
		t.Fatal("secureCookie(http) = true, want false")
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	if !secureCookie(req) {
		t.Fatal("secureCookie(X-Forwarded-Proto=https) = false, want true")
	}
}

func TestSecureEqual(t *testing.T) {
	if !secureEqual("secret", "secret") {
		t.Fatal("secureEqual matching strings = false")
	}
	if secureEqual("secret", "different") {
		t.Fatal("secureEqual different strings = true")
	}
}

func TestLoginLimiter(t *testing.T) {
	limiter := NewLoginLimiter(2, time.Minute)
	if !limiter.Allow("client") {
		t.Fatal("new client should be allowed")
	}
	limiter.RecordFailure("client")
	if !limiter.Allow("client") {
		t.Fatal("client should be allowed before threshold")
	}
	limiter.RecordFailure("client")
	if limiter.Allow("client") {
		t.Fatal("client should be blocked at threshold")
	}
	limiter.Reset("client")
	if !limiter.Allow("client") {
		t.Fatal("client should be allowed after reset")
	}
}
