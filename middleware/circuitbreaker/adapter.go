// Package circuitbreaker provides a thin HTTP adapter for resilience circuit breaking.
package circuitbreaker

import (
	"net/http"

	resilience "github.com/spcent/plumego/resilience/circuitbreaker"
)

type (
	State          = resilience.State
	CircuitBreaker = resilience.CircuitBreaker
	Config         = resilience.Config
	Counts         = resilience.Counts
	Stats          = resilience.Stats
	Clock          = resilience.Clock
	ErrorHandler   = resilience.ErrorHandler
)

const (
	StateClosed   = resilience.StateClosed
	StateOpen     = resilience.StateOpen
	StateHalfOpen = resilience.StateHalfOpen
)

var (
	ErrCircuitOpen     = resilience.ErrCircuitOpen
	ErrTooManyRequests = resilience.ErrTooManyRequests
	ErrServerError     = resilience.ErrServerError
	New                = resilience.New
)

func Middleware(config Config) func(http.Handler) http.Handler {
	return resilience.Middleware(config)
}

func MiddlewareWithErrorHandler(config Config, errorHandler ErrorHandler) func(http.Handler) http.Handler {
	return resilience.MiddlewareWithErrorHandler(config, errorHandler)
}
