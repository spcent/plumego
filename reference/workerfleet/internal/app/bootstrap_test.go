package app

import (
	"context"
	"strings"
	"testing"
)

func TestBootstrapMemoryBackend(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap memory: %v", err)
	}
	if runtime.Service == nil {
		t.Fatalf("service is nil")
	}
	if runtime.Handler == nil {
		t.Fatalf("handler is nil")
	}
	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close memory runtime: %v", err)
	}
}

func TestBootstrapMongoFailsClosedWhenMissingConfig(t *testing.T) {
	_, err := Bootstrap(context.Background(), Config{StoreBackend: StoreBackendMongo})
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_MONGO_URI") {
		t.Fatalf("error = %v, want missing mongo uri", err)
	}
}
