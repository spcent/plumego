package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseExportLimit(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "default", raw: "", want: DefaultExportRows},
		{name: "invalid", raw: "abc", want: DefaultExportRows},
		{name: "negative", raw: "-1", want: DefaultExportRows},
		{name: "custom", raw: "250", want: 250},
		{name: "clamped", raw: "999999", want: MaxExportRows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseExportLimit(tt.raw); got != tt.want {
				t.Fatalf("parseExportLimit(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

func TestDecodeJSONLimitedRejectsLargeBody(t *testing.T) {
	body := strings.NewReader(`{"value":"` + strings.Repeat("x", DefaultJSONBodyLimitBytes) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/test", body)
	rec := httptest.NewRecorder()

	var dst map[string]string
	if decodeJSONLimited(rec, req, testLogger{}, &dst) {
		t.Fatal("decodeJSONLimited returned true for oversized body")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
	if !strings.Contains(rec.Body.String(), "payload_too_large") {
		t.Fatalf("body = %s, want payload_too_large", rec.Body.String())
	}
}
