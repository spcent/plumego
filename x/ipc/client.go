package ipc

import (
	"context"
	"io"
	"net"
	"time"
)

// Reader defines the read interface with timeout support
type Reader interface {
	io.Reader
	ReadWithTimeout(buf []byte, timeout time.Duration) (int, error)
}

// Writer defines the write interface with timeout support
type Writer interface {
	io.Writer
	WriteWithTimeout(data []byte, timeout time.Duration) (int, error)
}

// AddrProvider defines the interface for getting remote address
type AddrProvider interface {
	// RemoteAddr returns the remote network address as net.Addr
	RemoteAddr() net.Addr
	// RemoteAddrString returns the remote address as a string (for backward compatibility)
	RemoteAddrString() string
}

// Client defines the IPC client interface
// It composes Reader, Writer, Closer, and AddrProvider interfaces
type Client interface {
	Reader
	Writer
	io.Closer
	AddrProvider
}

// Dial connects to a server at given address with default config
func Dial(addr string, opts ...Option) (Client, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return dialPlatform(addr, config)
}

// DialWithContext connects to a server with context
func DialWithContext(ctx context.Context, addr string, opts ...Option) (Client, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return dialPlatformWithContext(ctx, addr, config)
}
