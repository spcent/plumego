//go:build !windows

package ipc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// unixServer implements Server for Unix-like systems
type unixServer struct {
	listener   net.Listener // The underlying listener
	socketPath string       // The path to the Unix domain socket file (if applicable)
	config     *Config      // The configuration for the server
	mu         sync.RWMutex // Mutex for thread-safe access to server state
	closed     bool         // Indicates if the server is closed
}

// unixClient implements Client for Unix-like systems
type unixClient struct {
	conn   net.Conn     // The underlying connection
	config *Config      // The configuration for the client
	mu     sync.RWMutex // Mutex for thread-safe access to client state
	closed bool         // Indicates if the client is closed
}

func newPlatformServer(addr string, config *Config) (Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var listener net.Listener
	var socketPath string
	var err error

	// Try Unix domain socket first
	if !strings.Contains(addr, ":") {
		// Ensure directory exists
		dir := filepath.Dir(addr)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// Remove existing socket file
		if err = os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		listener, err = net.Listen("unix", addr)
		if err == nil {
			socketPath = addr
		}
	}

	// Fallback to TCP if Unix socket fails or addr contains port
	if listener == nil {
		tcpAddr := addr
		if !strings.Contains(addr, ":") {
			tcpAddr = "127.0.0.1:0" // Auto-assign port
		}
		listener, err = net.Listen("tcp", tcpAddr)
		if err != nil {
			return nil, err
		}
	}

	server := &unixServer{
		listener:   listener,
		socketPath: socketPath,
		config:     config,
	}

	return server, nil
}

func (s *unixServer) Accept() (Client, error) {
	return s.AcceptWithContext(context.Background())
}

func (s *unixServer) AcceptWithContext(ctx context.Context) (Client, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, net.ErrClosed
	}
	s.mu.RUnlock()

	// Set deadline if context has timeout
	if deadline, ok := ctx.Deadline(); ok {
		// Use UnixListener's SetDeadline method with correct type assertion
		if err := s.listener.(*net.UnixListener).SetDeadline(deadline); err != nil {
			return nil, err
		}
	}

	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}

	client := &unixClient{
		conn:   conn,
		config: s.config,
	}

	return client, nil
}

func (s *unixServer) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

func (s *unixServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}

	// Clean up socket file
	if s.socketPath != "" {
		if removeErr := os.Remove(s.socketPath); removeErr != nil && !os.IsNotExist(removeErr) {
			if err == nil {
				err = removeErr
			}
		}
	}

	return err
}

func dialPlatform(addr string, config *Config) (Client, error) {
	return dialPlatformWithContext(context.Background(), addr, config)
}

func dialPlatformWithContext(ctx context.Context, addr string, config *Config) (Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var conn net.Conn
	var err error

	dialer := &net.Dialer{
		Timeout: config.ConnectTimeout,
	}

	// Try Unix domain socket first if no port specified
	if !strings.Contains(addr, ":") {
		conn, err = dialer.DialContext(ctx, "unix", addr)
	}

	// Fallback to TCP if Unix socket fails
	if conn == nil {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}

	if err != nil {
		return nil, err
	}

	client := &unixClient{
		conn:   conn,
		config: config,
	}

	return client, nil
}

func (c *unixClient) Write(data []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.conn == nil {
		return 0, net.ErrClosed
	}

	return c.conn.Write(data)
}

func (c *unixClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.conn == nil {
		return 0, net.ErrClosed
	}

	if timeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(timeout))
		defer c.conn.SetWriteDeadline(time.Time{})
	}

	return c.conn.Write(data)
}

func (c *unixClient) Read(buf []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.conn == nil {
		return 0, net.ErrClosed
	}

	return c.conn.Read(buf)
}

func (c *unixClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed || c.conn == nil {
		return 0, net.ErrClosed
	}

	if timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(timeout))
		defer c.conn.SetReadDeadline(time.Time{})
	}

	return c.conn.Read(buf)
}

func (c *unixClient) RemoteAddr() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}

func (c *unixClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
