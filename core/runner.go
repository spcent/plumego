package core

import (
	"context"
)

// Runner is a compatibility interface for legacy lifecycle helpers.
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
