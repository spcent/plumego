package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_GenerateSpecAgainstWithRest(t *testing.T) {
	repoRoot := filepath.Clean("../../..")
	appDir := filepath.Join(repoRoot, "reference", "with-rest")
	stdout, _, err := runCLI(t, []string{
		"generate", "spec",
		"--dir", appDir,
		"--output", os.DevNull,
	}, "")
	if err != nil {
		t.Fatalf("generate spec failed: %v\noutput: %s", err, stdout)
	}
}

func TestCLI_GenerateSpecWritesJSONAndYAML(t *testing.T) {
	repoRoot := filepath.Clean("../../..")
	appDir := filepath.Join(repoRoot, "reference", "with-rest")
	tmpDir := t.TempDir()

	jsonPath := filepath.Join(tmpDir, "openapi.json")
	stdout, _, err := runCLI(t, []string{
		"generate", "spec",
		"--dir", appDir,
		"--output", jsonPath,
		"--format", "json",
	}, "")
	if err != nil {
		t.Fatalf("generate json spec failed: %v\noutput: %s", err, stdout)
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json spec: %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("generated spec is not JSON: %s", data)
	}
	if !strings.Contains(string(data), `"/api/items"`) {
		t.Fatalf("generated spec missing with-rest route: %s", data)
	}

	yamlPath := filepath.Join(tmpDir, "openapi.yaml")
	stdout, _, err = runCLI(t, []string{
		"generate", "spec",
		"--dir", appDir,
		"--output", yamlPath,
		"--format", "yaml",
	}, "")
	if err != nil {
		t.Fatalf("generate yaml spec failed: %v\noutput: %s", err, stdout)
	}
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read yaml spec: %v", err)
	}
	yamlOut := string(yamlData)
	if !strings.Contains(yamlOut, `openapi: "3.1.0"`) || !strings.Contains(yamlOut, "/api/items:") {
		t.Fatalf("generated spec is not expected YAML: %s", yamlOut)
	}
}
