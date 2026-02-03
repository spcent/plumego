package output

import (
	"errors"
	"fmt"
)

// ExitError represents a controlled process exit with a specific code.
type ExitError struct {
	Code    int
	Message string
	Err     error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		if e.Message != "" {
			return fmt.Sprintf("%s: %v", e.Message, e.Err)
		}
		return e.Err.Error()
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

// Exit returns an error that maps to the provided exit code.
func Exit(code int) error {
	return &ExitError{Code: code}
}

// ExitCode extracts the exit code from an error when present.
func ExitCode(err error) (int, bool) {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code, true
	}
	return 0, false
}
