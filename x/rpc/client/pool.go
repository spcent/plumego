// Package client provides gRPC client pooling and unary interceptors.
package client

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var ErrPoolClosed = errors.New("rpc client pool is closed")

type Pool struct {
	mu         sync.Mutex
	defaultOps []grpc.DialOption
	conns      map[string]*grpc.ClientConn
	closed     bool
}

func New(defaultOpts ...grpc.DialOption) *Pool {
	return &Pool{
		defaultOps: append([]grpc.DialOption(nil), defaultOpts...),
		conns:      make(map[string]*grpc.ClientConn),
	}
}

func (p *Pool) Dial(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrPoolClosed
	}
	if conn, ok := p.conns[target]; ok {
		p.mu.Unlock()
		return conn, nil
	}
	dialOpts := append([]grpc.DialOption(nil), p.defaultOps...)
	dialOpts = append(dialOpts, opts...)
	p.mu.Unlock()

	conn, err := grpc.DialContext(ctx, target, dialOpts...)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		_ = conn.Close()
		return nil, ErrPoolClosed
	}
	if existing, ok := p.conns[target]; ok {
		_ = conn.Close()
		return existing, nil
	}
	p.conns[target] = conn
	return conn, nil
}

func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	conns := make([]*grpc.ClientConn, 0, len(p.conns))
	for target, conn := range p.conns {
		conns = append(conns, conn)
		delete(p.conns, target)
	}
	p.mu.Unlock()

	var errs []error
	for _, conn := range conns {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func WithKeepAlive(params keepalive.ClientParameters) grpc.DialOption {
	return grpc.WithKeepaliveParams(params)
}
