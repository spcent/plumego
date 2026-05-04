package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestSnapshotGenDeclOmitsUnexportedStructFields(t *testing.T) {
	src := `package sample

type Exported struct {
	Visible string
	hidden string
	AlsoVisible, alsoHidden int
	*Embedded
	*hiddenEmbedded
}

type Embedded struct{}
type hiddenEmbedded struct{}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}

	var lines []string
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			lines = append(lines, snapshotGenDecl(fset, "example.com/sample", gen)...)
		}
	}

	got := strings.Join(lines, "\n")
	if !strings.Contains(got, "Visible string") {
		t.Fatalf("expected exported field in snapshot, got:\n%s", got)
	}
	if !strings.Contains(got, "AlsoVisible int") {
		t.Fatalf("expected exported field from mixed field list in snapshot, got:\n%s", got)
	}
	if !strings.Contains(got, "*Embedded") {
		t.Fatalf("expected exported embedded field in snapshot, got:\n%s", got)
	}
	for _, forbidden := range []string{"hidden string", "alsoHidden", "hiddenEmbedded"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("snapshot includes unexported field %q:\n%s", forbidden, got)
		}
	}
}
