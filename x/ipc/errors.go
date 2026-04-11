package ipc

import "errors"

var (
	// ErrServerClosed is returned when operations are attempted on a closed server.
	ErrServerClosed = errors.New("ipc: server closed")
	// ErrClientClosed is returned when operations are attempted on a closed client.
	ErrClientClosed = errors.New("ipc: client closed")
	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("ipc: invalid configuration")
	// ErrConnectTimeout is returned when connection times out.
	ErrConnectTimeout = errors.New("ipc: connection timeout")
	// ErrPlatformNotSupported is returned when the platform is not supported.
	ErrPlatformNotSupported = errors.New("ipc: platform not supported")
	// ErrInvalidAddress is returned when the address format is invalid.
	ErrInvalidAddress = errors.New("ipc: invalid address")
)
