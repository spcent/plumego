// Package mfa implements optional per-user TOTP-based multi-factor
// authentication (RFC 6238) and persists enrollment state in a KV store.
package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // RFC 6238 mandates HMAC-SHA1 for TOTP.
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// totpDigits is the number of decimal digits in a generated code.
	totpDigits = 6
	// totpStep is the time-step size used to derive the moving factor.
	totpStep = 30 * time.Second
	// totpSkew is the number of time-steps tolerated in either direction to
	// account for clock drift between client and server.
	totpSkew = 1
	// secretBytes is the number of random bytes used to generate a TOTP secret
	// (160 bits, matching the SHA-1 block size recommended by RFC 4226).
	secretBytes = 20
)

var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateSecret returns a new random base32-encoded TOTP secret.
func GenerateSecret() (string, error) {
	b := make([]byte, secretBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate totp secret: %w", err)
	}
	return base32Encoding.EncodeToString(b), nil
}

// otpauthURI builds an otpauth:// URI for the given secret, suitable for
// rendering as a QR code or manual entry by any external authenticator app.
func otpauthURI(issuer, account, secret string) string {
	label := url.PathEscape(issuer) + ":" + url.PathEscape(account)
	v := url.Values{}
	v.Set("secret", secret)
	v.Set("issuer", issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", strconv.Itoa(totpDigits))
	v.Set("period", strconv.Itoa(int(totpStep.Seconds())))
	return "otpauth://totp/" + label + "?" + v.Encode()
}

// GenerateCode computes the TOTP code for secret at time t.
func GenerateCode(secret string, t time.Time) (string, error) {
	return generateCodeForCounter(secret, counterAt(t))
}

// Validate reports whether code is a valid TOTP code for secret at time t,
// allowing for +/-1 time-step of clock skew. Comparison against each
// candidate code uses a timing-safe comparison. Any error (e.g. a malformed
// secret) results in rejection — this function fails closed.
func Validate(secret, code string, t time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}
	counter := counterAt(t)
	for skew := -totpSkew; skew <= totpSkew; skew++ {
		candidate, err := generateCodeForCounter(secret, counter+int64(skew))
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

func counterAt(t time.Time) int64 {
	return t.Unix() / int64(totpStep.Seconds())
}

func generateCodeForCounter(secret string, counter int64) (string, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", fmt.Errorf("decode totp secret: %w", err)
	}
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(counterBytes[:])
	sum := mac.Sum(nil)

	// Dynamic truncation per RFC 4226 section 5.3.
	offset := sum[len(sum)-1] & 0x0f
	truncated := (uint32(sum[offset])&0x7f)<<24 |
		uint32(sum[offset+1])<<16 |
		uint32(sum[offset+2])<<8 |
		uint32(sum[offset+3])

	mod := uint32(1)
	for i := 0; i < totpDigits; i++ {
		mod *= 10
	}
	code := truncated % mod
	return fmt.Sprintf("%0*d", totpDigits, code), nil
}

func decodeSecret(secret string) ([]byte, error) {
	clean := strings.ToUpper(strings.TrimSpace(secret))
	clean = strings.ReplaceAll(clean, " ", "")
	return base32Encoding.DecodeString(clean)
}
