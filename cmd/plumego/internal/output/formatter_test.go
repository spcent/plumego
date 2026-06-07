package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestFormatterDefaultsToJSON(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetWriters(&out, nil)

	if err := f.Success("created", map[string]string{"id": "app"}); err != nil {
		t.Fatalf("success: %v", err)
	}

	var result commandResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("default formatter output should be json: %v; output: %s", err, out.String())
	}
	if result.Status != "success" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestIsSupportedFormat(t *testing.T) {
	for _, format := range []string{"json", "yaml", "text"} {
		if !IsSupportedFormat(format) {
			t.Fatalf("expected %q to be supported", format)
		}
	}
	if IsSupportedFormat("bogus") {
		t.Fatalf("expected bogus format to be unsupported")
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

func TestFormatterWarningJSONUsesCommandResult(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetFormat("json")
	f.SetWriters(&out, nil)

	err := f.Warning("check warnings", 2, map[string]string{"field": "config"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.Code != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.Code)
	}

	var result commandResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("decode warning result: %v; output: %s", err, out.String())
	}
	if result.Status != "warning" || result.Message != "check warnings" || result.ExitCode != 2 || result.Data == nil {
		t.Fatalf("unexpected warning result: %+v", result)
	}
}

func TestFormatterPrintStringJSONUsesCommandResult(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetFormat("json")
	f.SetWriters(&out, nil)

	if err := f.Print("plain output"); err != nil {
		t.Fatalf("print string: %v", err)
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("decode string result: %v; output: %s", err, out.String())
	}
	if result.Status != "success" || result.Message != "output" || result.Data.Value != "plain output" {
		t.Fatalf("unexpected string result: %+v", result)
	}
}

func TestFormatterPrintStringYAMLUsesCommandResult(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetFormat("yaml")
	f.SetWriters(&out, nil)

	if err := f.Print("plain output"); err != nil {
		t.Fatalf("print string: %v", err)
	}

	text := out.String()
	if !strings.Contains(text, "status: success") ||
		!strings.Contains(text, "message: output") ||
		!strings.Contains(text, "value: plain output") {
		t.Fatalf("expected YAML command result, got: %s", text)
	}
}

func TestFormatterTextCommandResultAvoidsGoStructDump(t *testing.T) {
	var out bytes.Buffer
	f := NewFormatter()
	f.SetFormat("text")
	f.SetWriters(&out, nil)

	if err := f.Success("created", map[string]string{"id": "app"}); err != nil {
		t.Fatalf("success: %v", err)
	}

	text := out.String()
	if !strings.Contains(text, "SUCCESS: created") {
		t.Fatalf("expected text status line, got: %s", text)
	}
	if strings.Contains(text, "{success") || strings.Contains(text, "map[") {
		t.Fatalf("text output should not expose Go formatting: %s", text)
	}
	if !strings.Contains(text, "id: app") {
		t.Fatalf("expected YAML data block, got: %s", text)
	}
}

func TestFormatterCommandResultContracts(t *testing.T) {
	tests := []struct {
		name   string
		format string
		assert func(t *testing.T, out string)
	}{
		{
			name:   "json",
			format: "json",
			assert: func(t *testing.T, out string) {
				t.Helper()
				var result commandResult
				if err := json.Unmarshal([]byte(out), &result); err != nil {
					t.Fatalf("decode json result: %v; output: %s", err, out)
				}
				if result.Status != "success" || result.Message != "contract ok" || result.Data == nil {
					t.Fatalf("unexpected json result: %+v", result)
				}
			},
		},
		{
			name:   "yaml",
			format: "yaml",
			assert: func(t *testing.T, out string) {
				t.Helper()
				var result commandResult
				if err := yaml.Unmarshal([]byte(out), &result); err != nil {
					t.Fatalf("decode yaml result: %v; output: %s", err, out)
				}
				if result.Status != "success" || result.Message != "contract ok" || result.Data == nil {
					t.Fatalf("unexpected yaml result: %+v", result)
				}
			},
		},
		{
			name:   "text",
			format: "text",
			assert: func(t *testing.T, out string) {
				t.Helper()
				if !strings.Contains(out, "SUCCESS: contract ok") || !strings.Contains(out, "name: plumego") {
					t.Fatalf("unexpected text result: %s", out)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			f := NewFormatter()
			f.SetFormat(tt.format)
			f.SetWriters(&out, nil)

			if err := f.Success("contract ok", map[string]string{"name": "plumego"}); err != nil {
				t.Fatalf("success: %v", err)
			}

			tt.assert(t, out.String())
		})
	}
}

func TestFormatterEventContracts(t *testing.T) {
	event := Event{
		Event:   "dev.build.started",
		Level:   "info",
		Message: "Build started",
		Time:    "2026-05-06T12:00:00Z",
		Data: map[string]any{
			"target": "app",
		},
	}

	tests := []struct {
		name   string
		format string
		assert func(t *testing.T, out string)
	}{
		{
			name:   "json",
			format: "json",
			assert: func(t *testing.T, out string) {
				t.Helper()
				var got Event
				if err := json.Unmarshal([]byte(out), &got); err != nil {
					t.Fatalf("decode json event: %v; output: %s", err, out)
				}
				if got.Event != event.Event || got.Message != event.Message || got.Data["target"] != "app" {
					t.Fatalf("unexpected json event: %+v", got)
				}
				if strings.Contains(out, `"status"`) || strings.Contains(out, `"exit_code"`) {
					t.Fatalf("event output must not use command result envelope: %s", out)
				}
			},
		},
		{
			name:   "yaml",
			format: "yaml",
			assert: func(t *testing.T, out string) {
				t.Helper()
				var got Event
				if err := yaml.Unmarshal([]byte(out), &got); err != nil {
					t.Fatalf("decode yaml event: %v; output: %s", err, out)
				}
				if got.Event != event.Event || got.Message != event.Message || got.Data["target"] != "app" {
					t.Fatalf("unexpected yaml event: %+v", got)
				}
				if strings.Contains(out, "status:") || strings.Contains(out, "exit_code:") {
					t.Fatalf("event output must not use command result envelope: %s", out)
				}
			},
		},
		{
			name:   "text",
			format: "text",
			assert: func(t *testing.T, out string) {
				t.Helper()
				if strings.TrimSpace(out) != "Build started" {
					t.Fatalf("unexpected text event: %q", out)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			f := NewFormatter()
			f.SetFormat(tt.format)
			f.SetWriters(&out, nil)

			if err := f.Event(event); err != nil {
				t.Fatalf("event: %v", err)
			}

			tt.assert(t, out.String())
		})
	}
}
