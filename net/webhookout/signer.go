package webhookout

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SignV1(secret string, timestamp string, rawBody []byte) string {
	// v1 = hex(hmac_sha256(secret, timestamp + "." + rawBody))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}

func ConstantTimeEqualHex(aHex, bHex string) bool {
	a, errA := hex.DecodeString(aHex)
	b, errB := hex.DecodeString(bHex)
	if errA != nil || errB != nil {
		return false
	}
	return hmac.Equal(a, b)
}
