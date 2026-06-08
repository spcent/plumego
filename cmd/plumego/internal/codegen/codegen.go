package codegen

import (
	"fmt"
	"go/token"
	"os"
	"strings"
	"unicode"
)

// GenerateOptions represents code generation options
type GenerateOptions struct {
	Type           string
	Name           string
	OutputPath     string
	PackageName    string
	Methods        string
	WithTests      bool
	WithValidation bool
	Force          bool
}

// GenerateResult represents the result of code generation
type GenerateResult struct {
	Type    string              `json:"type" yaml:"type"`
	Name    string              `json:"name" yaml:"name"`
	Files   map[string][]string `json:"files" yaml:"files"`
	Imports []string            `json:"imports,omitempty" yaml:"imports,omitempty"`
}

// Generate generates code based on options
func Generate(dir string, opts GenerateOptions) (*GenerateResult, error) {
	// health-handler and metrics-middleware use fixed function names; name is display-only.
	if !isNameOptionalType(opts.Type) {
		if err := validateGoIdentifier("name", opts.Name); err != nil {
			return nil, err
		}
	}
	switch opts.Type {
	case "middleware":
		return generateMiddleware(dir, opts)
	case "handler":
		return generateHandler(dir, opts)
	case "model":
		return generateModel(dir, opts)
	case "endpoint":
		return generateEndpoint(dir, opts)
	case "service":
		return generateService(dir, opts)
	case "repo":
		return generateRepo(dir, opts)
	case "health-handler":
		return generateHealthHandler(dir, opts)
	case "metrics-middleware":
		return generateMetricsMiddleware(dir, opts)
	default:
		return nil, fmt.Errorf("unknown generation type %q (valid: handler, endpoint, middleware, health-handler, metrics-middleware, model, service, repo, spec)", opts.Type)
	}
}

// isNameOptionalType reports whether opts.Type does not require a Name argument.
func isNameOptionalType(t string) bool {
	return t == "health-handler" || t == "metrics-middleware"
}

func parseHTTPMethods(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	methods := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		method := strings.ToUpper(strings.TrimSpace(part))
		if method == "" {
			return nil, fmt.Errorf("empty HTTP method")
		}
		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE":
		default:
			return nil, fmt.Errorf("unsupported HTTP method: %s (valid: GET, POST, PUT, PATCH, DELETE)", method)
		}
		if _, ok := seen[method]; ok {
			continue
		}
		seen[method] = struct{}{}
		methods = append(methods, method)
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("at least one HTTP method is required")
	}
	return methods, nil
}

func validateOutputPaths(paths []string, force bool) error {
	for _, path := range paths {
		if err := validateOutputPath(path, force); err != nil {
			return err
		}
	}
	return nil
}

func validateOutputPath(path string, force bool) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to inspect output path %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("output path is a directory: %s", path)
	}
	if !force {
		return fmt.Errorf("file %s already exists (use --force to overwrite)", path)
	}
	return nil
}

func validateGoIdentifier(label, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}
	for i, r := range value {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return fmt.Errorf("%s %q is not a valid Go identifier", label, value)
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("%s %q is not a valid Go identifier", label, value)
		}
	}
	if token.Lookup(value).IsKeyword() {
		return fmt.Errorf("%s %q is not a valid Go identifier", label, value)
	}
	return nil
}
