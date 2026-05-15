package discovery

import "errors"

var (
	// ErrNotSupported is returned when an operation is not supported.
	ErrNotSupported = errors.New("discovery: operation not supported")
)

func unsupportedOperation(_ string) error {
	return ErrNotSupported
}
