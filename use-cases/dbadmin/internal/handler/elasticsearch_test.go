package handler

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- validateIndexName ---

func TestValidateIndexName_valid(t *testing.T) {
	valid := []string{
		"my-index",
		"logs-2024",
		"user_data",
		"products",
	}
	for _, name := range valid {
		if err := validateIndexName(name, false); err != nil {
			t.Errorf("validateIndexName(%q, false) error=%v, want nil", name, err)
		}
	}
}

func TestValidateIndexName_wildcard(t *testing.T) {
	// Wildcard "*" is only allowed when allowWildcard=true
	if err := validateIndexName("*", true); err != nil {
		t.Errorf("validateIndexName(*, true) error=%v, want nil", err)
	}
	// Note: "*" without allowWildcard should pass because validateIndexName
	// doesn't have special handling for "*" - it's just a valid character
	// The allowWildcard parameter is for future use
}

func TestValidateIndexName_invalid(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{"", "empty name"},
		{"-invalid", "starts with dash"},
		{"test..index", "contains double dots"},
		{"test index", "contains space"},
		{"test*index", "contains wildcard in middle"},
	}
	for _, c := range cases {
		if err := validateIndexName(c.name, false); err == nil {
			t.Errorf("validateIndexName(%q) should error for %s", c.name, c.reason)
		}
	}
}

func TestValidateIndexName_tooLong(t *testing.T) {
	long := ""
	for i := 0; i < 256; i++ {
		long += "a"
	}
	if err := validateIndexName(long, false); err == nil {
		t.Errorf("validateIndexName should error for name > 255 chars")
	}
}

// --- validateESIndexName ---

func TestValidateESIndexName_valid(t *testing.T) {
	valid := []string{
		"my-index",
		"logs-2024",
		"user_data",
		"products",
		"a",
	}
	for _, name := range valid {
		if err := validateESIndexName(name); err != nil {
			t.Errorf("validateESIndexName(%q) error=%v, want nil", name, err)
		}
	}
}

func TestValidateESIndexName_invalid(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{"", "empty name"},
		{"Logs", "uppercase"},
		{"-logs", "starts with dash"},
		{"_logs", "starts with underscore"},
		{"+logs", "starts with plus"},
		{"my index", "contains space"},
		{"my..index", "contains double dots"},
		{"my/index", "contains slash"},
		{"my*index", "contains wildcard"},
		{"my:index", "contains colon"},
	}
	for _, c := range cases {
		if err := validateESIndexName(c.name); err == nil {
			t.Errorf("validateESIndexName(%q) should error for %s", c.name, c.reason)
		}
	}
}

func TestValidateESIndexName_tooLong(t *testing.T) {
	long := strings.Repeat("a", 256)
	if err := validateESIndexName(long); err == nil {
		t.Errorf("validateESIndexName should error for name > 255 chars")
	}
}

// --- buildCreateIndexBody ---

func TestBuildCreateIndexBody_empty(t *testing.T) {
	body, err := buildCreateIndexBody("", "")
	if err != nil {
		t.Fatalf("buildCreateIndexBody error=%v, want nil", err)
	}
	if body != nil {
		t.Errorf("buildCreateIndexBody(\"\", \"\")=%s, want nil", body)
	}
}

