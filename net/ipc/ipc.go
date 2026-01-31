// Package ipc provides cross-platform inter-process communication (IPC) primitives.
//
// # Supported Transports
//
// - Unix Domain Sockets (Linux, macOS, BSD)
// - Windows Named Pipes (Windows)
// - TCP sockets (fallback for all platforms)
//
// # Performance Characteristics
//
// Unix Domain Sockets offer approximately 50% lower latency than TCP and 30% higher
// throughput for local communication. Windows Named Pipes provide similar performance
// benefits on Windows. TCP is universal but slower, suitable for network IPC.
//
// # Basic Usage
//
//	// Server
//	server, err := ipc.NewServer("/tmp/myapp.sock")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer server.Close()
//
//	// Accept connections
//	client, err := server.Accept()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Client
//	client, err := ipc.Dial("/tmp/myapp.sock")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// # Functional Options
//
// Configure timeouts, permissions, and buffer sizes using functional options:
//
//	server, err := ipc.NewServer("/tmp/myapp.sock",
//		ipc.WithUnixSocketPerm(0700),  // Owner only
//		ipc.WithTimeouts(5*time.Second, 30*time.Second, 30*time.Second),
//	)
//
//	client, err := ipc.Dial("/tmp/myapp.sock",
//		ipc.WithConnectTimeout(5*time.Second),
//		ipc.WithReadTimeout(10*time.Second),
//	)
//
// # Auto-Reconnection
//
// For resilient connections that automatically reconnect on failure:
//
//	reconnCfg := &ipc.ReconnectConfig{
//		MaxRetries:    5,
//		InitialDelay:  100 * time.Millisecond,
//		MaxDelay:      30 * time.Second,
//		BackoffFactor: 2.0,
//	}
//
//	client, err := ipc.DialWithReconnect("/tmp/myapp.sock", reconnCfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Write will automatically reconnect on connection failure
//	n, err := client.Write(data)
//
// # Thread Safety
//
// All types are safe for concurrent use from multiple goroutines.
//
// # Error Handling
//
// The package provides structured errors for better error handling:
//
//	if errors.Is(err, ipc.ErrClientClosed) {
//		// Handle closed client
//	}
//
//	var ipcErr *ipc.Error
//	if errors.As(err, &ipcErr) {
//		fmt.Printf("IPC %s error: %v\n", ipcErr.Op, ipcErr.Err)
//	}
package ipc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Common errors
var (
	// ErrServerClosed is returned when operations are attempted on a closed server
	ErrServerClosed = errors.New("ipc: server closed")
	// ErrClientClosed is returned when operations are attempted on a closed client
	ErrClientClosed = errors.New("ipc: client closed")
	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("ipc: invalid configuration")
	// ErrConnectTimeout is returned when connection times out
	ErrConnectTimeout = errors.New("ipc: connection timeout")
	// ErrPlatformNotSupported is returned when the platform is not supported
	ErrPlatformNotSupported = errors.New("ipc: platform not supported")
	// ErrInvalidAddress is returned when the address format is invalid
	ErrInvalidAddress = errors.New("ipc: invalid address")
)

// Error represents an IPC error with context
type Error struct {
	Op   string // Operation: "accept", "dial", "read", "write", "close"
	Addr string // Address
	Err  error  // Underlying error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Addr != "" {
		return fmt.Sprintf("ipc %s %s: %v", e.Op, e.Addr, e.Err)
	}
	return fmt.Sprintf("ipc %s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// Config holds configuration for IPC connections
type Config struct {
	ConnectTimeout   time.Duration // Timeout for connection establishment
	ReadTimeout      time.Duration // Timeout for read operations
	WriteTimeout     time.Duration // Timeout for write operations
	BufferSize       int           // Buffer size for read/write operations (Windows Named Pipe only)
	UnixSocketPerm   uint32        // Unix socket file permissions (e.g., 0700). Default: 0700 (owner only). Unix/Linux only.
	UnixSocketDirPerm uint32       // Unix socket directory permissions (e.g., 0755). Default: 0755. Unix/Linux only.
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ConnectTimeout:   10 * time.Second,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		BufferSize:       4096,
		UnixSocketPerm:   0700, // Owner only (rwx------)
		UnixSocketDirPerm: 0755, // Owner rwx, others rx (rwxr-xr-x)
	}
}

// Option is a functional option for configuring IPC connections
type Option func(*Config)

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = d
	}
}

// WithReadTimeout sets the read timeout
func WithReadTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = d
	}
}

// WithWriteTimeout sets the write timeout
func WithWriteTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = d
	}
}

// WithBufferSize sets the buffer size (Windows Named Pipe only)
func WithBufferSize(size int) Option {
	return func(c *Config) {
		c.BufferSize = size
	}
}

// WithUnixSocketPerm sets Unix socket file permissions (Unix/Linux only)
func WithUnixSocketPerm(perm uint32) Option {
	return func(c *Config) {
		c.UnixSocketPerm = perm
	}
}

// WithUnixSocketDirPerm sets Unix socket directory permissions (Unix/Linux only)
func WithUnixSocketDirPerm(perm uint32) Option {
	return func(c *Config) {
		c.UnixSocketDirPerm = perm
	}
}

