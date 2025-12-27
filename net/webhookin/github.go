package webhookin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
)

var ErrGitHubSignature = errors.New("invalid github signature")

// VerifyGitHub verifies X-Hub-Signature-256: "sha256=<hex>"
func VerifyGitHub(r *http.Request, secret string, maxBody int64) ([]byte, error) {
	sig := strings.TrimSpace(r.Header.Get("X-Hub-Signature-256"))
	if !strings.HasPrefix(sig, "sha256=") {
		return nil, ErrGitHubSignature
	}
	wantHex := strings.TrimPrefix(sig, "sha256=")

	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxBody))
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	got := mac.Sum(nil)

	want, err := hex.DecodeString(wantHex)
	if err != nil {
		return nil, ErrGitHubSignature
	}
	if !hmac.Equal(got, want) {
		return nil, ErrGitHubSignature
	}
	return body, nil
}
