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
	KeepAlive        bool          // Enable TCP keepalive for TCP connections. Default: true.
	KeepAlivePeriod  time.Duration // TCP keepalive period. Default: 30s. Only applies to TCP connections.
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ConnectTimeout:   10 * time.Second,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		BufferSize:       4096,
		UnixSocketPerm:   0700,  // Owner only (rwx------)
		UnixSocketDirPerm: 0755, // Owner rwx, others rx (rwxr-xr-x)
		KeepAlive:        true,
		KeepAlivePeriod:  30 * time.Second,
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

// WithKeepAlive enables or disables TCP keepalive
func WithKeepAlive(enable bool) Option {
	return func(c *Config) {
		c.KeepAlive = enable
	}
}

// WithKeepAlivePeriod sets the TCP keepalive period
func WithKeepAlivePeriod(period time.Duration) Option {
	return func(c *Config) {
		c.KeepAlivePeriod = period
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
	// Close closes the server immediately, releasing any resources
	Close() error
	// Shutdown gracefully shuts down the server without interrupting active connections.
	// It first stops accepting new connections, then waits for existing Accept calls
	// to complete or for the context to be cancelled.
	Shutdown(ctx context.Context) error
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

// FramedClient extends Client with message framing support using length-prefix protocol.
// Messages are encoded with a 4-byte length prefix (big-endian) followed by the payload.
type FramedClient interface {
	Client
	// WriteMessage writes a complete message with length prefix
	WriteMessage(msg []byte) error
	// WriteMessageWithTimeout writes a message with timeout
	WriteMessageWithTimeout(msg []byte, timeout time.Duration) error
	// ReadMessage reads a complete message
	ReadMessage() ([]byte, error)
	// ReadMessageWithTimeout reads a message with timeout
	ReadMessageWithTimeout(timeout time.Duration) ([]byte, error)
}

// framedClient wraps a Client with message framing support
type framedClient struct {
	client Client
	mu     sync.Mutex
}

// NewFramedClient creates a framed client wrapper with length-prefix protocol.
// The protocol uses 4 bytes (uint32, big-endian) for message length, followed by the payload.
// Maximum message size is 16MB.
func NewFramedClient(client Client) FramedClient {
	return &framedClient{
		client: client,
	}
}

const (
	maxFrameSize = 16 * 1024 * 1024 // 16MB max message size
	frameLenSize = 4                // 4 bytes for uint32 length
)

func (fc *framedClient) WriteMessage(msg []byte) error {
	return fc.WriteMessageWithTimeout(msg, 0)
}

func (fc *framedClient) WriteMessageWithTimeout(msg []byte, timeout time.Duration) error {
	if len(msg) > maxFrameSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(msg), maxFrameSize)
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Encode length prefix (4 bytes, big-endian)
	lengthPrefix := make([]byte, frameLenSize)
	lengthPrefix[0] = byte(len(msg) >> 24)
	lengthPrefix[1] = byte(len(msg) >> 16)
	lengthPrefix[2] = byte(len(msg) >> 8)
	lengthPrefix[3] = byte(len(msg))

	// Write length prefix
	var n int
	var err error
	if timeout > 0 {
		n, err = fc.client.WriteWithTimeout(lengthPrefix, timeout)
	} else {
		n, err = fc.client.Write(lengthPrefix)
	}
	if err != nil {
		return err
	}
	if n != frameLenSize {
		return fmt.Errorf("failed to write frame length: wrote %d bytes, expected %d", n, frameLenSize)
	}

	// Write message payload
	if timeout > 0 {
		n, err = fc.client.WriteWithTimeout(msg, timeout)
	} else {
		n, err = fc.client.Write(msg)
	}
	if err != nil {
		return err
	}
	if n != len(msg) {
		return fmt.Errorf("failed to write message: wrote %d bytes, expected %d", n, len(msg))
	}

	return nil
}

func (fc *framedClient) ReadMessage() ([]byte, error) {
	return fc.ReadMessageWithTimeout(0)
}

func (fc *framedClient) ReadMessageWithTimeout(timeout time.Duration) ([]byte, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Read length prefix (4 bytes)
	lengthPrefix := make([]byte, frameLenSize)
	var n int
	var err error
	if timeout > 0 {
		n, err = io.ReadFull(timeoutReader{fc.client, timeout}, lengthPrefix)
	} else {
		n, err = io.ReadFull(fc.client, lengthPrefix)
	}
	if err != nil {
		return nil, err
	}
	if n != frameLenSize {
		return nil, fmt.Errorf("failed to read frame length: read %d bytes, expected %d", n, frameLenSize)
	}

	// Decode message length (big-endian)
	msgLen := uint32(lengthPrefix[0])<<24 |
		uint32(lengthPrefix[1])<<16 |
		uint32(lengthPrefix[2])<<8 |
		uint32(lengthPrefix[3])

	if msgLen > maxFrameSize {
		return nil, fmt.Errorf("message too large: %d bytes (max %d)", msgLen, maxFrameSize)
	}

	if msgLen == 0 {
		return []byte{}, nil
	}

	// Read message payload
	msg := make([]byte, msgLen)
	if timeout > 0 {
		n, err = io.ReadFull(timeoutReader{fc.client, timeout}, msg)
	} else {
		n, err = io.ReadFull(fc.client, msg)
	}
	if err != nil {
		return nil, err
	}
	if n != int(msgLen) {
		return nil, fmt.Errorf("failed to read message: read %d bytes, expected %d", n, msgLen)
	}

	return msg, nil
}

// Delegate Client interface methods to underlying client
func (fc *framedClient) Write(data []byte) (int, error) {
	return fc.client.Write(data)
}

func (fc *framedClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	return fc.client.WriteWithTimeout(data, timeout)
}

func (fc *framedClient) Read(buf []byte) (int, error) {
	return fc.client.Read(buf)
}

func (fc *framedClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	return fc.client.ReadWithTimeout(buf, timeout)
}

func (fc *framedClient) RemoteAddr() string {
	return fc.client.RemoteAddr()
}

func (fc *framedClient) Close() error {
	return fc.client.Close()
}

// timeoutReader implements io.Reader with timeout support
type timeoutReader struct {
	client  Client
	timeout time.Duration
}

func (tr timeoutReader) Read(p []byte) (int, error) {
	return tr.client.ReadWithTimeout(p, tr.timeout)
}
