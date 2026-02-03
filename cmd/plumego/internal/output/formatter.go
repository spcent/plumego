package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Formatter handles output formatting for CLI
type Formatter struct {
	format  string // json, yaml, text
	quiet   bool
	verbose bool
	color   bool
	out     io.Writer
	err     io.Writer
	mu      sync.Mutex
}

// NewFormatter creates a new output formatter
func NewFormatter() *Formatter {
	return &Formatter{
		format: "json",
		color:  true,
		out:    os.Stdout,
		err:    os.Stderr,
	}
}

// SetFormat sets the output format
func (f *Formatter) SetFormat(format string) {
	f.format = format
}

// SetQuiet sets quiet mode
func (f *Formatter) SetQuiet(quiet bool) {
	f.quiet = quiet
}

// SetVerbose sets verbose mode
func (f *Formatter) SetVerbose(verbose bool) {
	f.verbose = verbose
}

// SetColor sets color output
func (f *Formatter) SetColor(color bool) {
	f.color = color
}

// SetWriters sets the output and error writers.
func (f *Formatter) SetWriters(out, err io.Writer) {
	if out != nil {
		f.out = out
	}
	if err != nil {
		f.err = err
	}
}

// Format returns the current output format.
func (f *Formatter) Format() string {
	return f.format
}

// IsQuiet returns whether quiet mode is enabled.
func (f *Formatter) IsQuiet() bool {
	return f.quiet
}

// Print outputs data in the configured format
func (f *Formatter) Print(data any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch f.format {
	case "json":
		return f.printJSON(data)
	case "yaml":
		return f.printYAML(data)
	case "text":
		return f.printText(data)
	default:
		return f.printJSON(data)
	}
}

// Success outputs a success message
func (f *Formatter) Success(message string, data any) error {
	if f.quiet {
		return nil
	}

	result := map[string]any{
		"status":  "success",
		"message": message,
	}

	if data != nil {
		result["data"] = data
	}

	return f.Print(result)
}

// Error outputs an error message
func (f *Formatter) Error(message string, code int, optionalData ...any) error {
	result := map[string]any{
		"status":    "error",
		"message":   message,
		"exit_code": code,
	}

	// Add optional data if provided
	if len(optionalData) > 0 && optionalData[0] != nil {
		result["data"] = optionalData[0]
	}

	if err := f.Print(result); err != nil {
		return err
	}

	return &ExitError{Code: code, Message: message}
}

// Verbose outputs verbose logging
func (f *Formatter) Verbose(message string) {
	if f.verbose {
		f.mu.Lock()
		defer f.mu.Unlock()
		fmt.Fprintf(f.err, "[VERBOSE] %s\n", message)
	}
}

// Info outputs informational message
func (f *Formatter) Info(message string) {
	if !f.quiet {
		f.mu.Lock()
		defer f.mu.Unlock()
		fmt.Fprintf(f.err, "[INFO] %s\n", message)
	}
}

func (f *Formatter) printJSON(data any) error {
	// If data is already a string, just print it
	if str, ok := data.(string); ok {
		fmt.Fprintln(f.out, str)
		return nil
	}

	encoder := json.NewEncoder(f.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) printYAML(data any) error {
	// If data is already a string, just print it
	if str, ok := data.(string); ok {
		fmt.Fprintln(f.out, str)
		return nil
	}

	encoder := yaml.NewEncoder(f.out)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

func (f *Formatter) printText(data any) error {
	// Simple text output
	fmt.Fprintln(f.out, data)
	return nil
}
