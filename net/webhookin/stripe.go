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
		opt.MaxBody = 1 << 20
	}
	if opt.Tolerance <= 0 {
		opt.Tolerance = 5 * time.Minute
	}

	header := strings.TrimSpace(r.Header.Get("Stripe-Signature"))
	if header == "" {
		return nil, ErrStripeSignature
	}

	t, sigs, err := parseStripeSigHeader(header)
	if err != nil {
		return nil, ErrStripeSignature
	}

	// replay protection
	now := nowFn().UTC()
	ts := time.Unix(t, 0).UTC()
	if ts.After(now.Add(opt.Tolerance)) || ts.Before(now.Add(-opt.Tolerance)) {
		return nil, ErrStripeSignature
	}

	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, opt.MaxBody))
	if err != nil {
		return nil, err
	}

	signed := strconv.FormatInt(t, 10) + "." + string(body)
	mac := hmac.New(sha256.New, []byte(endpointSecret))
	mac.Write([]byte(signed))
	expected := mac.Sum(nil)

	// any v1 matches is OK
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
		return 0, nil, ErrStripeSignature
	}
	return timestamp, v1s, nil
}