// WithTimeouts sets all timeout values at once
func WithTimeouts(connect, read, write time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = connect
		c.ReadTimeout = read
		c.WriteTimeout = write
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
func NewServer(addr string, opts ...Option) (Server, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return newPlatformServer(addr, config)
}

// NewServerWithConfig creates a new IPC server with custom config
// Deprecated: Use NewServer with functional options instead
func NewServerWithConfig(addr string, config *Config) (Server, error) {
	return newPlatformServer(addr, config)
}

// Dial connects to a server at given address with default config
func Dial(addr string, opts ...Option) (Client, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return dialPlatform(addr, config)
}

// DialWithConfig connects to a server with custom config
// Deprecated: Use Dial with functional options instead
func DialWithConfig(addr string, config *Config) (Client, error) {
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

// ReconnectConfig holds configuration for auto-reconnection
type ReconnectConfig struct {
	MaxRetries    int           // Maximum number of retry attempts (0 = infinite)
	InitialDelay  time.Duration // Initial delay before first retry
	MaxDelay      time.Duration // Maximum delay between retries
	BackoffFactor float64       // Multiplier for exponential backoff (e.g., 2.0)
}

// DefaultReconnectConfig returns default reconnection configuration
func DefaultReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		MaxRetries:    5,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// reconnectClient wraps a Client with auto-reconnection capability
type reconnectClient struct {
	addr      string
	config    *Config
	reconn    *ReconnectConfig
	mu        sync.RWMutex
	client    Client
	closed    bool
	reconnecting bool
}

// DialWithReconnect creates a client with automatic reconnection on failure
func DialWithReconnect(addr string, reconnCfg *ReconnectConfig, opts ...Option) (Client, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	if reconnCfg == nil {
		reconnCfg = DefaultReconnectConfig()
	}

	client, err := dialPlatform(addr, config)
	if err != nil {
		return nil, err
	}

	return &reconnectClient{
		addr:   addr,
		config: config,
		reconn: reconnCfg,
		client: client,
	}, nil
}

func (rc *reconnectClient) reconnect() error {
	rc.mu.Lock()
	if rc.closed || rc.reconnecting {
		rc.mu.Unlock()
		return ErrClientClosed
	}
	rc.reconnecting = true
	rc.mu.Unlock()

	defer func() {
		rc.mu.Lock()
		rc.reconnecting = false
		rc.mu.Unlock()
	}()

	delay := rc.reconn.InitialDelay
	retries := 0

	for {
		// Check if max retries reached
		if rc.reconn.MaxRetries > 0 && retries >= rc.reconn.MaxRetries {
			return fmt.Errorf("max reconnection attempts (%d) reached", rc.reconn.MaxRetries)
		}

		// Wait before retry
		if retries > 0 {
			time.Sleep(delay)
			// Exponential backoff
			delay = time.Duration(float64(delay) * rc.reconn.BackoffFactor)
			if delay > rc.reconn.MaxDelay {
				delay = rc.reconn.MaxDelay
			}
		}

		// Attempt reconnection
		client, err := dialPlatform(rc.addr, rc.config)
		if err == nil {
			rc.mu.Lock()
			if rc.client != nil {
				rc.client.Close()
			}
			rc.client = client
			rc.mu.Unlock()
			return nil
		}

		retries++
	}
}

func (rc *reconnectClient) Write(data []byte) (int, error) {
	rc.mu.RLock()
	client := rc.client
	closed := rc.closed
	rc.mu.RUnlock()

	if closed {
		return 0, ErrClientClosed
	}

	n, err := client.Write(data)
	if err != nil && rc.shouldReconnect(err) {
		if reconnErr := rc.reconnect(); reconnErr == nil {
			// Retry after successful reconnection
			rc.mu.RLock()
			client = rc.client
			rc.mu.RUnlock()
			return client.Write(data)
		}
	}
	return n, err
}

func (rc *reconnectClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	rc.mu.RLock()
	client := rc.client
	closed := rc.closed
	rc.mu.RUnlock()

	if closed {
		return 0, ErrClientClosed
	}

	n, err := client.WriteWithTimeout(data, timeout)
	if err != nil && rc.shouldReconnect(err) {
		if reconnErr := rc.reconnect(); reconnErr == nil {
			// Retry after successful reconnection
			rc.mu.RLock()
			client = rc.client
			rc.mu.RUnlock()
			return client.WriteWithTimeout(data, timeout)
		}
	}
	return n, err
}

func (rc *reconnectClient) Read(buf []byte) (int, error) {
	rc.mu.RLock()
	client := rc.client
	closed := rc.closed
	rc.mu.RUnlock()

	if closed {
		return 0, ErrClientClosed
	}

	n, err := client.Read(buf)
	if err != nil && rc.shouldReconnect(err) {
		if reconnErr := rc.reconnect(); reconnErr == nil {
			// Retry after successful reconnection
			rc.mu.RLock()
			client = rc.client
			rc.mu.RUnlock()
			return client.Read(buf)
		}
	}
	return n, err
}

func (rc *reconnectClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	rc.mu.RLock()
	client := rc.client
	closed := rc.closed
	rc.mu.RUnlock()

	if closed {
		return 0, ErrClientClosed
	}

	n, err := client.ReadWithTimeout(buf, timeout)
	if err != nil && rc.shouldReconnect(err) {
		if reconnErr := rc.reconnect(); reconnErr == nil {
			// Retry after successful reconnection
			rc.mu.RLock()
			client = rc.client
			rc.mu.RUnlock()
			return client.ReadWithTimeout(buf, timeout)
		}
	}
	return n, err
}

func (rc *reconnectClient) RemoteAddr() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if rc.client != nil {
		return rc.client.RemoteAddr()
	}
	return ""
}

func (rc *reconnectClient) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.closed {
		return nil
	}
	rc.closed = true

	if rc.client != nil {
		return rc.client.Close()
	}
	return nil
}

// shouldReconnect determines if an error should trigger reconnection
func (rc *reconnectClient) shouldReconnect(err error) bool {
	if err == nil {
		return false
	}
	// Check for common disconnection errors
	if errors.Is(err, io.EOF) || errors.Is(err, ErrClientClosed) {
		return true
	}
	// Check error message for common patterns
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "use of closed")
}
