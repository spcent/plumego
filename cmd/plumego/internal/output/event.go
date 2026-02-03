package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Event represents a streaming output item.
type Event struct {
	Event   string         `json:"event" yaml:"event"`
	Level   string         `json:"level,omitempty" yaml:"level,omitempty"`
	Message string         `json:"message,omitempty" yaml:"message,omitempty"`
	Time    string         `json:"time,omitempty" yaml:"time,omitempty"`
	Data    map[string]any `json:"data,omitempty" yaml:"data,omitempty"`
}

// Event outputs a structured event in the configured format.
func (f *Formatter) Event(e Event) error {
	level := strings.ToLower(strings.TrimSpace(e.Level))
	if level == "debug" && !f.verbose {
		return nil
	}
	if f.quiet && (level == "" || level == "info" || level == "debug") {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch f.format {
	case "json":
		return f.printEventJSON(e)
	case "yaml":
		return f.printEventYAML(e)
	case "text":
		return f.printEventText(e, level)
	default:
		return f.printEventJSON(e)
	}
}

// Line outputs a message as a generic event.
func (f *Formatter) Line(message string) error {
	return f.Event(Event{
		Event:   "message",
		Level:   "info",
		Message: message,
	})
}

// Linef outputs a formatted message as a generic event.
func (f *Formatter) Linef(format string, args ...any) error {
	return f.Line(fmt.Sprintf(format, args...))
}

// Textln outputs a plain text line when format is text.
func (f *Formatter) Textln(message string) error {
	if f.format != "text" || f.quiet {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	_, err := fmt.Fprintln(f.out, message)
	return err
}

// Textf outputs a formatted plain text message when format is text.
func (f *Formatter) Textf(format string, args ...any) error {
	if f.format != "text" || f.quiet {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	_, err := fmt.Fprintf(f.out, format, args...)
	return err
}

func (f *Formatter) printEventJSON(e Event) error {
	encoder := json.NewEncoder(f.out)
	return encoder.Encode(e)
}

func (f *Formatter) printEventYAML(e Event) error {
	encoder := yaml.NewEncoder(f.out)
	encoder.SetIndent(2)
	return encoder.Encode(e)
}

func (f *Formatter) printEventText(e Event, level string) error {
	message := e.Message
	if message == "" {
		message = e.Event
	}
	if message == "" {
		return nil
	}

	prefix := ""
	switch level {
	case "error":
		prefix = f.colorize("error", "ERROR: ")
	case "warn":
		prefix = f.colorize("warn", "WARN: ")
	case "debug":
		prefix = f.colorize("debug", "DEBUG: ")
	}

	writer := f.out
	if level == "error" || level == "warn" || level == "debug" {
		writer = f.err
	}
	_, err := fmt.Fprintln(writer, prefix+message)
	return err
}
