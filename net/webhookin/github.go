package webhookin

import (
	"errors"
	"net/http"
	"time"
)

var (
	ErrGitHubSignature       = errors.New("invalid github signature")
	ErrGitHubMissingHeader   = errors.New("missing github signature header")
	ErrGitHubInvalidEncoding = errors.New("invalid hex encoding")
)

// GitHubVerifyOptions configures GitHub webhook verification.
type GitHubVerifyOptions struct {
	MaxBody int64

	// EnableReplayProtection enables timestamp-based replay protection.
	// GitHub includes X-Hub-Delivery header (unique ID) that can be used with NonceStore.
	EnableReplayProtection bool

	// Tolerance specifies the time window for accepting requests.
	// Only used if EnableReplayProtection is true.
	// Default: 5 minutes
	Tolerance time.Duration

	// NonceStore for replay protection using X-Hub-Delivery header.
	// Only used if EnableReplayProtection is true.
	NonceStore NonceStore

	// Now function for testing. Defaults to time.Now.
	Now func() time.Time
}

// VerifyGitHub verifies X-Hub-Signature-256: "sha256=<hex>"
// Returns the verified body or an error.
//
// This function now uses the shared VerifyHMAC implementation and optionally
// provides replay protection using the X-Hub-Delivery header as a nonce.
func VerifyGitHub(r *http.Request, secret string, maxBody int64) ([]byte, error) {
	opts := GitHubVerifyOptions{
		MaxBody:                maxBody,
		EnableReplayProtection: false,
	}
	return VerifyGitHubWithOptions(r, secret, opts)
}

// VerifyGitHubWithOptions verifies GitHub webhooks with configurable replay protection.
//
// GitHub webhook replay protection works by:
// 1. Using X-Hub-Delivery header as a unique delivery ID (nonce)
// 2. Optionally checking request timestamp to prevent old requests
//
// Example with replay protection:
//
//	nonceStore := NewMemoryNonceStore(10 * time.Minute)
//	opts := GitHubVerifyOptions{
//	    MaxBody: 1 << 20,
//	    EnableReplayProtection: true,
//	    Tolerance: 5 * time.Minute,
//	    NonceStore: nonceStore,
//	}
//	body, err := VerifyGitHubWithOptions(r, secret, opts)
func VerifyGitHubWithOptions(r *http.Request, secret string, opts GitHubVerifyOptions) ([]byte, error) {
	if opts.MaxBody <= 0 {
		opts.MaxBody = 1 << 20 // 1MB default
	}

	cfg := HMACConfig{
		Secret:       []byte(secret),
		Header:       "X-Hub-Signature-256",
		Prefix:       "sha256=",
		MaxBody:      opts.MaxBody,
		Algorithm:    HashSHA256,
		Encoding:     EncodingHex,
		SignedFormat: PayloadRaw,
	}

	// Enable replay protection if requested
	if opts.EnableReplayProtection {
		tolerance := opts.Tolerance
		if tolerance <= 0 {
			tolerance = 5 * time.Minute
		}

		cfg.Replay = HMACReplayConfig{
			// GitHub uses X-Hub-Delivery as a unique delivery ID
			NonceHeader: "X-Hub-Delivery",
			Tolerance:   tolerance,
			NonceStore:  opts.NonceStore,
			Now:         opts.Now,
		}
	}

	result, err := VerifyHMAC(r, cfg)
	if err != nil {
		// Map generic errors to GitHub-specific errors for backwards compatibility
		ve, ok := err.(*VerifyError)
		if ok {
			switch ve.Code {
			case CodeMissingSignature:
				return nil, ErrGitHubMissingHeader
			case CodeInvalidSignature:
				return nil, ErrGitHubSignature
			case CodeInvalidEncoding:
				return nil, ErrGitHubInvalidEncoding
			}
		}
		return nil, err
	}

	return result.Body, nil
}
