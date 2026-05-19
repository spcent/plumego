package server

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestRegisterServeGracefulStop(t *testing.T) {
	runtime := &fakeRuntime{serveDone: make(chan struct{})}
	srv := New(runtime)

	if err := srv.RegisterService("desc", "impl"); err != nil {
		t.Fatalf("RegisterService() error = %v", err)
	}
	if runtime.desc != "desc" || runtime.impl != "impl" {
		t.Fatalf("registered service = %#v/%#v", runtime.desc, runtime.impl)
	}

	errs := make(chan error, 1)
	go func() {
		errs <- srv.Serve(nil)
	}()

	stopCtx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := srv.GracefulStop(stopCtx); err != nil {
		t.Fatalf("GracefulStop() error = %v", err)
	}
	if err := <-errs; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	if !runtime.graceful {
		t.Fatal("expected runtime.GracefulStop to be called")
	}
}

func TestGracefulStopCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := New(&fakeRuntime{}).GracefulStop(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("GracefulStop() error = %v, want %v", err, context.Canceled)
	}
}

func TestNilRuntimeReturnsError(t *testing.T) {
	if err := New(nil).Serve(nil); !errors.Is(err, ErrRuntimeNil) {
		t.Fatalf("Serve() error = %v, want %v", err, ErrRuntimeNil)
	}
}

type fakeRuntime struct {
	desc      any
	impl      any
	serveDone chan struct{}
	graceful  bool
	stopped   bool
}

func (r *fakeRuntime) RegisterService(desc any, impl any) error {
	r.desc = desc
	r.impl = impl
	return nil
}

func (r *fakeRuntime) Serve(net.Listener) error {
	if r.serveDone == nil {
		return nil
	}
	<-r.serveDone
	return nil
}

func (r *fakeRuntime) GracefulStop() {
	r.graceful = true
	if r.serveDone != nil {
		close(r.serveDone)
	}
}

func (r *fakeRuntime) Stop() {
	r.stopped = true
	if r.serveDone != nil {
		close(r.serveDone)
	}
}
