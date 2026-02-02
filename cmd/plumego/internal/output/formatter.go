package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Formatter handles output formatting for CLI
type Formatter struct {
	format  string // json, yaml, text
	quiet   bool
	verbose bool
	color   bool
}

// NewFormatter creates a new output formatter
func NewFormatter() *Formatter {
	return &Formatter{
		format: "json",
		color:  true,
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

// Print outputs data in the configured format
func (f *Formatter) Print(data interface{}) error {
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
func (f *Formatter) Success(message string, data interface{}) error {
	if f.quiet {
		return nil
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": message,
	}

	if data != nil {
		result["data"] = data
	}

	return f.Print(result)
}

// Error outputs an error message
func (f *Formatter) Error(message string, code int, optionalData ...interface{}) error {
	result := map[string]interface{}{
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

	os.Exit(code)
	return nil
}

// Verbose outputs verbose logging
func (f *Formatter) Verbose(message string) {
	if f.verbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] %s\n", message)
	}
}

// Info outputs informational message
func (f *Formatter) Info(message string) {
	if !f.quiet {
		fmt.Fprintf(os.Stderr, "[INFO] %s\n", message)
	}
}

func (f *Formatter) printJSON(data interface{}) error {
	// If data is already a string, just print it
	if str, ok := data.(string); ok {
		fmt.Println(str)
		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) printYAML(data interface{}) error {
	// If data is already a string, just print it
	if str, ok := data.(string); ok {
		fmt.Println(str)
		return nil
	}

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

func (f *Formatter) printText(data interface{}) error {
	// Simple text output
	fmt.Println(data)
	return nil
}
