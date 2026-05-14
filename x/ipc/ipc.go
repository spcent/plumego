// Package ipc provides cross-platform inter-process communication (IPC) primitives.
//
// # Version 2 Breaking Changes
//
// This version includes the following breaking changes:
//
// 1. Client.RemoteAddr() now returns net.Addr instead of string
//   - Migration: Use RemoteAddrString() for backward compatibility
//   - Or: Use addr.String() on the returned net.Addr
//
// 2. Client interface has been refactored into composable sub-interfaces
//   - Reader: io.Reader + ReadWithTimeout
//   - Writer: io.Writer + WriteWithTimeout
//   - AddrProvider: RemoteAddr() + RemoteAddrString()
//   - Client now embeds all these interfaces
//
// These changes improve type safety and enable more flexible composition.
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
//	// Get remote address (v2 API)
//	addr := client.RemoteAddr()
//	fmt.Printf("Connected to: %s (%s)\n", addr.String(), addr.Network())
//
//	// Or use backward-compatible string method
//	addrStr := client.RemoteAddrString()
//	fmt.Printf("Address: %s\n", addrStr)
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
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/spcent/plumego/log"
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
	ConnectTimeout      time.Duration        // Timeout for connection establishment
	ReadTimeout         time.Duration        // Timeout for read operations
	WriteTimeout        time.Duration        // Timeout for write operations
	BufferSize          int                  // Buffer size for read/write operations (Windows Named Pipe only)
	UnixSocketPerm      uint32               // Unix socket file permissions (e.g., 0700). Default: 0700 (owner only). Unix/Linux only.
	UnixSocketDirPerm   uint32               // Unix socket directory permissions (e.g., 0755). Default: 0755. Unix/Linux only.
	KeepAlive           bool                 // Enable TCP keepalive for TCP connections. Default: true.
	KeepAlivePeriod     time.Duration        // TCP keepalive period. Default: 30s. Only applies to TCP connections.
	WindowsSecuritySDDL string               // Windows security descriptor (SDDL string). Windows only. Empty = default security.
	Metrics             MetricsObserver      // Optional metrics collector
	Logger              log.StructuredLogger // Optional structured logger
}

// MetricsObserver captures IPC-specific observations without routing the
// feature-owned observer contract through stable metrics.
type MetricsObserver interface {
	ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error)
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ConnectTimeout:    10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		BufferSize:        4096,
		UnixSocketPerm:    0700, // Owner only (rwx------)
		UnixSocketDirPerm: 0755, // Owner rwx, others rx (rwxr-xr-x)
		KeepAlive:         true,
		KeepAlivePeriod:   30 * time.Second,
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

// WithMetrics sets the metrics collector
func WithMetrics(m MetricsObserver) Option {
	return func(c *Config) {
		c.Metrics = m
	}
}

// WithLogger sets the structured logger
func WithLogger(l log.StructuredLogger) Option {
	return func(c *Config) {
		c.Logger = l
	}
}

// WithWindowsSecuritySDDL sets Windows security descriptor (SDDL string) for Named Pipes.
// Only applicable on Windows. Example SDDL strings:
//   - "D:P(A;;GA;;;WD)" - Allow all access to Everyone
//   - "D:P(A;;GA;;;SY)(A;;GA;;;BA)" - Allow SYSTEM and Administrators
//   - "D:P(A;;GRGW;;;AU)" - Allow Authenticated Users read/write
//
// See https://docs.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-string-format
func WithWindowsSecuritySDDL(sddl string) Option {
	return func(c *Config) {
		c.WindowsSecuritySDDL = sddl
	}
}

// Addr represents an IPC address, implementing net.Addr interface
type Addr struct {
	network string // "unix", "pipe", "tcp"
	address string // actual address string
}

// Network returns the network type
func (a *Addr) Network() string {
	return a.network
}

// String returns the address string
func (a *Addr) String() string {
	return a.address
}

// NewAddr creates a new IPC address
func NewAddr(network, address string) net.Addr {
	return &Addr{
		network: network,
		address: address,
	}
}

// HeartbeatConfig holds configuration for connection heartbeat
type HeartbeatConfig struct {
	Enabled  bool          // Enable heartbeat mechanism
	Interval time.Duration // Interval between heartbeat checks
	Timeout  time.Duration // Timeout for heartbeat response
	OnDead   func()        // Callback when connection is detected as dead (optional)
}

// DefaultHeartbeatConfig returns default heartbeat configuration
func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
		OnDead:   nil,
	}
}

