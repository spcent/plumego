// Package server wraps caller-owned RPC runtimes with explicit lifecycle
// helpers.
package server

import (
	"context"
	"errors"
	"net"
)

var ErrRuntimeNil = errors.New("rpc server runtime is nil")

// Runtime is the dependency-free lifecycle surface required by Server.
type Runtime interface {
	RegisterService(desc any, impl any) error
	Serve(net.Listener) error
	GracefulStop()
	Stop()
}

type Server struct {
	runtime Runtime
}

func New(runtime Runtime) *Server {
	return &Server{runtime: runtime}
}

func (s *Server) RegisterService(desc any, impl any) error {
	if s == nil || s.runtime == nil {
		return ErrRuntimeNil
	}
	return s.runtime.RegisterService(desc, impl)
}

func (s *Server) Serve(lis net.Listener) error {
	if s == nil || s.runtime == nil {
		return ErrRuntimeNil
	}
	return s.runtime.Serve(lis)
}

func (s *Server) GracefulStop(ctx context.Context) error {
	if s == nil || s.runtime == nil {
		return ErrRuntimeNil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	done := make(chan struct{})
	go func() {
		s.runtime.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.runtime.Stop()
		<-done
		return ctx.Err()
	}
}
