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
		return nil, net.ErrClosed
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

func (s *winServer) Accept() (Client, error) {
	return s.AcceptWithContext(context.Background())
}

func (s *winServer) AcceptWithContext(ctx context.Context) (Client, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, net.ErrClosed
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

	s.prepareNextHandle()

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
		Timeout: config.ConnectTimeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
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

func connectNamedPipeWithContext(ctx context.Context, handle syscall.Handle) error {
	done := make(chan error, 1)
	go func() {
		ret, _, err := procConnectNamedPipe.Call(uintptr(handle), 0)
		if ret == 0 {
			if err == errorPipeConnected {
				done <- nil
				return
			}
			if err != nil {
				done <- err
				return
			}
			done <- fmt.Errorf("failed to connect named pipe")
			return
		}
		done <- nil
	}()

	select {
	case <-ctx.Done():
		go syscall.CloseHandle(handle)
		return ctx.Err()
	case err := <-done:
		return err
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
	defer c.mu.RUnlock()

	if c.closed {
		return 0, net.ErrClosed
	}

	if c.conn != nil {
		if timeout > 0 {
			c.conn.SetWriteDeadline(time.Now().Add(timeout))
			defer c.conn.SetWriteDeadline(time.Time{})
		}
		return c.conn.Write(data)
	}

	if c.handle != 0 {
		var written uint32
		err := syscall.WriteFile(c.handle, data, &written, nil)
		return int(written), err
	}

	return 0, fmt.Errorf("no valid connection")
}

func (c *winClient) Read(buf []byte) (int, error) {
	return c.ReadWithTimeout(buf, c.config.ReadTimeout)
}

func (c *winClient) ReadWithTimeout(buf []byte, timeout time.Duration) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, net.ErrClosed
	}

	if c.conn != nil {
		if timeout > 0 {
			c.conn.SetReadDeadline(time.Now().Add(timeout))
			defer c.conn.SetReadDeadline(time.Time{})
		}
		return c.conn.Read(buf)
	}

	if c.handle != 0 {
		var read uint32
		err := syscall.ReadFile(c.handle, buf, &read, nil)
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
