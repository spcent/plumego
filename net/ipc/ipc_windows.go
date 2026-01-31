// ipc_windows.go
//go:build windows

package ipc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procCreateNamedPipeW    = kernel32.NewProc("CreateNamedPipeW")
	procConnectNamedPipe    = kernel32.NewProc("ConnectNamedPipe")
	procDisconnectNamedPipe = kernel32.NewProc("DisconnectNamedPipe")
	procWaitNamedPipeW      = kernel32.NewProc("WaitNamedPipeW")
)

const (
	PIPE_ACCESS_DUPLEX       = 0x3
	PIPE_TYPE_BYTE           = 0x0
	PIPE_READMODE_BYTE       = 0x0
	PIPE_WAIT                = 0x0
	PIPE_UNLIMITED_INSTANCES = 255
	GENERIC_READ             = 0x80000000
	GENERIC_WRITE            = 0x40000000
	OPEN_EXISTING            = 3
	INVALID_HANDLE_VALUE     = ^uintptr(0)
	errorPipeConnected       = syscall.Errno(535)
	errorPipeClosing         = syscall.Errno(232)
)

// winServer implements Server for Windows using Named Pipes
type winServer struct {
	pipeName string
	config   *Config
	mu       sync.RWMutex
	closed   bool
	next     syscall.Handle
}

// winClient implements Client for Windows
type winClient struct {
	conn   net.Conn
	handle syscall.Handle
	config *Config
	mu     sync.RWMutex
	closed bool
}

func newPlatformServer(addr string, config *Config) (Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Try named pipe first for local communication
	if isNamedPipeAddr(addr) || !strings.Contains(addr, ":") {
		pipeName := normalizePipeAddr(addr)
		handle, err := createNamedPipe(pipeName, config.BufferSize)
		if err != nil {
			return nil, err
		}
		server := &winServer{
			pipeName: pipeName,
			config:   config,
			next:     handle,
		}
		return server, nil
	}

	// Fallback to TCP for network communication
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &tcpServer{
		listener: listener,
		config:   config,
	}, nil
}

// tcpServer for TCP fallback on Windows
type tcpServer struct {
	listener net.Listener
	config   *Config
	mu       sync.RWMutex
	closed   bool
}

func (s *tcpServer) Accept() (Client, error) {
	return s.AcceptWithContext(context.Background())
}

func (s *tcpServer) AcceptWithContext(ctx context.Context) (Client, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, ErrServerClosed
	}
	s.mu.RUnlock()

	if deadline, ok := ctx.Deadline(); ok {
		if listener, ok := s.listener.(*net.TCPListener); ok {
			if err := listener.SetDeadline(deadline); err != nil {
				return nil, err
			}
			defer listener.SetDeadline(time.Time{})
		}
	}

	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}

	// Enable keepalive for TCP connections
	if tcpConn, ok := conn.(*net.TCPConn); ok && s.config.KeepAlive {
		if err := tcpConn.SetKeepAlive(true); err == nil {
			tcpConn.SetKeepAlivePeriod(s.config.KeepAlivePeriod)
		}
	}

	return &winClient{
		conn:   conn,
		config: s.config,
	}, nil
}

func (s *tcpServer) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

func (s *tcpServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *tcpServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}

	// Wait for context cancellation (graceful period)
	<-ctx.Done()

	return err
}

func (s *winServer) Accept() (Client, error) {
	return s.AcceptWithContext(context.Background())
}

func (s *winServer) AcceptWithContext(ctx context.Context) (Client, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, ErrServerClosed
	}
	handle := s.next
	s.next = 0
	s.mu.Unlock()

	if handle == 0 {
		var err error
		handle, err = createNamedPipe(s.pipeName, s.config.BufferSize)
		if err != nil {
			return nil, err
		}
	}

	for {
		if err := connectNamedPipeWithContext(ctx, handle); err != nil {
			if err == errorPipeClosing {
				go syscall.CloseHandle(handle)
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				handle, err = createNamedPipe(s.pipeName, s.config.BufferSize)
				if err != nil {
					return nil, err
				}
				continue
			}
			go syscall.CloseHandle(handle)
			return nil, err
		}
		break
	}

	// Asynchronously prepare the next handle to reduce Accept latency
	go s.prepareNextHandle()

	return &winClient{
		handle: handle,
		config: s.config,
	}, nil
}

func (s *winServer) Addr() string {
	return s.pipeName
}

func (s *winServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	if s.next != 0 {
		_ = syscall.CloseHandle(s.next)
		s.next = 0
	}
	return nil
}

func (s *winServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	if s.next != 0 {
		_ = syscall.CloseHandle(s.next)
		s.next = 0
	}
	s.mu.Unlock()

	// Wait for context cancellation (graceful period)
	<-ctx.Done()

	return nil
}

func dialPlatform(addr string, config *Config) (Client, error) {
	return dialPlatformWithContext(context.Background(), addr, config)
}