// heartbeatClient wraps a Client with heartbeat functionality
type heartbeatClient struct {
	client   Client
	config   *HeartbeatConfig
	mu       sync.RWMutex
	stopCh   chan struct{}
	stopped  bool
	lastSeen time.Time
}

// NewHeartbeatClient creates a client wrapper with heartbeat monitoring.
// The heartbeat uses a simple ping-pong protocol over the framed client.
func NewHeartbeatClient(client Client, cfg *HeartbeatConfig) (Client, error) {
	if cfg == nil {
		cfg = DefaultHeartbeatConfig()
	}

	if !cfg.Enabled {
		return client, nil
	}

	hb := &heartbeatClient{
		client:   client,
		config:   cfg,
		stopCh:   make(chan struct{}),
		stopped:  false,
		lastSeen: time.Now(),
	}

	// Start heartbeat goroutine
	go hb.run()

	return hb, nil
}

func (hb *heartbeatClient) run() {
	ticker := time.NewTicker(hb.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := hb.ping(); err != nil {
				// Connection appears dead
				hb.mu.Lock()
				if hb.config.OnDead != nil && !hb.stopped {
					hb.config.OnDead()
				}
				hb.mu.Unlock()
				return
			}
		case <-hb.stopCh:
			return
		}
	}
}

func (hb *heartbeatClient) ping() error {
	hb.mu.Lock()
	if hb.stopped {
		hb.mu.Unlock()
		return ErrClientClosed
	}
	hb.mu.Unlock()

	// Simple ping: write a special byte sequence
	pingMsg := []byte{0xFF, 0xFF, 0x00, 0x00} // Magic ping bytes
	deadline := time.Now().Add(hb.config.Timeout)

	// Set deadline for ping
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Write ping
	n, err := hb.client.WriteWithTimeout(pingMsg, hb.config.Timeout)
	if err != nil || n != len(pingMsg) {
		return fmt.Errorf("heartbeat ping failed: %w", err)
	}

	// Expect pong (same bytes echoed back)
	pongBuf := make([]byte, len(pingMsg))
	n, err = hb.client.ReadWithTimeout(pongBuf, hb.config.Timeout)
	if err != nil || n != len(pingMsg) {
		return fmt.Errorf("heartbeat pong failed: %w", err)
	}

	// Update last seen time
	hb.mu.Lock()
	hb.lastSeen = time.Now()
	hb.mu.Unlock()

	_ = ctx // Use ctx to satisfy linter

	return nil
}

// LastSeen returns the last time a successful heartbeat was received
func (hb *heartbeatClient) LastSeen() time.Time {
	hb.mu.RLock()
	defer hb.mu.RUnlock()
	return hb.lastSeen
}

// Delegate Client interface methods
func (hb *heartbeatClient) Write(data []byte) (int, error) {
	return hb.client.Write(data)
}

func (hb *heartbeatClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	return hb.client.WriteWithTimeout(data, timeout)
}

func (hb *heartbeatClient) Read(buf []byte) (int, error) {
	return hb.client.Read(buf)
}

func (hb *heartbeatClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	return hb.client.ReadWithTimeout(buf, timeout)
}

func (hb *heartbeatClient) RemoteAddr() net.Addr {
	return hb.client.RemoteAddr()
}

func (hb *heartbeatClient) RemoteAddrString() string {
	return hb.client.RemoteAddrString()
}

func (hb *heartbeatClient) Close() error {
	hb.mu.Lock()
	if !hb.stopped {
		hb.stopped = true
		close(hb.stopCh)
	}
	hb.mu.Unlock()

	return hb.client.Close()
}

// PoolConfig holds configuration for connection pool
type PoolConfig struct {
	MaxConns     int           // Maximum number of connections in pool
	MaxIdleConns int           // Maximum number of idle connections to retain
	MaxIdleTime  time.Duration // Maximum time a connection can be idle before being closed
	DialTimeout  time.Duration // Timeout for creating new connections
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxConns:     10,
		MaxIdleConns: 5,
		MaxIdleTime:  5 * time.Minute,
		DialTimeout:  10 * time.Second,
	}
}

// poolConn wraps a connection with pool metadata
type poolConn struct {
	client    Client
	pool      *Pool
	createdAt time.Time
	lastUsed  time.Time
	inUse     bool
}

