package document

import (
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestBuildListWhere_defaults(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{})
	if len(wheres) != 1 {
		t.Fatalf("want 1 clause, got %d: %v", len(wheres), wheres)
	}
	if wheres[0] != "status = ?" {
		t.Errorf("default clause: %q", wheres[0])
	}
	if args[0] != StatusActive {
		t.Errorf("default status arg: %q", args[0])
	}
}

func TestBuildListWhere_allStatus(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{Status: "all"})
	if wheres[0] != "status != 'deleted'" {
		t.Errorf("all status clause: %q", wheres[0])
	}
	if len(args) != 0 {
		t.Errorf("all status should have no args")
	}
}

func TestBuildListWhere_archivedStatus(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{Status: StatusArchived})
	if wheres[0] != "status = ?" || args[0] != StatusArchived {
		t.Errorf("archived: wheres=%v args=%v", wheres, args)
	}
}

func TestBuildListWhere_searchQuery(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{Q: "hello"})
	found := false
	for i, w := range wheres {
		if w == "title LIKE ?" {
			found = true
			if args[i] != "%hello%" {
				t.Errorf("search arg: want '%%hello%%', got %q", args[i])
			}
		}
	}
	if !found {
		t.Errorf("search clause missing: %v", wheres)
	}
}

func TestBuildListWhere_sourceType(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{SourceType: SourceTypeImported})
	assertContainsClause(t, wheres, args, "source_type = ?", SourceTypeImported)
}

func TestBuildListWhere_importJobID(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{ImportJobID: "job-1"})
	assertContainsClause(t, wheres, args, "import_job_id = ?", "job-1")
}

func TestBuildListWhere_isFavoriteTrue(t *testing.T) {
	wheres, _ := buildListWhere(ListQuery{IsFavorite: boolPtr(true)})
	assertClausePresent(t, wheres, "is_favorite = 1")
}

func TestBuildListWhere_isFavoriteFalse(t *testing.T) {
	wheres, _ := buildListWhere(ListQuery{IsFavorite: boolPtr(false)})
	assertClausePresent(t, wheres, "is_favorite = 0")
}

func TestBuildListWhere_reviewStatus(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{ReviewStatus: ReviewStatusReviewed})
	assertContainsClause(t, wheres, args, "review_status = ?", ReviewStatusReviewed)
}

func TestBuildListWhere_tagID(t *testing.T) {
	wheres, args := buildListWhere(ListQuery{TagID: "tag-1"})
	found := false
	for i, w := range wheres {
		if w == "id IN (SELECT document_id FROM document_tags WHERE tag_id = ?)" {
			found = true
			if args[i] != "tag-1" {
				t.Errorf("tag arg: want 'tag-1', got %v", args[i])
			}
		}
	}
	if !found {
		t.Errorf("tag clause missing: %v", wheres)
	}
}

func TestBuildOrderClause_defaults(t *testing.T) {
	got := buildOrderClause(ListQuery{})
	if got != "ORDER BY updated_at DESC, id DESC" {
		t.Errorf("default order: %q", got)
	}
}

func TestBuildOrderClause_titleAsc(t *testing.T) {
	got := buildOrderClause(ListQuery{SortBy: "title", Order: "ASC"})
	if got != "ORDER BY title ASC, id ASC" {
		t.Errorf("title asc: %q", got)
	}
}

func TestBuildOrderClause_invalidColFallback(t *testing.T) {
	got := buildOrderClause(ListQuery{SortBy: "injection; DROP TABLE", Order: "asc"})
	if got != "ORDER BY updated_at ASC, id ASC" {
		t.Errorf("invalid col should fall back to updated_at: %q", got)
	}
}

// --- helpers ---

func assertContainsClause(t *testing.T, wheres []string, args []any, clause string, val any) {
	t.Helper()
	for i, w := range wheres {
		if w == clause {
			if i < len(args) && args[i] == val {
				return
			}
			t.Errorf("clause %q found but arg mismatch: want %v, got %v", clause, val, args[i])
			return
		}
	}
	t.Errorf("clause %q not found in %v", clause, wheres)
}

func assertClausePresent(t *testing.T, wheres []string, clause string) {
	t.Helper()
	for _, w := range wheres {
		if w == clause {
			return
		}
	}
	t.Errorf("clause %q not found in %v", clause, wheres)
}
