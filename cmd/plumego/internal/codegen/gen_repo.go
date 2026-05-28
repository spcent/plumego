package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateRepo(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:    "repo",
		Name:    opts.Name,
		Files:   make(map[string][]string),
		Imports: []string{"context"},
	}
	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(dir, "internal", "domain", strings.ToLower(opts.Name), "repository.go")
	}
	packageName := opts.PackageName
	if packageName == "" {
		packageName = strings.ToLower(opts.Name)
	}
	if err := validateGoIdentifier("package", packageName); err != nil {
		return nil, err
	}
	if err := validateOutputPaths([]string{outputPath}, opts.Force); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	content := generateRepoCode(opts.Name, packageName)
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	result.Files["created"] = []string{outputPath}
	return result, nil
}

func generateRepoCode(name, pkg string) string {
	lower := strings.ToLower(name)
	return fmt.Sprintf(`package %s

import "context"

// %sRepository is the persistence boundary for the %s domain.
// Implement this interface to connect any storage backend.
type %sRepository interface {
	FindByID(ctx context.Context, id string) (*%s, error)
	Save(ctx context.Context, entity *%s) error
	Delete(ctx context.Context, id string) error
}

// %sRepositoryStub is a no-op stub for tests and early development.
type %sRepositoryStub struct{}

func (r *%sRepositoryStub) FindByID(_ context.Context, _ string) (*%s, error) { return nil, nil }
func (r *%sRepositoryStub) Save(_ context.Context, _ *%s) error               { return nil }
func (r *%sRepositoryStub) Delete(_ context.Context, _ string) error          { return nil }
`,
		pkg,
		name, lower,
		name,
		name,
		name,
		name,
		name,
		name, name,
		name, name,
		name,
	)
}