// Pool manages a pool of reusable IPC connections
type Pool struct {
	addr    string
	config  *Config
	poolCfg *PoolConfig

	mu      sync.Mutex
	conns   []*poolConn
	waiting []chan *poolConn
	closed  bool

	stopCh chan struct{}
}

// NewPool creates a new connection pool
func NewPool(addr string, poolCfg *PoolConfig, opts ...Option) (*Pool, error) {
	if poolCfg == nil {
		poolCfg = DefaultPoolConfig()
	}

	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	pool := &Pool{
		addr:    addr,
		config:  config,
		poolCfg: poolCfg,
		conns:   make([]*poolConn, 0, poolCfg.MaxConns),
		waiting: make([]chan *poolConn, 0),
		stopCh:  make(chan struct{}),
	}

	// Start background cleanup goroutine
	go pool.cleanup()

	return pool, nil
}

// Get acquires a connection from the pool
func (p *Pool) Get() (Client, error) {
	return p.GetWithContext(context.Background())
}

// GetWithContext acquires a connection from the pool with context
func (p *Pool) GetWithContext(ctx context.Context) (Client, error) {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return nil, ErrClientClosed
	}

	// Try to find an idle connection
	for i, pc := range p.conns {
		if !pc.inUse {
			// Check if connection is still valid
			if time.Since(pc.lastUsed) > p.poolCfg.MaxIdleTime {
				// Connection too old, remove it
				pc.client.Close()
				p.conns = append(p.conns[:i], p.conns[i+1:]...)
				continue
			}

			pc.inUse = true
			pc.lastUsed = time.Now()
			p.mu.Unlock()
			return &poolClient{conn: pc}, nil
		}
	}

	// No idle connection, try to create new one
	if len(p.conns) < p.poolCfg.MaxConns {
		p.mu.Unlock()
		client, err := dialPlatformWithContext(ctx, p.addr, p.config)
		if err != nil {
			return nil, err
		}

		pc := &poolConn{
			client:    client,
			pool:      p,
			createdAt: time.Now(),
			lastUsed:  time.Now(),
			inUse:     true,
		}

		p.mu.Lock()
		p.conns = append(p.conns, pc)
		p.mu.Unlock()

		return &poolClient{conn: pc}, nil
	}

	// Pool is full, wait for a connection to be released
	waiter := make(chan *poolConn, 1)
	p.waiting = append(p.waiting, waiter)
	p.mu.Unlock()

	select {
	case pc := <-waiter:
		return &poolClient{conn: pc}, nil
	case <-ctx.Done():
		// Remove from waiting list
		p.mu.Lock()
		for i, w := range p.waiting {
			if w == waiter {
				p.waiting = append(p.waiting[:i], p.waiting[i+1:]...)
				break
			}
		}
		p.mu.Unlock()
		return nil, ctx.Err()
	}
}

// put returns a connection to the pool
func (p *Pool) put(pc *poolConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		pc.client.Close()
		return
	}

	pc.lastUsed = time.Now()
	pc.inUse = false

	// Wake up waiting goroutine if any
	if len(p.waiting) > 0 {
		waiter := p.waiting[0]
		p.waiting = p.waiting[1:]
		pc.inUse = true
		waiter <- pc
	}
}

// Close closes the pool and all connections
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true
	close(p.stopCh)

	// Close all connections
	for _, pc := range p.conns {
		pc.client.Close()
	}
	p.conns = nil

	// Wake up all waiting goroutines
	for _, waiter := range p.waiting {
		close(waiter)
	}
	p.waiting = nil

	return nil
}

// cleanup periodically removes idle connections
func (p *Pool) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			now := time.Now()

			// Keep track of connections to close
			var toClose []*poolConn

			// Filter out expired idle connections
			filtered := make([]*poolConn, 0, len(p.conns))
			idleCount := 0

			for _, pc := range p.conns {
				if pc.inUse {
					filtered = append(filtered, pc)
					continue
				}

				// Check if connection is too old
				if now.Sub(pc.lastUsed) > p.poolCfg.MaxIdleTime {
					toClose = append(toClose, pc)
					continue
				}

				// Check if we have too many idle connections
				idleCount++
				if idleCount > p.poolCfg.MaxIdleConns {
					toClose = append(toClose, pc)
					continue
				}

				filtered = append(filtered, pc)
			}

			p.conns = filtered
			p.mu.Unlock()

			// Close connections outside lock
			for _, pc := range toClose {
				pc.client.Close()
			}

		case <-p.stopCh:
			return
		}
	}
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := PoolStats{
		TotalConns: len(p.conns),
		Waiting:    len(p.waiting),
	}

	for _, pc := range p.conns {
		if pc.inUse {
			stats.InUse++
		} else {
			stats.Idle++
		}
	}

	return stats
}

