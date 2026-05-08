package main

import (
	"strings"
	"testing"
)

func TestInventoryWarningsRequireDecisionMetadata(t *testing.T) {
	entries := []inventoryEntry{{
		ID:     "missing-metadata",
		Status: "keep",
		Paths:  []string{"example.go"},
	}}

	warnings := inventoryWarnings(entries, []marker{{Path: "example.go", Line: 1, Text: "// compatibility alias"}})
	joined := strings.Join(warnings, "\n")

	for _, want := range []string{
		"missing category",
		"missing owner",
		"missing decision",
		"missing replacement",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("warnings missing %q: %v", want, warnings)
		}
	}
}

func TestInventoryWarningsAcceptCompleteRetainedEntry(t *testing.T) {
	entries := []inventoryEntry{{
		ID:          "complete-entry",
		Category:    "public_alias",
		Status:      "keep",
		Owner:       "transport",
		Decision:    "Keep for v1 as an explicit compatibility alias.",
		Replacement: "canonical.Symbol",
		Paths:       []string{"example.go"},
	}}

	warnings := inventoryWarnings(entries, []marker{{Path: "example.go", Line: 1, Text: "type Old = canonical.Symbol"}})
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}
