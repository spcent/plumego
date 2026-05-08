package devserver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEnvEntriesUsesSharedDotenvParser(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	content := `
export APP_ADDR=":8080" # local port
JWT_SECRET='quoted secret'
APP_DEBUG=true # inline comment
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	entries, exists, _, err := readEnvEntries(path)
	if err != nil {
		t.Fatalf("read env entries: %v", err)
	}
	if !exists {
		t.Fatal("expected env file to exist")
	}

	got := map[string]string{}
	for _, entry := range entries {
		got[entry.Key] = entry.Value
	}
	if got["APP_ADDR"] != ":8080" || got["JWT_SECRET"] != "quoted secret" || got["APP_DEBUG"] != "true" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}

func TestWriteEnvEntriesQuotesSpecialValues(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	entries := []ConfigEditEntry{
		{Key: "APP_NAME", Value: "demo app"},
		{Key: "APP_NOTE", Value: `value # with "quotes"`},
	}

	if err := writeEnvEntries(path, entries); err != nil {
		t.Fatalf("write env entries: %v", err)
	}
	read, exists, _, err := readEnvEntries(path)
	if err != nil {
		t.Fatalf("read env entries: %v", err)
	}
	if !exists {
		t.Fatal("expected env file to exist")
	}

	got := map[string]string{}
	for _, entry := range read {
		got[entry.Key] = entry.Value
	}
	if got["APP_NAME"] != "demo app" || got["APP_NOTE"] != `value # with "quotes"` {
		t.Fatalf("unexpected round trip entries: %#v", read)
	}
}

func TestNormalizeConfigEntriesRejectsInvalidKeys(t *testing.T) {
	if _, err := normalizeConfigEntries([]ConfigEditEntry{{Key: "1INVALID", Value: "value"}}); err == nil {
		t.Fatal("expected invalid key error")
	}
}

func TestWriteEnvEntriesRejectsUnsupportedValues(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	err := writeEnvEntries(path, []ConfigEditEntry{{Key: "APP_VALUE", Value: "line1\nline2"}})
	if err == nil {
		t.Fatal("expected newline value to fail")
	}
}
