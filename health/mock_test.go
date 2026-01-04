package health

import (
	"context"
	"errors"
)

// mockComponent is a simple mock implementation of ComponentChecker for testing
type mockComponent struct {
	name    string
	healthy bool
	err     error
}

func (m *mockComponent) Name() string {
	return m.name
}

func (m *mockComponent) Check(ctx context.Context) error {
	if m.err != nil {
		return m.err
	}
	if !m.healthy {
		return errors.New("mock component unhealthy")
	}
	return nil
}

// mockError is a simple error type for testing
type mockError string

func (e mockError) Error() string {
	return string(e)
}
