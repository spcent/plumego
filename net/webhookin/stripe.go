package webhookin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrStripeSignature        = errors.New("invalid stripe signature")
	ErrStripeMissingHeader    = errors.New("missing stripe signature header")
	ErrStripeInvalidTimestamp = errors.New("invalid timestamp in signature")
	ErrStripeExpired          = errors.New("signature expired")
	ErrStripeInvalidEncoding  = errors.New("invalid hex encoding")
)

// StripeVerifyOptions configures Stripe webhook verification.
type StripeVerifyOptions struct {
	MaxBody   int64
	Tolerance time.Duration // e.g. 5 minutes
	Now       func() time.Time
}

// StripeSignatureExtractor extracts t and v1 values from Stripe-Signature header.
type StripeSignatureExtractor struct {
	Timestamp  int64
	Signatures []string
}

// parseStripeSigHeader parses the Stripe-Signature header format: "t=...,v1=...,v1=..."
func parseStripeSigHeader(h string) (StripeSignatureExtractor, error) {
	var result StripeSignatureExtractor
	parts := strings.Split(h, ",")
	var tStr string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "t=") {
			tStr = strings.TrimPrefix(p, "t=")
		} else if strings.HasPrefix(p, "v1=") {
			result.Signatures = append(result.Signatures, strings.TrimPrefix(p, "v1="))
		}
	}

	if tStr == "" || len(result.Signatures) == 0 {
		return result, ErrStripeSignature
	}

	timestamp, err := strconv.ParseInt(tStr, 10, 64)
	if err != nil {
		return result, ErrStripeInvalidTimestamp
	}
	result.Timestamp = timestamp

	return result, nil
}

// VerifyStripe verifies "Stripe-Signature" header.
// Signed payload = "<t>.<rawBody>"
// Header contains: "t=...,v1=...,v1=..."
//
// Now uses the shared VerifyHMAC implementation with custom logic for Stripe's
// multi-signature format.
func VerifyStripe(r *http.Request, endpointSecret string, opt StripeVerifyOptions) ([]byte, error) {
	if opt.MaxBody <= 0 {
		opt.MaxBody = 1 << 20 // 1MB default
	}
	if opt.Tolerance <= 0 {
		opt.Tolerance = 5 * time.Minute
	}

	// Parse Stripe-Signature header to extract timestamp
	header := strings.TrimSpace(r.Header.Get("Stripe-Signature"))
	if header == "" {
		return nil, ErrStripeMissingHeader
	}

	parsed, err := parseStripeSigHeader(header)
	if err != nil {
		return nil, err
	}

	// Inject timestamp into a temporary header for VerifyHMAC
	// This allows us to use the shared verification logic
	r.Header.Set("X-Stripe-Timestamp", strconv.FormatInt(parsed.Timestamp, 10))

	// Try each v1 signature until one matches
	var lastErr error
	for _, sig := range parsed.Signatures {
		// Temporarily set this v1 signature
		r.Header.Set("Stripe-Signature", sig)

		cfg := HMACConfig{
			Secret:       []byte(endpointSecret),
			Header:       "Stripe-Signature",
			MaxBody:      opt.MaxBody,
			Algorithm:    HashSHA256,
			Encoding:     EncodingHex,
			SignedFormat: PayloadTimestampBody,
			Replay: HMACReplayConfig{
				TimestampHeader: "X-Stripe-Timestamp",
				Tolerance:       opt.Tolerance,
				Now:             opt.Now,
			},
		}

		result, err := VerifyHMAC(r, cfg)
		if err == nil {
			// Clean up temporary headers
			r.Header.Del("X-Stripe-Timestamp")
			r.Header.Set("Stripe-Signature", header) // restore original
			return result.Body, nil
		}
		lastErr = err
	}

	// Clean up temporary headers
	r.Header.Del("X-Stripe-Timestamp")
	r.Header.Set("Stripe-Signature", header) // restore original

	// Map generic errors to Stripe-specific errors for backwards compatibility
	if lastErr != nil {
		ve, ok := lastErr.(*VerifyError)
		if ok {
			switch ve.Code {
			case CodeMissingSignature:
				return nil, ErrStripeMissingHeader
			case CodeInvalidSignature:
				return nil, ErrStripeSignature
			case CodeInvalidEncoding:
				return nil, ErrStripeInvalidEncoding
			case CodeTimestampExpired:
				return nil, ErrStripeExpired
			case CodeInvalidTimestamp:
				return nil, ErrStripeInvalidTimestamp
			}
		}
	}

	return nil, ErrStripeSignature
}
