package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateModel(dir string, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Type:    "model",
		Name:    opts.Name,
		Files:   make(map[string][]string),
		Imports: []string{},
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		modelName := strings.ToLower(opts.Name)
		outputPath = filepath.Join(dir, "internal", "domain", modelName, modelName+".go")
	}

	packageName := opts.PackageName
	if packageName == "" {
		packageName = strings.ToLower(opts.Name)
	}
	if err := validateGoIdentifier("package", packageName); err != nil {
		return nil, err
	}

	if err := validateOutputPath(outputPath, opts.Force); err != nil {
		return nil, err
	}

	content := generateModelCode(opts.Name, packageName, opts.WithValidation)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result.Files["created"] = []string{outputPath}

	return result, nil
}

func generateModelCode(name, pkg string, withValidation bool) string {
	validation := ""
	if withValidation {
		validation = fmt.Sprintf(`
// Validate checks required fields.
func (m *%s) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}`, name)
	}

	imports := ""
	if withValidation {
		imports = `
import (
	"fmt"
	"strings"
)
`
	}

	return fmt.Sprintf(`package %s
%s
// %s represents a %s entity.
type %s struct {
	ID        int64  `+"`json:\"id\"`"+`
	CreatedAt string `+"`json:\"created_at\"`"+`
	UpdatedAt string `+"`json:\"updated_at\"`"+`
	Name      string `+"`json:\"name\"`"+`
}
%s
`, pkg, imports, name, strings.ToLower(name), name, validation)
}