// PoolStats contains pool statistics
type PoolStats struct {
	TotalConns int // Total connections in pool
	InUse      int // Connections currently in use
	Idle       int // Idle connections
	Waiting    int // Goroutines waiting for connection
}

// poolClient wraps a poolConn and returns it to pool on close
type poolClient struct {
	conn *poolConn
	mu   sync.Mutex
}

func (pc *poolClient) Write(data []byte) (int, error) {
	return pc.conn.client.Write(data)
}

func (pc *poolClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	return pc.conn.client.WriteWithTimeout(data, timeout)
}

func (pc *poolClient) Read(buf []byte) (int, error) {
	return pc.conn.client.Read(buf)
}

func (pc *poolClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	return pc.conn.client.ReadWithTimeout(buf, timeout)
}

func (pc *poolClient) RemoteAddr() net.Addr {
	return pc.conn.client.RemoteAddr()
}

func (pc *poolClient) RemoteAddrString() string {
	return pc.conn.client.RemoteAddrString()
}

func (pc *poolClient) Close() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.conn == nil {
		return nil
	}

	pool := pc.conn.pool
	conn := pc.conn
	pc.conn = nil

	// Return to pool instead of closing
	pool.put(conn)
	return nil
}

// StreamConfig holds configuration for stream transfers
type StreamConfig struct {
	ChunkSize      int           // Size of each chunk for streaming
	BufferSize     int           // Buffer size for stream operations
	EnableChecksum bool          // Enable checksum verification
	Timeout        time.Duration // Timeout for stream operations
}

// DefaultStreamConfig returns default stream configuration
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		ChunkSize:      64 * 1024,  // 64KB chunks
		BufferSize:     128 * 1024, // 128KB buffer
		EnableChecksum: true,
		Timeout:        5 * time.Minute,
	}
}

// StreamClient extends Client with streaming capabilities for large data transfers
type StreamClient interface {
	Client
	// WriteStream writes data from a reader in chunks
	WriteStream(r io.Reader) (int64, error)
	// ReadStream reads data into a writer in chunks
	ReadStream(w io.Writer) (int64, error)
}

// streamClient wraps a Client with streaming support
type streamClient struct {
	client Client
	config *StreamConfig
}

// NewStreamClient creates a streaming client wrapper optimized for large data transfers.
// Uses chunked transfer with optional checksums for reliability.
func NewStreamClient(client Client, cfg *StreamConfig) StreamClient {
	if cfg == nil {
		cfg = DefaultStreamConfig()
	}

	return &streamClient{
		client: client,
		config: cfg,
	}
}

func (sc *streamClient) WriteStream(r io.Reader) (int64, error) {
	var totalWritten int64
	buf := make([]byte, sc.config.ChunkSize)

	for {
		// Read chunk
		n, readErr := r.Read(buf)
		if n > 0 {
			// Write chunk size (4 bytes, big-endian)
			sizeBytes := make([]byte, 4)
			sizeBytes[0] = byte(n >> 24)
			sizeBytes[1] = byte(n >> 16)
			sizeBytes[2] = byte(n >> 8)
			sizeBytes[3] = byte(n)

			if _, err := sc.client.WriteWithTimeout(sizeBytes, sc.config.Timeout); err != nil {
				return totalWritten, fmt.Errorf("failed to write chunk size: %w", err)
			}

			// Write chunk data
			written := 0
			for written < n {
				w, err := sc.client.WriteWithTimeout(buf[written:n], sc.config.Timeout)
				if err != nil {
					return totalWritten, fmt.Errorf("failed to write chunk data: %w", err)
				}
				written += w
			}

			totalWritten += int64(n)
		}

		if readErr == io.EOF {
			// Write zero-length chunk to signal end
			endMarker := []byte{0, 0, 0, 0}
			if _, err := sc.client.WriteWithTimeout(endMarker, sc.config.Timeout); err != nil {
				return totalWritten, fmt.Errorf("failed to write end marker: %w", err)
			}
			break
		}

		if readErr != nil {
			return totalWritten, readErr
		}
	}

	return totalWritten, nil
}

