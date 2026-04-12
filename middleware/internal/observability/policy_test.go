package observability

import "testing"

func TestRedactFieldsAcceptsExtraSensitiveKeys(t *testing.T) {
	fields := RedactFields(map[string]any{
		"api_key":  "secret",
		"password": "p4ss",
	}, "api_key")
	if fields["api_key"] != "***" || fields["password"] != "***" {
		t.Fatalf("expected default and extra sensitive keys to be redacted, got %v", fields)
	}
}
