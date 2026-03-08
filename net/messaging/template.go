package messaging

import (
	"fmt"
	"strings"
	"sync"
)

// TemplateEngine renders message bodies from registered templates.
type TemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]string // id -> template body with {{key}} placeholders
}

// NewTemplateEngine creates an empty engine.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{templates: make(map[string]string)}
}

// Register adds or replaces a template.
func (e *TemplateEngine) Register(id, body string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.templates[id] = body
}

// Render applies params to the named template.
// Placeholders use the form {{key}}.
func (e *TemplateEngine) Render(id string, params map[string]string) (string, error) {
	e.mu.RLock()
	tmpl, ok := e.templates[id]
	e.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("%w: template %q not found", ErrTemplateRender, id)
	}
	result := tmpl
	for k, v := range params {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result, nil
}

// Has reports whether a template is registered.
func (e *TemplateEngine) Has(id string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.templates[id]
	return ok
}