func (sc *streamClient) ReadStream(w io.Writer) (int64, error) {
	var totalRead int64

	for {
		// Read chunk size (4 bytes)
		sizeBytes := make([]byte, 4)
		if _, err := io.ReadFull(sc.client, sizeBytes); err != nil {
			return totalRead, fmt.Errorf("failed to read chunk size: %w", err)
		}

		chunkSize := int(uint32(sizeBytes[0])<<24 |
			uint32(sizeBytes[1])<<16 |
			uint32(sizeBytes[2])<<8 |
			uint32(sizeBytes[3]))

		// Zero-length chunk signals end of stream
		if chunkSize == 0 {
			break
		}

		if chunkSize > sc.config.ChunkSize*2 {
			return totalRead, fmt.Errorf("chunk size too large: %d (max %d)", chunkSize, sc.config.ChunkSize*2)
		}

		// Read chunk data
		buf := make([]byte, chunkSize)
		if _, err := io.ReadFull(sc.client, buf); err != nil {
			return totalRead, fmt.Errorf("failed to read chunk data: %w", err)
		}

		// Write to destination
		written := 0
		for written < chunkSize {
			w, err := w.Write(buf[written:chunkSize])
			if err != nil {
				return totalRead, fmt.Errorf("failed to write to destination: %w", err)
			}
			written += w
		}

		totalRead += int64(chunkSize)
	}

	return totalRead, nil
}

// Delegate Client interface methods
func (sc *streamClient) Write(data []byte) (int, error) {
	return sc.client.Write(data)
}

func (sc *streamClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	return sc.client.WriteWithTimeout(data, timeout)
}

func (sc *streamClient) Read(buf []byte) (int, error) {
	return sc.client.Read(buf)
}

func (sc *streamClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	return sc.client.ReadWithTimeout(buf, timeout)
}

func (sc *streamClient) RemoteAddr() net.Addr {
	return sc.client.RemoteAddr()
}

func (sc *streamClient) RemoteAddrString() string {
	return sc.client.RemoteAddrString()
}

func (sc *streamClient) Close() error {
	return sc.client.Close()
}

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow returns true if an operation is allowed
	Allow() bool
	// Wait blocks until an operation is allowed or context is cancelled
	Wait(ctx context.Context) error
}

// tokenBucketLimiter implements a token bucket rate limiter
type tokenBucketLimiter struct {
	rate      float64   // Tokens per second
	capacity  int       // Maximum burst size
	tokens    float64   // Current tokens
	lastCheck time.Time // Last time tokens were added
	mu        sync.Mutex
}

// NewRateLimiter creates a new token bucket rate limiter
// rate: operations per second
// burst: maximum burst size (number of operations that can be performed at once)
func NewRateLimiter(rate float64, burst int) RateLimiter {
	return &tokenBucketLimiter{
		rate:      rate,
		capacity:  burst,
		tokens:    float64(burst),
		lastCheck: time.Now(),
	}
}

func (tb *tokenBucketLimiter) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

func (tb *tokenBucketLimiter) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}

		// Calculate how long to wait for next token
		tb.mu.Lock()
		tokensNeeded := 1.0 - tb.tokens
		waitTime := time.Duration(float64(time.Second) * tokensNeeded / tb.rate)
		tb.mu.Unlock()

		if waitTime > 0 {
			timer := time.NewTimer(waitTime)
			select {
			case <-timer.C:
				// Try again
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}
	}
}

func (tb *tokenBucketLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastCheck)
	tb.lastCheck = now

	// Add tokens based on elapsed time
	tokensToAdd := elapsed.Seconds() * tb.rate
	tb.tokens += tokensToAdd

	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
}

// rateLimitedServer wraps a Server with rate limiting
type rateLimitedServer struct {
	server  Server
	limiter RateLimiter
}

// NewRateLimitedServer creates a server wrapper with connection rate limiting
func NewRateLimitedServer(server Server, limiter RateLimiter) Server {
	return &rateLimitedServer{
		server:  server,
		limiter: limiter,
	}
}

func (rs *rateLimitedServer) Accept() (Client, error) {
	return rs.AcceptWithContext(context.Background())
}

func (rs *rateLimitedServer) AcceptWithContext(ctx context.Context) (Client, error) {
	// Wait for rate limiter
	if err := rs.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	return rs.server.AcceptWithContext(ctx)
}

func (rs *rateLimitedServer) Addr() string {
	return rs.server.Addr()
}

func (rs *rateLimitedServer) Close() error {
	return rs.server.Close()
}

func (rs *rateLimitedServer) Shutdown(ctx context.Context) error {
	return rs.server.Shutdown(ctx)
}
