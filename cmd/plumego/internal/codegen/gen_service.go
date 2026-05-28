package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateService(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:    "service",
		Name:    opts.Name,
		Files:   make(map[string][]string),
		Imports: []string{"context"},
	}
	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(dir, "internal", "domain", strings.ToLower(opts.Name), "service.go")
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
	content := generateServiceCode(opts.Name, packageName)
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	result.Files["created"] = []string{outputPath}
	return result, nil
}

func generateServiceCode(name, pkg string) string {
	lower := strings.ToLower(name)
	return fmt.Sprintf(`package %s

import "context"

// %s is the domain entity.
type %s struct {
	ID   string
	Name string
}

// Create%sRequest carries input for creating a %s.
type Create%sRequest struct {
	Name string
}

// %sService defines the domain operations for %s.
type %sService interface {
	Get(ctx context.Context, id string) (*%s, error)
	Create(ctx context.Context, req Create%sRequest) (*%s, error)
}

// %sServiceImpl is the default implementation stub.
type %sServiceImpl struct{}

// New%sService returns a new %sServiceImpl.
func New%sService() %sService {
	return &%sServiceImpl{}
}

func (s *%sServiceImpl) Get(_ context.Context, _ string) (*%s, error) {
	return nil, nil
}

func (s *%sServiceImpl) Create(_ context.Context, _ Create%sRequest) (*%s, error) {
	return nil, nil
}
`,
		pkg,
		name, name,
		name, lower,
		name,
		name, lower,
		name,
		name,
		name, name,
		name,
		name,
		name, name,
		name, name,
		name,
		name, name,
		name,
		name, name,
	)
}
