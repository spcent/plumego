package messaging

import (
	"errors"
	"testing"
)

func TestTemplateEngine_RenderBasic(t *testing.T) {
	e := NewTemplateEngine()
	e.Register("welcome", "Hello {{name}}, welcome to {{site}}!")

	got, err := e.Render("welcome", map[string]string{
		"name": "Alice",
		"site": "Plumego",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "Hello Alice, welcome to Plumego!"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTemplateEngine_RenderNotFound(t *testing.T) {
	e := NewTemplateEngine()
	_, err := e.Render("nonexistent", nil)
	if !errors.Is(err, ErrTemplateRender) {
		t.Fatalf("expected ErrTemplateRender, got %v", err)
	}
}

func TestTemplateEngine_Has(t *testing.T) {
	e := NewTemplateEngine()
	if e.Has("x") {
		t.Fatal("should not have 'x'")
	}
	e.Register("x", "body")
	if !e.Has("x") {
		t.Fatal("should have 'x'")
	}
}

func TestTemplateEngine_ReplaceExisting(t *testing.T) {
	e := NewTemplateEngine()
	e.Register("t", "v1: {{val}}")
	e.Register("t", "v2: {{val}}")

	got, _ := e.Render("t", map[string]string{"val": "test"})
	if got != "v2: test" {
		t.Fatalf("expected replacement, got %q", got)
	}
}
