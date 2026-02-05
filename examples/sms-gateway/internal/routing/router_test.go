package routing

import (
	"context"
	"testing"
	"time"

	"github.com/spcent/plumego/tenant"
)

func TestPolicyRouterWithCachedProvider(t *testing.T) {
	store := tenant.NewInMemoryRoutePolicyStore()
	cache := tenant.NewInMemoryRoutePolicyCache(10, time.Minute)
	provider := tenant.NewCachedRoutePolicyProvider(store, cache)

	payload := []byte(`{"providers":[{"provider":"a","weight":70},{"provider":"b","weight":30}]}`)
	if err := store.SetRoutePolicy(context.Background(), tenant.RoutePolicy{
		TenantID: "t-1",
		Strategy: "weighted",
		Payload:  payload,
	}); err != nil {
		t.Fatalf("set policy: %v", err)
	}

	router := &PolicyRouter{
		Provider: provider,
		Random:   fixedRand{value: 10},
	}

	first, _, err := router.SelectProvider(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("select provider: %v", err)
	}
	if first != "a" {
		t.Fatalf("expected provider a, got %s", first)
	}

	_ = store.DeleteRoutePolicy(context.Background(), "t-1")

	second, _, err := router.SelectProvider(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("select provider after delete: %v", err)
	}
	if second != "a" {
		t.Fatalf("expected cached provider a, got %s", second)
	}
}

type fixedRand struct {
	value int
}

func (r fixedRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return r.value % n
}
