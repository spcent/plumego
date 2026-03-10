package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type snippet struct {
	File      string
	StartLine int
	Code      string
}

var defaultTargets = []string{
	"README.md",
	"README_CN.md",
	"docs/getting-started.md",
	"docs/modules/core/README.md",
	"docs/modules/health/README.md",
	"docs/modules/health/readiness.md",
	"docs/modules/router/path-parameters.md",
	"docs/modules/router/advanced-patterns.md",
	"examples/docs/en/guide.md",
	"examples/docs/zh/guide.md",
	"examples/docs/en/websocket/websocket.md",
}

func main() {
	targets := defaultTargets
	if len(os.Args) > 1 {
		targets = os.Args[1:]
	}

	root, err := os.Getwd()
	if err != nil {
		exitf("get working directory: %v", err)
	}

	snippets := make([]snippet, 0)
	for _, target := range targets {
		if _, err := os.Stat(target); err != nil {
			exitf("target %q is not readable: %v", target, err)
		}
		extracted, err := extractMainSnippets(target)
		if err != nil {
			exitf("extract snippets from %q: %v", target, err)
		}
		snippets = append(snippets, extracted...)
	}

	if len(snippets) == 0 {
		exitf("no `package main` Go snippets found in targets")
	}

	goVersion, err := detectGoVersion(filepath.Join(root, "go.mod"))
	if err != nil {
		exitf("detect go version: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "plumego-doc-snippets-*")
	if err != nil {
		exitf("create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := writeTempModule(tmpDir, root, goVersion); err != nil {
		exitf("prepare temp module: %v", err)
	}

	caseMap := make([]string, 0, len(snippets))
	for i, snip := range snippets {
		caseName := fmt.Sprintf("case%03d", i+1)
		caseDir := filepath.Join(tmpDir, caseName)
		if err := os.MkdirAll(caseDir, 0o755); err != nil {
			exitf("create %s: %v", caseName, err)
		}
		mainPath := filepath.Join(caseDir, "main.go")
		if err := os.WriteFile(mainPath, []byte(snip.Code), 0o644); err != nil {
			exitf("write %s: %v", mainPath, err)
		}
		caseMap = append(caseMap, fmt.Sprintf("%s => %s:%d", caseName, snip.File, snip.StartLine))
	}

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "doc-snippet-compile: failed\n")
		fmt.Fprintf(os.Stderr, "snippet case map:\n")
		for _, item := range caseMap {
			fmt.Fprintf(os.Stderr, "  %s\n", item)
		}
		fmt.Fprintf(os.Stderr, "\n%s", out)
		os.Exit(1)
	}

	fmt.Printf("doc-snippet-compile: ok (%d snippets)\n", len(snippets))
	_ = out
}

func extractMainSnippets(path string) ([]snippet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		lineNo       int
		inFence      bool
		startLine    int
		currentLines []string
		result       []snippet
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inFence {
			if strings.HasPrefix(trimmed, "```go") {
				inFence = true
				startLine = lineNo + 1
				currentLines = currentLines[:0]
			}
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			code := strings.Join(currentLines, "\n")
			if strings.Contains(code, "package main") {
				result = append(result, snippet{
					File:      path,
					StartLine: startLine,
					Code:      code + "\n",
				})
			}
			inFence = false
			continue
		}

		currentLines = append(currentLines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inFence {
		return nil, errors.New("unterminated ```go block")
	}
	return result, nil
}

func detectGoVersion(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("go directive not found in %s", goModPath)
}

func writeTempModule(tempDir, root, goVersion string) error {
	mod := strings.Builder{}
	mod.WriteString("module plumego_doc_snippet_check\n\n")
	mod.WriteString("go " + goVersion + "\n\n")
	mod.WriteString("require github.com/spcent/plumego v0.0.0\n")
	mod.WriteString("replace github.com/spcent/plumego => " + root + "\n")
	return os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(mod.String()), 0o644)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
