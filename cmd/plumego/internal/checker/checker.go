package checker

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckResult represents the overall health check result
type CheckResult struct {
	Status string                  `json:"status" yaml:"status"` // healthy, degraded, unhealthy
	Checks map[string]CheckDetail  `json:"checks" yaml:"checks"`
}

// CheckDetail represents a single check result
type CheckDetail struct {
	Status   string        `json:"status" yaml:"status"`     // passed, warning, failed
	Issues   []CheckIssue  `json:"issues" yaml:"issues"`
	Outdated []string      `json:"outdated,omitempty" yaml:"outdated,omitempty"`
}

// CheckIssue represents a specific issue found
type CheckIssue struct {
	Severity string `json:"severity" yaml:"severity"` // low, medium, high, critical
	Message  string `json:"message" yaml:"message"`
	Fix      string `json:"fix,omitempty" yaml:"fix,omitempty"`
}

// CheckConfig validates configuration files and environment
func CheckConfig(dir, envFile string) CheckDetail {
	detail := CheckDetail{
		Status: "passed",
		Issues: []CheckIssue{},
	}

	// Check if go.mod exists
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		detail.Status = "failed"
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "high",
			Message:  "go.mod not found",
			Fix:      "Run 'go mod init <module-path>' to initialize Go module",
		})
	}

	// Check if env.example exists
	envExamplePath := filepath.Join(dir, "env.example")
	if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
		detail.Status = "warning"
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "low",
			Message:  "env.example not found",
			Fix:      "Create env.example file to document environment variables",
		})
	}

	// Check if .env exists (optional)
	envPath := filepath.Join(dir, envFile)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "low",
			Message:  fmt.Sprintf("%s not found (optional)", envFile),
			Fix:      fmt.Sprintf("Copy env.example to %s and configure", envFile),
		})
	}

	return detail
}

// CheckDependencies validates Go dependencies
func CheckDependencies(dir string) CheckDetail {
	detail := CheckDetail{
		Status: "passed",
		Issues: []CheckIssue{},
	}

	// Check if go.mod exists
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		detail.Status = "failed"
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "high",
			Message:  "go.mod not found",
			Fix:      "Run 'go mod init <module-path>' to initialize Go module",
		})
		return detail
	}

	// Run go mod verify
	cmd := exec.Command("go", "mod", "verify")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail.Status = "failed"
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "high",
			Message:  fmt.Sprintf("go mod verify failed: %s", strings.TrimSpace(string(output))),
			Fix:      "Run 'go mod tidy' to fix dependencies",
		})
	}

	// Check for outdated dependencies (list packages that could be updated)
	cmd = exec.Command("go", "list", "-u", "-m", "all")
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		outdated := []string{}
		for _, line := range lines {
			if strings.Contains(line, "[") && strings.Contains(line, "]") {
				// Format: module version [available]
				outdated = append(outdated, strings.TrimSpace(line))
			}
		}
		if len(outdated) > 0 {
			detail.Status = "warning"
			detail.Outdated = outdated
			detail.Issues = append(detail.Issues, CheckIssue{
				Severity: "low",
				Message:  fmt.Sprintf("%d dependencies have updates available", len(outdated)),
				Fix:      "Run 'go get -u <module>' to update specific modules",
			})
		}
	}

	return detail
}

// CheckSecurity runs security checks
func CheckSecurity(dir, envFile string) CheckDetail {
	detail := CheckDetail{
		Status: "passed",
		Issues: []CheckIssue{},
	}

	// Check for sensitive environment variables
	requiredSecrets := []string{"WS_SECRET", "JWT_SECRET"}

	envPath := filepath.Join(dir, envFile)
	envVars := make(map[string]bool)

	// Try to read .env file
	if file, err := os.Open(envPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = true
			}
		}
	}

	// Check environment variables
	for _, secret := range requiredSecrets {
		hasEnvVar := envVars[secret]
		hasSystemEnv := os.Getenv(secret) != ""

		if !hasEnvVar && !hasSystemEnv {
			detail.Status = "warning"
			detail.Issues = append(detail.Issues, CheckIssue{
				Severity: "medium",
				Message:  fmt.Sprintf("%s not set in environment or %s", secret, envFile),
				Fix:      fmt.Sprintf("Set %s to a secure random string (32+ bytes)", secret),
			})
		}
	}

	// Check for .env in git
	gitignorePath := filepath.Join(dir, ".gitignore")
	if file, err := os.Open(gitignorePath); err == nil {
		defer file.Close()
		hasEnvIgnore := false
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == ".env" || line == ".env*" || strings.HasPrefix(line, ".env") {
				hasEnvIgnore = true
				break
			}
		}

		if !hasEnvIgnore {
			detail.Status = "warning"
			detail.Issues = append(detail.Issues, CheckIssue{
				Severity: "high",
				Message:  ".env files not in .gitignore",
				Fix:      "Add '.env' to .gitignore to prevent committing secrets",
			})
		}
	}

	// Check for common security issues in code
	mainGoPath := filepath.Join(dir, "main.go")
	if file, err := os.Open(mainGoPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Check for hardcoded secrets (basic check)
			if strings.Contains(line, `secret := "`) ||
			   strings.Contains(line, `password := "`) ||
			   strings.Contains(line, `key := "`) {
				detail.Status = "warning"
				detail.Issues = append(detail.Issues, CheckIssue{
					Severity: "high",
					Message:  fmt.Sprintf("Possible hardcoded secret in main.go:%d", lineNum),
					Fix:      "Use environment variables for secrets",
				})
			}
		}
	}

	return detail
}

// CheckStructure validates project structure
func CheckStructure(dir string) CheckDetail {
	detail := CheckDetail{
		Status: "passed",
		Issues: []CheckIssue{},
	}

	// Check for main.go
	mainGoPath := filepath.Join(dir, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		detail.Status = "warning"
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "medium",
			Message:  "main.go not found in project root",
			Fix:      "Create main.go or ensure it exists in the correct location",
		})
	}

	// Check for README
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "low",
			Message:  "README.md not found",
			Fix:      "Create README.md to document the project",
		})
	}

	// Check for .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		detail.Status = "warning"
		detail.Issues = append(detail.Issues, CheckIssue{
			Severity: "medium",
			Message:  ".gitignore not found",
			Fix:      "Create .gitignore to exclude build artifacts and secrets",
		})
	}

	return detail
}
