package health

import (
	"context"
	"time"
)

// MockChecker is a simple ComponentChecker used by manager and metrics tests.
type MockChecker struct {
	name    string
	healthy bool
	delay   time.Duration
}

func (mc *MockChecker) Name() string {
	return mc.name
}

func (mc *MockChecker) Check(ctx context.Context) error {
	if mc.delay > 0 {
		select {
		case <-time.After(mc.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if mc.healthy {
		return nil
	}
	return MockError("mock component failure")
}

type MockError string

func (e MockError) Error() string {
	return string(e)
}
