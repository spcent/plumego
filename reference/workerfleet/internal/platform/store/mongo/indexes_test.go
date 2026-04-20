package mongo

import "testing"

func TestCollectionIndexSpecsIncludeTTLAndUniqueRules(t *testing.T) {
	specs := CollectionIndexSpecs()

	active := specs[CollectionWorkerActive]
	if len(active) == 0 || !active[0].Unique {
		t.Fatalf("worker_active first index should be unique, got %#v", active)
	}

	for _, collection := range []string{CollectionTaskHistory, CollectionWorkerEvents, CollectionAlertEvents} {
		foundTTL := false
		for _, spec := range specs[collection] {
			if spec.Name == "expire_at_ttl" && spec.ExpireAfterSeconds != nil && *spec.ExpireAfterSeconds == 0 {
				foundTTL = true
				break
			}
		}
		if !foundTTL {
			t.Fatalf("%s missing expire_at TTL index", collection)
		}
	}
}

func TestCollectionIndexSpecsAreStableAcrossCalls(t *testing.T) {
	first := CollectionIndexSpecs()
	second := CollectionIndexSpecs()

	if len(first) != len(second) {
		t.Fatalf("spec count mismatch: %d vs %d", len(first), len(second))
	}
	if len(first[CollectionWorkerSnapshots]) != len(second[CollectionWorkerSnapshots]) {
		t.Fatalf("worker snapshot index count mismatch")
	}
}
