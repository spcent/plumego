package webhookout

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrInvalidHex = errors.New("invalid hex encoding")

// SignV1 creates a v1 webhook signature using HMAC-SHA256.
// Format: hex(hmac_sha256(secret, timestamp + "." + rawBody))
func SignV1(secret string, timestamp string, rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}

// ConstantTimeEqualHex performs constant-time comparison of two hex-encoded strings.
// Returns false if either string is invalid hex.
func ConstantTimeEqualHex(aHex, bHex string) bool {
	a, errA := hex.DecodeString(aHex)
	b, errB := hex.DecodeString(bHex)
	if errA != nil || errB != nil {
		return false
	}
	return hmac.Equal(a, b)
}

// VerifySignature verifies a webhook signature using constant-time comparison.
func VerifySignature(secret, timestamp, signature string, payload []byte) error {
	expected := SignV1(secret, timestamp, payload)
	if !ConstantTimeEqualHex(signature, expected) {
		return errors.New("signature verification failed")
	}
	return nil
}
