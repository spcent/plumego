// ipc.go
package ipc

import (
	"context"
	"io"
	"time"
)

// Config holds configuration for IPC connections
type Config struct {
	ConnectTimeout time.Duration // Timeout for connection establishment
	ReadTimeout    time.Duration // Timeout for read operations
	WriteTimeout   time.Duration // Timeout for write operations
	BufferSize     int           // Buffer size for read/write operations
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ConnectTimeout: 10 * time.Second,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		BufferSize:     4096,
	}
}

// Server defines the IPC server interface
type Server interface {
	// Accept waits for and returns the next connection to the server
	Accept() (Client, error)
	// AcceptWithContext waits for and returns the next connection to the server with context
	AcceptWithContext(ctx context.Context) (Client, error)
	// Addr returns the server's address
	Addr() string
	// Close closes the server, releasing any resources
	Close() error
}

// Client defines the IPC client interface
type Client interface {
	io.ReadWriteCloser
	WriteWithTimeout(data []byte, timeout time.Duration) (int, error)
	ReadWithTimeout(buf []byte, timeout time.Duration) (int, error)
	RemoteAddr() string
}

// NewServer creates a new IPC server with default config
func NewServer(addr string) (Server, error) {
	return NewServerWithConfig(addr, DefaultConfig())
}

// NewServerWithConfig creates a new IPC server with custom config
func NewServerWithConfig(addr string, config *Config) (Server, error) {
	return newPlatformServer(addr, config)
}

// Dial connects to a server at given address with default config
func Dial(addr string) (Client, error) {
	return DialWithConfig(addr, DefaultConfig())
}

// DialWithConfig connects to a server with custom config
func DialWithConfig(addr string, config *Config) (Client, error) {
	return dialPlatform(addr, config)
}

// DialWithContext connects to a server with context
func DialWithContext(ctx context.Context, addr string, config *Config) (Client, error) {
	return dialPlatformWithContext(ctx, addr, config)
}
