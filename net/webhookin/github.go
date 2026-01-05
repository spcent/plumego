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
var ErrGitHubMissingHeader = errors.New("missing github signature header")
var ErrGitHubInvalidEncoding = errors.New("invalid hex encoding")

// VerifyGitHub verifies X-Hub-Signature-256: "sha256=<hex>"
// Returns the verified body or an error.
func VerifyGitHub(r *http.Request, secret string, maxBody int64) ([]byte, error) {
	sig := strings.TrimSpace(r.Header.Get("X-Hub-Signature-256"))
	if sig == "" {
		return nil, ErrGitHubMissingHeader
	}
	if !strings.HasPrefix(sig, "sha256=") {
		return nil, ErrGitHubSignature
	}
	wantHex := strings.TrimPrefix(sig, "sha256=")
	if wantHex == "" {
		return nil, ErrGitHubSignature
	}

	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxBody))
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	got := mac.Sum(nil)

	want, err := hex.DecodeString(wantHex)
	if err != nil {
		return nil, ErrGitHubInvalidEncoding
	}
	if !hmac.Equal(got, want) {
		return nil, ErrGitHubSignature
	}
	return body, nil
}
