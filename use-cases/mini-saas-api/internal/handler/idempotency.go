package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
	"github.com/spcent/plumego/store/idempotency"
)

// IdempotencyHeader is the client-provided dedup key for mutating requests.
const IdempotencyHeader = "Idempotency-Key"

// replayHeader marks responses served from the idempotency store.
const replayHeader = "X-Idempotency-Replay"

// storedResponse is the replayable response payload kept in the record.
type storedResponse struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

// Idempotent wraps mutating handlers with Idempotency-Key replay protection
// backed by the stable store/idempotency contract.
//
// Requests without the header pass through untouched. With the header, the
// first request runs the handler and stores its response; duplicates with the
// same key and payload replay the stored response; the same key with a
// different payload is a 422 mismatch; a concurrent duplicate while the first
// request is still running is a 409. Keys are scoped per tenant + method +
// path so tenants can never collide. Must run inside RequireAuth.
func Idempotent(store idempotency.Store, ttl time.Duration, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := r.Header.Get(IdempotencyHeader)
			if clientKey == "" {
				next.ServeHTTP(w, r)
				return
			}
			p := authn.PrincipalFromContext(r.Context())
			if p == nil {
				writeUnauthorized(w, r, logger, "auth.missing_principal", "authentication required")
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeBadJSON(w, r, logger)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			key := p.TenantID + "|" + r.Method + "|" + r.URL.Path + "|" + clientKey
			hash := requestHash(r.Method, r.URL.Path, body)
			now := time.Now().UTC()

			if rec, found, err := store.Get(r.Context(), key); err != nil {
				writeIdemError(w, r, logger, err)
				return
			} else if found {
				serveExisting(w, r, logger, rec, hash)
				return
			}

			inserted, err := store.PutIfAbsent(r.Context(), idempotency.Record{
				Key:         key,
				RequestHash: hash,
				Status:      idempotency.StatusInProgress,
				CreatedAt:   now,
				UpdatedAt:   now,
				ExpiresAt:   now.Add(ttl),
			})
			if err != nil {
				writeIdemError(w, r, logger, err)
				return
			}
			if !inserted {
				// Lost the claim race; the concurrent owner decides the outcome.
				if rec, found, err := store.Get(r.Context(), key); err == nil && found {
					serveExisting(w, r, logger, rec, hash)
					return
				}
				writeInFlight(w, r, logger)
				return
			}

			// Run the handler against a buffer so the response can be stored
			// before it is sent.
			capture := &captureWriter{header: make(http.Header)}
			next.ServeHTTP(capture, r)

			if capture.status >= http.StatusInternalServerError {
				// Server-side failure: release the key so the client can retry.
				if err := store.Delete(r.Context(), key); err != nil {
					logWriteErr(logger, err)
				}
			} else {
				payload, err := json.Marshal(storedResponse{
					Status:      capture.statusOrOK(),
					ContentType: capture.header.Get("Content-Type"),
					Body:        capture.body.Bytes(),
				})
				if err == nil {
					err = completeRecord(r, store, key, hash, payload)
				}
				if err != nil {
					logWriteErr(logger, err)
				}
			}
			capture.flushTo(w)
		})
	}
}

// completeRecord prefers hash-aware completion when the backend supports it.
func completeRecord(r *http.Request, store idempotency.Store, key, hash string, payload []byte) error {
	if hs, ok := store.(idempotency.HashAwareStore); ok {
		return hs.CompleteWithRequestHash(r.Context(), key, hash, payload)
	}
	return store.Complete(r.Context(), key, payload)
}

// serveExisting handles a found record: replay, mismatch, or in-flight.
func serveExisting(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger, rec idempotency.Record, hash string) {
	if rec.RequestHash != hash {
		logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code("idempotency.payload_mismatch").
			Message("this Idempotency-Key was already used with a different request payload").
			Build()))
		return
	}
	if rec.Status != idempotency.StatusCompleted {
		writeInFlight(w, r, logger)
		return
	}
	var stored storedResponse
	if err := json.Unmarshal(rec.Response, &stored); err != nil {
		writeIdemError(w, r, logger, err)
		return
	}
	if stored.ContentType != "" {
		w.Header().Set("Content-Type", stored.ContentType)
	}
	w.Header().Set(replayHeader, "true")
	w.WriteHeader(stored.Status)
	if len(stored.Body) > 0 {
		if _, err := w.Write(stored.Body); err != nil {
			logWriteErr(logger, err)
		}
	}
}

func writeInFlight(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger) {
	logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeConflict).
		Code("idempotency.in_flight").
		Message("a request with this Idempotency-Key is still being processed").
		Build()))
}

func writeIdemError(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger, err error) {
	if logger != nil {
		logger.Error("idempotency store error", plumelog.Fields{"error": err.Error()})
	}
	logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).
		Code("idempotency.store_error").
		Message("internal error").
		Build()))
}

func requestHash(method, path string, body []byte) string {
	sum := sha256.New()
	sum.Write([]byte(method))
	sum.Write([]byte{0})
	sum.Write([]byte(path))
	sum.Write([]byte{0})
	sum.Write(body)
	return hex.EncodeToString(sum.Sum(nil))
}

// captureWriter buffers a handler response for storage before sending.
type captureWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (c *captureWriter) Header() http.Header { return c.header }

func (c *captureWriter) WriteHeader(status int) {
	if c.status == 0 {
		c.status = status
	}
}

func (c *captureWriter) Write(b []byte) (int, error) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	return c.body.Write(b)
}

func (c *captureWriter) statusOrOK() int {
	if c.status == 0 {
		return http.StatusOK
	}
	return c.status
}

func (c *captureWriter) flushTo(w http.ResponseWriter) {
	for k, vals := range c.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(c.statusOrOK())
	if c.body.Len() > 0 {
		_, _ = w.Write(c.body.Bytes())
	}
}
