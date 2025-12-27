package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewCtxPopulatesFields(t *testing.T) {
	deadline := time.Now().Add(time.Minute)
	baseCtx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/users/123?foo=bar", nil).WithContext(baseCtx)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	ctx := NewCtx(httptest.NewRecorder(), req, map[string]string{"id": "123"})

	if ctx.Params["id"] != "123" {
		t.Fatalf("expected param to be propagated")
	}

	if ctx.Query.Get("foo") != "bar" {
		t.Fatalf("expected query to be parsed")
	}

	if ctx.ClientIP != "1.2.3.4" {
		t.Fatalf("expected client ip from header, got %s", ctx.ClientIP)
	}

	if ctx.Deadline.IsZero() || !ctx.Deadline.Equal(deadline) {
		t.Fatalf("expected deadline to be copied")
	}
}

func TestCtxResponseHelpers(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := NewCtx(recorder, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	if err := ctx.JSON(http.StatusAccepted, map[string]string{"msg": "ok"}); err != nil {
		t.Fatalf("json response failed: %v", err)
	}

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d got %d", http.StatusAccepted, recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected json content type, got %s", ct)
	}

	var payload map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["msg"] != "ok" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestBindJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"demo"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	var payload struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		t.Fatalf("expected successful bind, got %v", err)
	}

	if payload.Name != "demo" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestBindJSONErrors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{name: "empty", body: "", wantMsg: "request body is empty"},
		{name: "invalid", body: "{", wantMsg: "invalid JSON payload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body)), nil)
			var payload map[string]any
			err := ctx.BindJSON(&payload)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}

			var bindErr *BindError
			if !errors.As(err, &bindErr) {
				t.Fatalf("expected BindError, got %T", err)
			}

			if bindErr.Message != tt.wantMsg {
				t.Fatalf("unexpected message: %s", bindErr.Message)
			}
		})
	}
}
