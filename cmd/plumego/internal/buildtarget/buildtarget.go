package buildtarget

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Default returns the default build target for a Plumego application project.
func Default(dir string) string {
	if HasMainPackage(dir) {
		return "."
	}
	if HasMainPackage(filepath.Join(dir, "cmd", "app")) {
		return "./cmd/app"
	}
	return "."
}

// HasDefaultEntrypoint reports whether the project has a supported main package.
func HasDefaultEntrypoint(dir string) bool {
	return HasMainPackage(dir) || HasMainPackage(filepath.Join(dir, "cmd", "app"))
}

// HasMainPackage reports whether a directory contains at least one non-test Go
// file whose package clause is `package main`.
func HasMainPackage(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if fileHasMainPackage(filepath.Join(dir, entry.Name())) {
			return true
		}
	}
	return false
}

func fileHasMainPackage(path string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.PackageClauseOnly)
	if err != nil || file == nil || file.Name == nil {
		return false
	}
	return file.Name.Name == "main"
}