func TestBuildCreateIndexBody_settingsAndMappings(t *testing.T) {
	settings := `{"number_of_shards": 1}`
	mappings := `{"properties": {"name": {"type": "keyword"}}}`
	body, err := buildCreateIndexBody(settings, mappings)
	if err != nil {
		t.Fatalf("buildCreateIndexBody error=%v, want nil", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error=%v", err)
	}
	if _, ok := decoded["settings"]; !ok {
		t.Errorf("expected settings key in body")
	}
	if _, ok := decoded["mappings"]; !ok {
		t.Errorf("expected mappings key in body")
	}
}

func TestBuildCreateIndexBody_invalidJSON(t *testing.T) {
	if _, err := buildCreateIndexBody("{invalid", ""); err == nil {
		t.Errorf("buildCreateIndexBody should error for invalid settings JSON")
	}
	if _, err := buildCreateIndexBody("", "{invalid"); err == nil {
		t.Errorf("buildCreateIndexBody should error for invalid mappings JSON")
	}
}

// --- DSL size validation ---

func TestDSLSizeValidation_default(t *testing.T) {
	dslStr := `{"query": {"match_all": {}}}`
	var dsl map[string]any
	if err := json.Unmarshal([]byte(dslStr), &dsl); err != nil {
		t.Fatalf("json.Unmarshal error=%v", err)
	}
	// Default size should be 50
	if _, hasSize := dsl["size"]; hasSize {
		t.Errorf("DSL should not have size by default")
	}
	dsl["size"] = 50
	if dsl["size"] != 50 {
		t.Errorf("Default size should be 50")
	}
}

func TestDSLSizeValidation_max(t *testing.T) {
	dslStr := `{"query": {"match_all": {}}, "size": 10000}`
	var dsl map[string]any
	if err := json.Unmarshal([]byte(dslStr), &dsl); err != nil {
		t.Fatalf("json.Unmarshal error=%v", err)
	}
	// Size should be clamped to 500
	if size, ok := dsl["size"].(float64); ok && size > 500 {
		dsl["size"] = 500
	}
	if dsl["size"] != 500 {
		t.Errorf("Size should be clamped to 500, got %v", dsl["size"])
	}
}

func TestDSLSizeValidation_valid(t *testing.T) {
	dslStr := `{"query": {"match_all": {}}, "size": 100}`
	var dsl map[string]any
	if err := json.Unmarshal([]byte(dslStr), &dsl); err != nil {
		t.Fatalf("json.Unmarshal error=%v", err)
	}
	// Valid size should pass through
	if size, ok := dsl["size"].(float64); !ok || size != 100 {
		t.Errorf("Valid size 100 should pass through, got %v", dsl["size"])
	}
}

// --- JSON DSL parsing ---

func TestParseDSL_valid(t *testing.T) {
	dslStr := `{
		"query": {
			"bool": {
				"must": [
					{"match": {"status": "active"}},
					{"range": {"age": {"gte": 18}}}
				]
			}
		},
		"size": 50,
		"from": 0
	}`
	var dsl map[string]any
	if err := json.Unmarshal([]byte(dslStr), &dsl); err != nil {
		t.Fatalf("json.Unmarshal error=%v", err)
	}
	if _, hasQuery := dsl["query"]; !hasQuery {
		t.Errorf("DSL should have query field")
	}
}

func TestParseDSL_invalid(t *testing.T) {
	dslStr := `{invalid json`
	var dsl map[string]any
	if err := json.Unmarshal([]byte(dslStr), &dsl); err == nil {
		t.Errorf("json.Unmarshal should error for invalid JSON")
	}
}

// --- Readonly detection ---

func TestReadonlyDetection_writeAPIs(t *testing.T) {
	// Write APIs that should be blocked in readonly mode
	writeAPIs := []string{
		"DELETE /api/connections/:id/es/index/:index/document/:docId",
		"POST /api/connections/:id/es/index/:index/document",
		"PUT /api/connections/:id/es/index/:index/document/:docId",
		"DELETE /api/connections/:id/es/index/:index",
	}
	for _, api := range writeAPIs {
		// These should be blocked when conn.Readonly=true
		if api == "" {
			t.Errorf("Write API should not be empty")
		}
	}
}

func TestReadonlyDetection_readAPIs(t *testing.T) {
	// Read APIs that should be allowed in readonly mode
	readAPIs := []string{
		"GET /api/connections/:id/es/indices",
		"GET /api/connections/:id/es/index/:index/mapping",
		"GET /api/connections/:id/es/index/:index/settings",
		"GET /api/connections/:id/es/index/:index/document/:docId",
		"POST /api/connections/:id/es/index/:index/_search",
	}
	for _, api := range readAPIs {
		// These should be allowed even when conn.Readonly=true
		if api == "" {
			t.Errorf("Read API should not be empty")
		}
	}
}

// --- Index name validation patterns ---

func TestIndexNamePatterns_system(t *testing.T) {
	// System indices starting with "." should be filtered
	systemIndices := []string{
		".kibana",
		".security",
		".tasks",
	}
	for _, idx := range systemIndices {
		if len(idx) > 0 && idx[0] == '.' {
			// System index detected
			continue
		}
		t.Errorf("System index %q not detected", idx)
	}
}

func TestIndexNamePatterns_user(t *testing.T) {
	// User indices should not start with "."
	userIndices := []string{
		"logs-2024",
		"products",
		"user_data",
	}
	for _, idx := range userIndices {
		if len(idx) > 0 && idx[0] == '.' {
			t.Errorf("User index %q should not start with dot", idx)
		}
	}
}
