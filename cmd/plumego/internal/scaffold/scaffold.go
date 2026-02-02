package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GetTemplateFiles returns the files that would be created for a template
func GetTemplateFiles(template string) []string {
	base := []string{
		"main.go",
		"go.mod",
		"env.example",
		".gitignore",
		"README.md",
	}

	switch template {
	case "minimal":
		return base
	case "api":
		return append(base, []string{
			"handlers/health.go",
			"handlers/user.go",
			"middleware/auth.go",
			"config/config.go",
		}...)
	case "fullstack":
		return append(base, []string{
			"handlers/api.go",
			"frontend/index.html",
			"frontend/app.js",
			"frontend/styles.css",
		}...)
	case "microservice":
		return append(base, []string{
			"handlers/health.go",
			"handlers/metrics.go",
			"middleware/logging.go",
			"middleware/auth.go",
			"config/config.go",
			"store/db.go",
			"scheduler/jobs.go",
			"Dockerfile",
			"docker-compose.yml",
		}...)
	default:
		return base
	}
}

// CreateProject creates a new project from a template
func CreateProject(dir, name, module, template string, initGit bool) ([]string, error) {
	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	files := GetTemplateFiles(template)
	created := []string{}

	// Create each file
	for _, file := range files {
		path := filepath.Join(dir, file)

		// Create parent directory if needed
		parent := filepath.Dir(path)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return created, fmt.Errorf("failed to create directory %s: %w", parent, err)
		}

		// Get template content
		content := getTemplateContent(file, name, module, template)

		// Write file
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("failed to write file %s: %w", path, err)
		}

		created = append(created, file)
	}

	// Initialize git if requested
	if initGit {
		cmd := exec.Command("git", "init")
		cmd.Dir = dir
		if err := cmd.Run(); err == nil {
			created = append(created, ".git/")
		}
	}

	// Initialize go module
	cmd := exec.Command("go", "mod", "init", module)
	cmd.Dir = dir
	if err := cmd.Run(); err == nil {
		// go.mod already in created list

		// Run go mod tidy
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = dir
		_ = tidyCmd.Run() // Ignore error
	}

	return created, nil
}

func getTemplateContent(file, name, module, template string) string {
	switch file {
	case "main.go":
		return getMainGoContent(module, template)
	case "go.mod":
		return fmt.Sprintf("module %s\n\ngo 1.24\n", module)
	case "env.example":
		return getEnvExampleContent()
	case ".gitignore":
		return getGitignoreContent()
	case "README.md":
		return getReadmeContent(name, template)
	default:
		return fmt.Sprintf("// TODO: Implement %s\npackage main\n", file)
	}
}

func getMainGoContent(module, template string) string {
	switch template {
	case "minimal":
		return fmt.Sprintf(`package main

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
)

func main() {
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithRecovery(),
		core.WithLogging(),
	)

	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from Plumego!"))
	})

	app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	if err := app.Boot(); err != nil {
		log.Fatal(err)
	}
}
`)
	default:
		return getMainGoContent(module, "minimal")
	}
}

func getEnvExampleContent() string {
	return `APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000
APP_MAX_BODY_BYTES=10485760
`
}

func getGitignoreContent() string {
	return `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test coverage
*.out
coverage.html

# Environment
.env
.env.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
}

func getReadmeContent(name, template string) string {
	return fmt.Sprintf(`# %s

A plumego application built with the **%s** template.

## Getting Started

### Install dependencies
`+"```bash"+`
go mod tidy
`+"```"+`

### Run development server
`+"```bash"+`
plumego dev
# or
go run main.go
`+"```"+`

### Run tests
`+"```bash"+`
plumego test
# or
go test ./...
`+"```"+`

### Build
`+"```bash"+`
plumego build
# or
go build -o bin/app
`+"```"+`

## Documentation

See [Plumego documentation](https://github.com/spcent/plumego) for more information.
`, name, template)
}
