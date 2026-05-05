package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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

type commandResult struct {
	Status   string `json:"status" yaml:"status"`
	Message  string `json:"message" yaml:"message"`
	ExitCode int    `json:"exit_code,omitempty" yaml:"exit_code,omitempty"`
	Data     any    `json:"data,omitempty" yaml:"data,omitempty"`
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

// IsSupportedFormat reports whether format is a stable CLI output format.
func IsSupportedFormat(format string) bool {
	switch format {
	case "json", "yaml", "text":
		return true
	default:
		return false
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

// IsVerbose returns whether verbose mode is enabled.
func (f *Formatter) IsVerbose() bool {
	return f.verbose
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
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

// Success outputs a success message
func (f *Formatter) Success(message string, data any) error {
	if f.quiet {
		return nil
	}

	// CLI output intentionally uses a command-result envelope instead of
	// contract.Response, which is reserved for HTTP handlers.
	result := commandResult{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	return f.Print(result)
}

// Warning outputs a warning message and returns a non-zero exit code.
func (f *Formatter) Warning(message string, code int, data any) error {
	result := commandResult{
		Status:   "warning",
		Message:  message,
		ExitCode: code,
		Data:     data,
	}

	if err := f.Print(result); err != nil {
		return err
	}

	return &ExitError{Code: code, Message: message}
}

// Error outputs an error message
func (f *Formatter) Error(message string, code int, optionalData ...any) error {
	// CLI output intentionally reports process status and exit code. Keep this
	// separate from HTTP contract.ErrorResponse shapes.
	result := commandResult{
		Status:   "error",
		Message:  message,
		ExitCode: code,
	}

	// Add optional data if provided
	if len(optionalData) > 0 && optionalData[0] != nil {
		result.Data = optionalData[0]
	}

	if err := f.Print(result); err != nil {
		return err
	}

	return &ExitError{Code: code, Message: message}
}

// Verbose outputs verbose logging
func (f *Formatter) Verbose(message string) {
	_ = f.Event(Event{
		Event:   "cli.verbose",
		Level:   "debug",
		Message: message,
	})
}

// Info outputs informational message
func (f *Formatter) Info(message string) {
	_ = f.Event(Event{
		Event:   "cli.info",
		Level:   "info",
		Message: message,
	})
}

func (f *Formatter) printJSON(data any) error {
	if str, ok := data.(string); ok {
		data = commandResult{
			Status:  "success",
			Message: "output",
			Data: map[string]string{
				"value": str,
			},
		}
	}

	encoder := json.NewEncoder(f.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) printYAML(data any) error {
	if str, ok := data.(string); ok {
		data = commandResult{
			Status:  "success",
			Message: "output",
			Data: map[string]string{
				"value": str,
			},
		}
	}

	encoder := yaml.NewEncoder(f.out)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

func (f *Formatter) printText(data any) error {
	if result, ok := data.(commandResult); ok {
		return f.printCommandResultText(result)
	}
	if result, ok := data.(*commandResult); ok && result != nil {
		return f.printCommandResultText(*result)
	}
	fmt.Fprintln(f.out, data)
	return nil
}

func (f *Formatter) printCommandResultText(result commandResult) error {
	prefix := strings.ToUpper(result.Status)
	if prefix == "" {
		prefix = "RESULT"
	}
	line := prefix + ": " + result.Message
	if result.ExitCode != 0 {
		line = fmt.Sprintf("%s (exit %d)", line, result.ExitCode)
	}
	writer := f.out
	if result.Status == "error" || result.Status == "warning" {
		writer = f.err
	}
	if _, err := fmt.Fprintln(writer, line); err != nil {
		return err
	}
	if result.Data == nil {
		return nil
	}
	encoded, err := json.MarshalIndent(result.Data, "", "  ")
	if err != nil {
		_, err = fmt.Fprintf(writer, "%v\n", result.Data)
		return err
	}
	_, err = fmt.Fprintf(writer, "%s\n", encoded)
	return err
}

func (f *Formatter) colorize(level, text string) string {
	if !f.color || f.format != "text" || text == "" {
		return text
	}
	const (
		reset  = "\x1b[0m"
		red    = "\x1b[31m"
		yellow = "\x1b[33m"
		blue   = "\x1b[34m"
	)
	switch level {
	case "error":
		return red + text + reset
	case "warn":
		return yellow + text + reset
	case "debug":
		return blue + text + reset
	default:
		return text
	}
}
