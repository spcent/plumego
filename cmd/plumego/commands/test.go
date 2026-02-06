package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TestCmd struct{}

func (c *TestCmd) Name() string {
	return "test"
}

func (c *TestCmd) Short() string {
	return "Run tests with enhanced reporting"
}

func (c *TestCmd) Long() string {
	return `Run Go tests with enhanced output and reporting.

This command wraps 'go test' with structured output formats and
additional analysis for better CI/CD integration.

Examples:
  plumego test
  plumego test --race
  plumego test --cover
  plumego test --bench
  plumego test --format json
  plumego test ./pkg/...`
}

func (c *TestCmd) Flags() []Flag {
	return []Flag{
		{Name: "dir", Default: ".", Usage: "Project directory"},
		{Name: "race", Default: "false", Usage: "Enable race detector"},
		{Name: "cover", Default: "false", Usage: "Enable coverage analysis"},
		{Name: "bench", Default: "false", Usage: "Run benchmarks"},
		{Name: "timeout", Default: "20s", Usage: "Test timeout"},
		{Name: "tags", Default: "", Usage: "Build tags"},
		{Name: "run", Default: "", Usage: "Run only tests matching pattern"},
		{Name: "short", Default: "false", Usage: "Run short tests only"},
	}
}

func (c *TestCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("dir", ".", "Project directory")
	race := fs.Bool("race", false, "Enable race detector")
	cover := fs.Bool("cover", false, "Enable coverage")
	bench := fs.Bool("bench", false, "Run benchmarks")
	timeout := fs.String("timeout", "20s", "Test timeout")
	tags := fs.String("tags", "", "Build tags")
	runPattern := fs.String("run", "", "Run pattern")
	short := fs.Bool("short", false, "Short tests")

	if err := fs.Parse(args); err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	absDir, err := resolveDir(*dir)
	if err != nil {
		return ctx.Out.Error(err.Error(), 1)
	}

	// Build test arguments
	testArgs := []string{"test"}

	// Determine packages to test
	packages := []string{"./..."}
	if len(fs.Args()) > 0 {
		packages = fs.Args()
	}

	if *race {
		testArgs = append(testArgs, "-race")
	}

	if *cover {
		testArgs = append(testArgs, "-cover", "-coverprofile=coverage.out")
	}

	if *bench {
		testArgs = append(testArgs, "-bench=.")
	}

	if *timeout != "" {
		testArgs = append(testArgs, "-timeout", *timeout)
	}

	if *tags != "" {
		testArgs = append(testArgs, "-tags", *tags)
	}

	if *runPattern != "" {
		testArgs = append(testArgs, "-run", *runPattern)
	}

	if *short {
		testArgs = append(testArgs, "-short")
	}

	// Add JSON output for parsing
	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, packages...)

	startTime := time.Now()

	// Run tests
	cmd := exec.Command("go", testArgs...)
	cmd.Dir = absDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if flagVerbose {
		ctx.Out.Verbose(fmt.Sprintf("Running: go %s", strings.Join(testArgs, " ")))
	}

	testErr := cmd.Run()
	duration := time.Since(startTime)

	// Parse JSON output
	testResult := parseTestOutput(stdout.String())
	testResult["duration_ms"] = duration.Milliseconds()
	testResult["race_detector"] = *race
	testResult["coverage_enabled"] = *cover
	testResult["benchmark"] = *bench

	// Read coverage if available
	if *cover {
		coveragePath := filepath.Join(absDir, "coverage.out")
		if coverage, err := parseCoverage(coveragePath); err == nil {
			testResult["coverage_percent"] = coverage
		}
	}

	if testErr != nil {
		testResult["status"] = "failed"
		return ctx.Out.Error("Tests failed", 1, testResult)
	}

	return ctx.Out.Success("Tests passed", testResult)
}

type testEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

func parseTestOutput(jsonOutput string) map[string]any {
	result := map[string]any{
		"status":  "passed",
		"tests":   0,
		"passed":  0,
		"failed":  0,
		"skipped": 0,
	}

	failures := []map[string]any{}
	lines := strings.Split(jsonOutput, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var event testEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Action {
		case "pass":
			if event.Test != "" {
				result["tests"] = result["tests"].(int) + 1
				result["passed"] = result["passed"].(int) + 1
			}
		case "fail":
			if event.Test != "" {
				result["tests"] = result["tests"].(int) + 1
				result["failed"] = result["failed"].(int) + 1
				failures = append(failures, map[string]any{
					"package": event.Package,
					"test":    event.Test,
				})
			}
		case "skip":
			if event.Test != "" {
				result["tests"] = result["tests"].(int) + 1
				result["skipped"] = result["skipped"].(int) + 1
			}
		}
	}

	if len(failures) > 0 {
		result["failures"] = failures
	}

	return result
}

func parseCoverage(coverageFile string) (float64, error) {
	data, err := os.ReadFile(coverageFile)
	if err != nil {
		return 0, err
	}

	// Simple coverage parsing - look for "total:" line
	// Format: "total:	(statements)	XX.X%"
	re := regexp.MustCompile(`total:.*?\s+([\d.]+)%`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) > 1 {
		var coverage float64
		fmt.Sscanf(matches[1], "%f", &coverage)
		return coverage, nil
	}

	// If no total line, calculate from coverage data
	lines := strings.Split(string(data), "\n")
	if len(lines) <= 1 {
		return 0, fmt.Errorf("no coverage data")
	}

	var totalStatements, coveredStatements int
	for _, line := range lines[1:] { // Skip "mode:" line
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			var stmts, count int
			fmt.Sscanf(parts[1], "%d", &stmts)
			fmt.Sscanf(parts[2], "%d", &count)
			totalStatements += stmts
			if count > 0 {
				coveredStatements += stmts
			}
		}
	}

	if totalStatements == 0 {
		return 0, nil
	}

	return float64(coveredStatements) / float64(totalStatements) * 100, nil
}
