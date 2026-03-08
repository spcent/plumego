package bind

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
)

type testPayload struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required" mask:"true"`
}

type sensitivePayload struct {
	Name      string `json:"name" validate:"required"`
	Password  string `json:"password"`
	Token     string `json:"token"`
	Secret    string `json:"secret"`
	Signature string `json:"signature"`
}

func TestBindJSONSuccess(t *testing.T) {
	body := []byte(`{"name":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler := BindJSON[testPayload](JSONOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p testPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("expected request body to remain readable: %v", err)
		}
		if p.Name != "alice" {
			t.Fatalf("unexpected name: %s", p.Name)
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestBindJSONValidationError(t *testing.T) {
	body := []byte(`{"name":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler := BindJSON[testPayload](JSONOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Category != contract.CategoryValidation {
		t.Fatalf("expected validation category, got %s", resp.Error.Category)
	}
	fieldsRaw, ok := resp.Error.Details["fields"]
	if !ok {
		t.Fatalf("expected field errors in details")
	}
	fields, ok := fieldsRaw.([]any)
	if !ok || len(fields) == 0 {
		t.Fatalf("expected non-empty field errors, got %T", fieldsRaw)
	}
}

func TestRedactor(t *testing.T) {
	redactor := DefaultRedactor()
	payload := testPayload{Name: "alice", Password: "secret"}
	result := redactor.Redact(payload)

	data, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if data["password"] == "secret" {
		t.Fatalf("expected password to be redacted")
	}
}

type bindTestLogger struct {
	last log.Fields
}

func (l *bindTestLogger) WithFields(fields log.Fields) log.StructuredLogger { return l }
func (l *bindTestLogger) With(key string, value any) log.StructuredLogger   { return l }
func (l *bindTestLogger) Debug(msg string, fields ...log.Fields)            {}
func (l *bindTestLogger) Info(msg string, fields ...log.Fields)             {}
func (l *bindTestLogger) Warn(msg string, fields ...log.Fields)             {}
func (l *bindTestLogger) Error(msg string, fields ...log.Fields)            {}
func (l *bindTestLogger) Fatal(msg string, fields ...log.Fields)            {}
func (l *bindTestLogger) DebugCtx(ctx context.Context, msg string, fields ...log.Fields) {
}
func (l *bindTestLogger) InfoCtx(ctx context.Context, msg string, fields ...log.Fields) {}
func (l *bindTestLogger) WarnCtx(ctx context.Context, msg string, fields ...log.Fields) {
	if len(fields) > 0 {
		l.last = fields[0]
	}
}
func (l *bindTestLogger) ErrorCtx(ctx context.Context, msg string, fields ...log.Fields) {}
func (l *bindTestLogger) FatalCtx(ctx context.Context, msg string, fields ...log.Fields) {}

func TestBindJSONMasksSensitiveFieldsInLogs(t *testing.T) {
	logger := &bindTestLogger{}
	body := []byte(`{"name":"","password":"my-password","token":"abc","secret":"def","signature":"ghi"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler := BindJSON[sensitivePayload](JSONOptions{Logger: logger})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	payload, ok := logger.last["payload"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload map in logs, got %T", logger.last["payload"])
	}
	for _, key := range []string{"token", "secret", "signature", "password"} {
		if payload[key] != "***" {
			t.Fatalf("expected %s to be masked, got %v", key, payload[key])
		}
	}
}