func dialPlatformWithContext(ctx context.Context, addr string, config *Config) (Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Try named pipe first for local communication
	if isNamedPipeAddr(addr) || !strings.Contains(addr, ":") {
		pipeName := normalizePipeAddr(addr)
		handle, err := dialNamedPipe(ctx, pipeName, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		return &winClient{
			handle: handle,
			config: config,
		}, nil
	}

	// Fallback to TCP
	dialer := &net.Dialer{
		Timeout:   config.ConnectTimeout,
		KeepAlive: config.KeepAlivePeriod,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	// Enable keepalive for TCP connections
	if tcpConn, ok := conn.(*net.TCPConn); ok && config.KeepAlive {
		if setErr := tcpConn.SetKeepAlive(true); setErr == nil {
			tcpConn.SetKeepAlivePeriod(config.KeepAlivePeriod)
		}
	}

	return &winClient{
		conn:   conn,
		config: config,
	}, nil
}

func createNamedPipe(pipeName string, bufSize int) (syscall.Handle, error) {
	pipeNamePtr, err := syscall.UTF16PtrFromString(pipeName)
	if err != nil {
		return 0, err
	}

	handle, _, _ := procCreateNamedPipeW.Call(
		uintptr(unsafe.Pointer(pipeNamePtr)),
		PIPE_ACCESS_DUPLEX,
		PIPE_TYPE_BYTE|PIPE_READMODE_BYTE|PIPE_WAIT,
		PIPE_UNLIMITED_INSTANCES,
		uintptr(bufSize),
		uintptr(bufSize),
		0,
		0,
	)

	if handle == INVALID_HANDLE_VALUE {
		return 0, fmt.Errorf("failed to create named pipe")
	}

	return syscall.Handle(handle), nil
}

// connectNamedPipeWithContext attempts to connect a named pipe with context cancellation.
// Note: On Windows, ConnectNamedPipe is a blocking system call that cannot be truly
// interrupted. If the context is cancelled, we close the handle which will eventually
// unblock the system call, but there may be a brief delay. The goroutine will complete
// once the system call returns.
func connectNamedPipeWithContext(ctx context.Context, handle syscall.Handle) error {
	type result struct {
		err    error
		handle syscall.Handle
	}
	done := make(chan result, 1)

	go func() {
		ret, _, err := procConnectNamedPipe.Call(uintptr(handle), 0)
		if ret == 0 {
			if err == errorPipeConnected {
				done <- result{err: nil, handle: handle}
				return
			}
			if err != nil {
				done <- result{err: err, handle: handle}
				return
			}
			done <- result{err: fmt.Errorf("failed to connect named pipe"), handle: handle}
			return
		}
		done <- result{err: nil, handle: handle}
	}()

	select {
	case <-ctx.Done():
		// Close the handle to unblock the ConnectNamedPipe call
		// The goroutine will complete shortly after the handle is closed
		syscall.CloseHandle(handle)
		// Wait for the goroutine to finish to avoid leaks
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(100 * time.Millisecond):
			// Timeout waiting for goroutine - it will complete eventually
			// This is acceptable as we've closed the handle
		}
		return ctx.Err()
	case res := <-done:
		return res.err
	}
}

func (s *winServer) prepareNextHandle() {
	s.mu.RLock()
	closed := s.closed
	hasNext := s.next != 0
	s.mu.RUnlock()
	if closed || hasNext {
		return
	}

	handle, err := createNamedPipe(s.pipeName, s.config.BufferSize)
	if err != nil {
		return
	}

	s.mu.Lock()
	if s.closed || s.next != 0 {
		s.mu.Unlock()
		_ = syscall.CloseHandle(handle)
		return
	}
	s.next = handle
	s.mu.Unlock()
}

func isNamedPipeAddr(addr string) bool {
	return strings.HasPrefix(addr, `\\.\pipe\`)
}

func normalizePipeAddr(addr string) string {
	if isNamedPipeAddr(addr) {
		return addr
	}
	return fmt.Sprintf(`\\.\pipe\%s`, addr)
}

func dialNamedPipe(ctx context.Context, pipeName string, timeout time.Duration) (syscall.Handle, error) {
	pipeNamePtr, err := syscall.UTF16PtrFromString(pipeName)
	if err != nil {
		return 0, err
	}

	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, fmt.Errorf("named pipe connect timeout")
		}
		procWaitNamedPipeW.Call(uintptr(unsafe.Pointer(pipeNamePtr)), uintptr(remaining/time.Millisecond))

		handle, err := syscall.CreateFile(
			pipeNamePtr,
			GENERIC_READ|GENERIC_WRITE,
			0,
			nil,
			OPEN_EXISTING,
			0,
			0,
		)
		if err == nil {
			return handle, nil
		}
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		if time.Now().After(deadline) {
			return 0, err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *winClient) Write(data []byte) (int, error) {
	return c.WriteWithTimeout(data, c.config.WriteTimeout)
}

func (c *winClient) WriteWithTimeout(data []byte, timeout time.Duration) (int, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrClientClosed
	}
	conn := c.conn
	handle := c.handle
	c.mu.RUnlock()

	if conn != nil {
		if timeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(timeout))
			defer conn.SetWriteDeadline(time.Time{})
		}
		return conn.Write(data)
	}

	if handle != 0 {
		var written uint32
		err := syscall.WriteFile(handle, data, &written, nil)
		return int(written), err
	}

	return 0, fmt.Errorf("no valid connection")
}

func (c *winClient) Read(buf []byte) (int, error) {
	return c.ReadWithTimeout(buf, c.config.ReadTimeout)
}

func (c *winClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrClientClosed
	}
	conn := c.conn
	handle := c.handle
	c.mu.RUnlock()

	if conn != nil {
		if timeout > 0 {
			conn.SetReadDeadline(time.Now().Add(timeout))
			defer conn.SetReadDeadline(time.Time{})
		}
		return conn.Read(buf)
	}

	if handle != 0 {
		var read uint32
		err := syscall.ReadFile(handle, buf, &read, nil)
		return int(read), err
	}

	return 0, fmt.Errorf("no valid connection")
}

func (c *winClient) RemoteAddr() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return "named-pipe"
}

func (c *winClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}

	if c.handle != 0 {
		procDisconnectNamedPipe.Call(uintptr(c.handle))
		if closeErr := syscall.CloseHandle(c.handle); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	return err
}
