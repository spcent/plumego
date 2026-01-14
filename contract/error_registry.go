package contract

import (
	"fmt"
	"strings"
	"sync"
)

// ErrorRegistry manages custom error codes and their documentation
type ErrorRegistry struct {
	mu    sync.RWMutex
	codes map[string]ErrorEntry
}

// ErrorEntry represents a registered error code
type ErrorEntry struct {
	Code     ErrorCode
	Message  string
	Category ErrorCategory
}

// NewErrorRegistry creates a new error registry
func NewErrorRegistry() *ErrorRegistry {
	return &ErrorRegistry{
		codes: make(map[string]ErrorEntry),
	}
}

// Register registers a new error code
func (er *ErrorRegistry) Register(code string, message string, category ErrorCategory) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.codes[code] = ErrorEntry{
		Code:     ErrorCode(code),
		Message:  message,
		Category: category,
	}
}

// Get retrieves an error entry by code
func (er *ErrorRegistry) Get(code string) (ErrorEntry, bool) {
	er.mu.RLock()
	defer er.mu.RUnlock()

	entry, exists := er.codes[code]
	return entry, exists
}

// List returns all registered error codes
func (er *ErrorRegistry) List() []ErrorEntry {
	er.mu.RLock()
	defer er.mu.RUnlock()

	var entries []ErrorEntry
	for _, entry := range er.codes {
		entries = append(entries, entry)
	}
	return entries
}

// GenerateDocumentation generates documentation for all error codes
func (er *ErrorRegistry) GenerateDocumentation() string {
	var builder strings.Builder

	entries := er.List()
	if len(entries) == 0 {
		return "No custom error codes registered.\n"
	}

	// Group by category
	categories := make(map[ErrorCategory][]ErrorEntry)
	for _, entry := range entries {
		categories[entry.Category] = append(categories[entry.Category], entry)
	}

	builder.WriteString("# Custom Error Codes\n\n")
	builder.WriteString("This document describes all custom error codes used in the application.\n\n")

	for category, categoryEntries := range categories {
		builder.WriteString(fmt.Sprintf("## %s\n\n", category))
		builder.WriteString("| Code | Message |\n")
		builder.WriteString("|------|---------|\n")
		for _, entry := range categoryEntries {
			builder.WriteString(fmt.Sprintf("| `%s` | %s |\n", entry.Code, entry.Message))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// CreateStructuredError creates a structured error using a registered code
func (er *ErrorRegistry) CreateStructuredError(code string, detail *ErrorDetail) error {
	entry, exists := er.Get(code)
	if !exists {
		return NewStructuredError(ErrInternal, "Unknown error code: "+code, nil)
	}

	metadata := make(map[string]interface{})
	if detail != nil {
		metadata["detail"] = *detail
	}

	return NewStructuredError(entry.Code, entry.Message, metadata)
}

// MustCreateStructuredError creates a structured error and panics if code doesn't exist
func (er *ErrorRegistry) MustCreateStructuredError(code string, detail *ErrorDetail) error {
	entry, exists := er.Get(code)
	if !exists {
		panic(fmt.Sprintf("error code %s not registered", code))
	}

	metadata := make(map[string]interface{})
	if detail != nil {
		metadata["detail"] = *detail
	}

	return NewStructuredError(entry.Code, entry.Message, metadata)
}
