package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spcent/plumego/x/webhook"
)

type webhookEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}

	allowlist, err := webhook.NewIPAllowlist([]string{"203.0.113.0/24", "198.51.100.10"})
	if err != nil {
		log.Fatalf("allowlist: %v", err)
	}
	nonceStore := webhook.NewMemoryNonceStore(10 * time.Minute)

	mux := http.NewServeMux()
	mux.Handle("/webhooks/inbound", genericHMACHandler([]byte(secret), allowlist, nonceStore))

	addr := ":8080"
	log.Printf("webhook example listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func genericHMACHandler(secret []byte, allowlist *webhook.IPAllowlist, nonceStore webhook.NonceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := webhook.VerifyHMAC(r, webhook.HMACConfig{
			Secret:   secret,
			Header:   "X-Signature",
			Prefix:   "sha256=",
			Encoding: webhook.EncodingHex,
			MaxBody:  1 << 20,
			Replay: webhook.HMACReplayConfig{
				TimestampHeader: "X-Timestamp",
				NonceHeader:     "X-Nonce",
				Tolerance:       5 * time.Minute,
				NonceStore:      nonceStore,
			},
			IPAllowlist: allowlist,
		})
		if err != nil {
			http.Error(w, "invalid webhook signature", webhook.HTTPStatus(err))
			return
		}

		var event webhookEvent
		if err := json.Unmarshal(result.Body, &event); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		log.Printf("webhook accepted id=%s type=%s ip=%s", event.ID, event.Type, result.IP)
		w.WriteHeader(http.StatusOK)
	}
}
