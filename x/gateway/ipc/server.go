package ipc

import "context"

// Server defines the IPC server interface
type Server interface {
	// Accept waits for and returns the next connection to the server
	Accept() (Client, error)
	// AcceptWithContext waits for and returns the next connection to the server with context
	AcceptWithContext(ctx context.Context) (Client, error)
	// Addr returns the server's address
	Addr() string
	// Close closes the server immediately, releasing any resources
	Close() error
	// Shutdown gracefully shuts down the server without interrupting active connections.
	// It first stops accepting new connections, then waits for existing Accept calls
	// to complete or for the context to be cancelled.
	Shutdown(ctx context.Context) error
}

// NewServer creates a new IPC server with default config
func NewServer(addr string, opts ...Option) (Server, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return newPlatformServer(addr, config)
}
