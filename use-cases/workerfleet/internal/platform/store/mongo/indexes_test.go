package mongo

import "testing"

func TestCollectionIndexSpecsIncludeTTLAndUniqueRules(t *testing.T) {
	specs := CollectionIndexSpecs()

	active := specs[CollectionWorkerActive]
	if len(active) == 0 || !active[0].Unique {
		t.Fatalf("worker_active first index should be unique, got %#v", active)
	}

	for _, collection := range []string{CollectionTaskHistory, CollectionCaseStepHistory, CollectionWorkerEvents, CollectionAlertEvents} {
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

	loopLeases := specs[CollectionLoopLeases]
	if len(loopLeases) == 0 {
		t.Fatalf("loop lease indexes missing")
	}
	foundLeaseTTL := false
	for _, spec := range loopLeases {
		if spec.Name == "expires_at_ttl" && spec.ExpireAfterSeconds != nil && *spec.ExpireAfterSeconds == 0 {
			foundLeaseTTL = true
			break
		}
	}
	if !foundLeaseTTL {
		t.Fatalf("loop leases missing expires_at TTL index")
	}

	notificationJobs := specs[CollectionNotificationJobs]
	if len(notificationJobs) == 0 || !notificationJobs[0].Unique {
		t.Fatalf("notification jobs first index should be unique, got %#v", notificationJobs)
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
	if len(first[CollectionCaseStepHistory]) != len(second[CollectionCaseStepHistory]) {
		t.Fatalf("case step history index count mismatch")
	}
	if len(first[CollectionLoopLeases]) != len(second[CollectionLoopLeases]) {
		t.Fatalf("loop lease index count mismatch")
	}
	if len(first[CollectionNotificationJobs]) != len(second[CollectionNotificationJobs]) {
		t.Fatalf("notification job index count mismatch")
	}
}
