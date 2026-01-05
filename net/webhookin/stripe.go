package webhookin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrStripeSignature = errors.New("invalid stripe signature")
var ErrStripeMissingHeader = errors.New("missing stripe signature header")
var ErrStripeInvalidTimestamp = errors.New("invalid timestamp in signature")
var ErrStripeExpired = errors.New("signature expired")
var ErrStripeInvalidEncoding = errors.New("invalid hex encoding")

// StripeVerifyOptions configures Stripe webhook verification.
type StripeVerifyOptions struct {
	MaxBody   int64
	Tolerance time.Duration // e.g. 5 minutes
	Now       func() time.Time
}

// VerifyStripe verifies "Stripe-Signature" header.
// Signed payload = "<t>.<rawBody>"
// Header contains: "t=...,v1=...,v1=..."
func VerifyStripe(r *http.Request, endpointSecret string, opt StripeVerifyOptions) ([]byte, error) {
	nowFn := opt.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	if opt.MaxBody <= 0 {
		opt.MaxBody = 1 << 20 // 1MB default
	}
	if opt.Tolerance <= 0 {
		opt.Tolerance = 5 * time.Minute
	}

	header := strings.TrimSpace(r.Header.Get("Stripe-Signature"))
	if header == "" {
		return nil, ErrStripeMissingHeader
	}

	t, sigs, err := parseStripeSigHeader(header)
	if err != nil {
		return nil, err
	}

	// replay protection with tolerance window
	now := nowFn().UTC()
	ts := time.Unix(t, 0).UTC()
	earliest := now.Add(-opt.Tolerance)
	latest := now.Add(opt.Tolerance)

	if ts.Before(earliest) {
		return nil, ErrStripeExpired
	}
	if ts.After(latest) {
		return nil, ErrStripeInvalidTimestamp
	}

	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, opt.MaxBody))
	if err != nil {
		return nil, err
	}

	signed := strconv.FormatInt(t, 10) + "." + string(body)
	mac := hmac.New(sha256.New, []byte(endpointSecret))
	mac.Write([]byte(signed))
	expected := mac.Sum(nil)

	// any v1 signature match is acceptable
	for _, v1 := range sigs {
		b, err := hex.DecodeString(v1)
		if err != nil {
			continue
		}
		if hmac.Equal(expected, b) {
			return body, nil
		}
	}
	return nil, ErrStripeSignature
}

func parseStripeSigHeader(h string) (timestamp int64, v1s []string, err error) {
	parts := strings.Split(h, ",")
	var tStr string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "t=") {
			tStr = strings.TrimPrefix(p, "t=")
		} else if strings.HasPrefix(p, "v1=") {
			v1s = append(v1s, strings.TrimPrefix(p, "v1="))
		}
	}
	if tStr == "" || len(v1s) == 0 {
		return 0, nil, ErrStripeSignature
	}
	timestamp, err = strconv.ParseInt(tStr, 10, 64)
	if err != nil {
		return 0, nil, ErrStripeInvalidTimestamp
	}
	return timestamp, v1s, nil
}
