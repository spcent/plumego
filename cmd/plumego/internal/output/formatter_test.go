package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestFormatterSuccessJSONUsesCommandResult(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetFormat("json")
	f.SetWriters(&out, nil)

	if err := f.Success("created", map[string]string{"id": "app"}); err != nil {
		t.Fatalf("success: %v", err)
	}

	var result commandResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("decode success result: %v; output: %s", err, out.String())
	}
	if result.Status != "success" || result.Message != "created" || result.Data == nil {
		t.Fatalf("unexpected success result: %+v", result)
	}
}

func TestFormatterErrorJSONUsesCommandResult(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetFormat("json")
	f.SetWriters(&out, nil)

	err := f.Error("failed", 7, map[string]string{"reason": "bad input"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.Code != 7 {
		t.Fatalf("exit code = %d, want 7", exitErr.Code)
	}

	var result commandResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("decode error result: %v; output: %s", err, out.String())
	}
	if result.Status != "error" || result.Message != "failed" || result.ExitCode != 7 || result.Data == nil {
		t.Fatalf("unexpected error result: %+v", result)
	}
}
