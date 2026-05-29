package common

import "errors"

var (
	// ErrEndpointNotFound is returned when no endpoint matches the requested key.
	ErrEndpointNotFound = errors.New("endpoint not found")
	// ErrInvalidTimeRange is returned when "from" is after "to".
	ErrInvalidTimeRange = errors.New("'from' cannot be older than 'to'")
)
