package websocket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestNewHS256TokenAuthRejectsWeakSecret(t *testing.T) {
	if _, err := NewHS256TokenAuth([]byte("short")); !errors.Is(err, ErrWeakJWTSecret) {
		t.Fatalf("NewHS256TokenAuth error = %v, want ErrWeakJWTSecret", err)
	}
}

func TestHS256TokenAuthRejectsMalformedExp(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	tokenAuth, err := NewHS256TokenAuth(secret)
	if err != nil {
		t.Fatalf("NewHS256TokenAuth: %v", err)
	}

	tests := []struct {
		name    string
		payload string
	}{
		{name: "string exp", payload: `{"sub":"user1","exp":"soon"}`},
		{name: "fractional exp", payload: `{"sub":"user1","exp":123.5}`},
		{name: "negative exp", payload: `{"sub":"user1","exp":-1}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tokenAuth.AuthenticateToken(signHS256TestToken(secret, tc.payload))
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("AuthenticateToken error = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestHS256TokenAuthAcceptsIntegerExp(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	tokenAuth, err := NewHS256TokenAuth(secret)
	if err != nil {
		t.Fatalf("NewHS256TokenAuth: %v", err)
	}

	payload := `{"sub":"user1","exp":` + itoaUnix(time.Now().Add(time.Minute).Unix()) + `}`
	if _, err := tokenAuth.AuthenticateToken(signHS256TestToken(secret, payload)); err != nil {
		t.Fatalf("AuthenticateToken error = %v", err)
	}
}

func signHS256TestToken(secret []byte, payloadJSON string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(payloadJSON))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return header + "." + payload + "." + sig
}

func itoaUnix(v int64) string {
	return strconv.FormatInt(v, 10)
}
