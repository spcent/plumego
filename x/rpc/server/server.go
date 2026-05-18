// Package server wraps grpc.Server with explicit lifecycle helpers.
package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Option = grpc.ServerOption

type Server struct {
	grpc *grpc.Server
}

func New(opts ...grpc.ServerOption) *Server {
	return &Server{grpc: grpc.NewServer(opts...)}
}

func WithInterceptors(unary grpc.UnaryServerInterceptor, stream grpc.StreamServerInterceptor) []Option {
	opts := make([]Option, 0, 2)
	if unary != nil {
		opts = append(opts, grpc.ChainUnaryInterceptor(unary))
	}
	if stream != nil {
		opts = append(opts, grpc.ChainStreamInterceptor(stream))
	}
	return opts
}

func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.grpc.RegisterService(desc, impl)
}

func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

func (s *Server) GracefulStop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	done := make(chan struct{})
	go func() {
		s.grpc.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.grpc.Stop()
		<-done
		return ctx.Err()
	}
}
