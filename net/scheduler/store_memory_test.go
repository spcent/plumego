package scheduler

import "testing"

func TestWithDependsOnCopiesInputSlice(t *testing.T) {
	deps := []JobID{"dep-a", "dep-b"}
	opts := defaultJobOptions()

	WithDependsOn(DependencyFailureSkip, deps...)(&opts)
	deps[0] = "mutated"

	if len(opts.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(opts.Dependencies))
	}
	if opts.Dependencies[0] != "dep-a" {
		t.Fatalf("expected dependency copy to remain dep-a, got %s", opts.Dependencies[0])
	}
}

func TestMemoryStoreCopiesStoredJobSlices(t *testing.T) {
	store := NewMemoryStore()
	original := StoredJob{
		ID:           "job-1",
		Tags:         []string{"tag-a"},
		Dependencies: []JobID{"dep-a"},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("save: %v", err)
	}

	original.Tags[0] = "mutated-tag"
	original.Dependencies[0] = "mutated-dep"

	items, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 stored job, got %d", len(items))
	}
	if items[0].Tags[0] != "tag-a" {
		t.Fatalf("expected stored tag tag-a, got %s", items[0].Tags[0])
	}
	if items[0].Dependencies[0] != "dep-a" {
		t.Fatalf("expected stored dep dep-a, got %s", items[0].Dependencies[0])
	}

	items[0].Tags[0] = "external-tag"
	items[0].Dependencies[0] = "external-dep"

	items2, err := store.List()
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	if items2[0].Tags[0] != "tag-a" {
		t.Fatalf("expected internal tag to stay tag-a, got %s", items2[0].Tags[0])
	}
	if items2[0].Dependencies[0] != "dep-a" {
		t.Fatalf("expected internal dep to stay dep-a, got %s", items2[0].Dependencies[0])
	}
}
