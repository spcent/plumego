package file

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates the requested file was not found.
	ErrNotFound = errors.New("file: not found")

	// ErrAlreadyExists indicates a file already exists at the path.
	ErrAlreadyExists = errors.New("file: already exists")

	// ErrInvalidPath indicates an invalid or unsafe file path.
	ErrInvalidPath = errors.New("file: invalid path")

	// ErrInvalidSize indicates an invalid file size.
	ErrInvalidSize = errors.New("file: invalid size")

	// ErrUnsupportedFormat indicates an unsupported file format.
	ErrUnsupportedFormat = errors.New("file: unsupported format")

	// ErrStorageUnavailable indicates the storage backend is unavailable.
	ErrStorageUnavailable = errors.New("file: storage unavailable")
)

// Error represents a file operation error with context.
type Error struct {
	Op   string // Operation name
	Path string // File path (if applicable)
	Err  error  // Underlying error
}

// Error returns the error message.
func (e *Error) Error() string {
	if e == nil {
		return "file: <nil>"
	}

	message := "file:"
	if e.Op != "" {
		message += " " + e.Op
	}
	if e.Path != "" {
		message += " " + e.Path
	}
	if message == "file:" {
		message = "file: error"
	}
	if e.Err == nil {
		return message
	}
	return fmt.Sprintf("%s: %v", message, e.Err)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
