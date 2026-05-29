package search

import (
	"strings"
	"testing"
)

// --- appendDocFilters ---

func TestAppendDocFilters_defaultStatus(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{})
	if len(wheres) != 1 || wheres[0] != "d.status = 'active'" {
		t.Errorf("default status: want [d.status = 'active'], got %v", wheres)
	}
	if len(args) != 0 {
		t.Errorf("default status: want no args, got %v", args)
	}
}

func TestAppendDocFilters_statusAll(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{Status: "all"})
	if len(wheres) != 1 || wheres[0] != "d.status != 'deleted'" {
		t.Errorf("status=all: want [d.status != 'deleted'], got %v", wheres)
	}
	if len(args) != 0 {
		t.Errorf("status=all: want no args, got %v", args)
	}
}

func TestAppendDocFilters_statusExplicit(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{Status: "archived"})
	if len(wheres) != 1 || wheres[0] != "d.status = ?" {
		t.Errorf("status=archived: want [d.status = ?], got %v", wheres)
	}
	if len(args) != 1 || args[0] != "archived" {
		t.Errorf("status=archived: want ['archived'], got %v", args)
	}
}

func TestAppendDocFilters_sourceType(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{SourceType: "manual"})
	found := false
	for _, w := range wheres {
		if w == "d.source_type = ?" {
			found = true
		}
	}
	if !found {
		t.Errorf("source_type filter missing: %v", wheres)
	}
	foundArg := false
	for _, a := range args {
		if a == "manual" {
			foundArg = true
		}
	}
	if !foundArg {
		t.Errorf("source_type arg missing: %v", args)
	}
}

func TestAppendDocFilters_reviewStatus(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{ReviewStatus: "pending"})
	var found bool
	for _, w := range wheres {
		if w == "d.review_status = ?" {
			found = true
		}
	}
	if !found {
		t.Errorf("review_status filter missing: %v", wheres)
	}
	if len(args) == 0 || args[len(args)-1] != "pending" {
		t.Errorf("review_status arg missing: %v", args)
	}
}

func TestAppendDocFilters_importJobID(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{ImportJobID: "job-1"})
	var found bool
	for _, w := range wheres {
		if w == "d.import_job_id = ?" {
			found = true
		}
	}
	if !found {
		t.Errorf("import_job_id filter missing: %v", wheres)
	}
	if len(args) == 0 || args[len(args)-1] != "job-1" {
		t.Errorf("import_job_id arg missing: %v", args)
	}
}

func TestAppendDocFilters_isFavoriteTrue(t *testing.T) {
	yes := true
	wheres, args := appendDocFilters(nil, nil, SearchQuery{IsFavorite: &yes})
	var found bool
	for _, w := range wheres {
		if w == "d.is_favorite = 1" {
			found = true
		}
	}
	if !found {
		t.Errorf("is_favorite=true filter missing: %v", wheres)
	}
	if len(args) != 0 {
		t.Errorf("is_favorite=true should add no args, got %v", args)
	}
}

func TestAppendDocFilters_isFavoriteFalse(t *testing.T) {
	no := false
	wheres, _ := appendDocFilters(nil, nil, SearchQuery{IsFavorite: &no})
	var found bool
	for _, w := range wheres {
		if w == "d.is_favorite = 0" {
			found = true
		}
	}
	if !found {
		t.Errorf("is_favorite=false filter missing: %v", wheres)
	}
}

func TestAppendDocFilters_tagID(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{TagID: "tag-42"})
	var found bool
	for _, w := range wheres {
		if strings.Contains(w, "document_tags") {
			found = true
		}
	}
	if !found {
		t.Errorf("tag subquery filter missing: %v", wheres)
	}
	if len(args) == 0 || args[len(args)-1] != "tag-42" {
		t.Errorf("tag arg missing: %v", args)
	}
}

func TestAppendDocFilters_dateRange(t *testing.T) {
	wheres, args := appendDocFilters(nil, nil, SearchQuery{From: "2026-01-01T00:00:00Z", To: "2026-12-31T23:59:59Z"})
	var hasFrom, hasTo bool
	for _, w := range wheres {
		if w == "d.updated_at >= ?" {
			hasFrom = true
		}
		if w == "d.updated_at <= ?" {
			hasTo = true
		}
	}
	if !hasFrom {
		t.Errorf("from filter missing: %v", wheres)
	}
	if !hasTo {
		t.Errorf("to filter missing: %v", wheres)
	}
	_ = args
}

// --- ftsOrderClause ---

func TestFtsOrderClause_default(t *testing.T) {
	got := ftsOrderClause(SearchQuery{})
	if got != "ORDER BY bm25(document_fts)" {
		t.Errorf("default: want bm25, got %q", got)
	}
}

func TestFtsOrderClause_updatedAtDesc(t *testing.T) {
	got := ftsOrderClause(SearchQuery{Sort: "updated_at"})
	if got != "ORDER BY d.updated_at DESC" {
		t.Errorf("updated_at desc: got %q", got)
	}
}

func TestFtsOrderClause_updatedAtAsc(t *testing.T) {
	got := ftsOrderClause(SearchQuery{Sort: "updated_at", Order: "ASC"})
	if got != "ORDER BY d.updated_at ASC" {
		t.Errorf("updated_at asc: got %q", got)
	}
}

func TestFtsOrderClause_titleDesc(t *testing.T) {
	got := ftsOrderClause(SearchQuery{Sort: "title"})
	if got != "ORDER BY d.title DESC" {
		t.Errorf("title desc: got %q", got)
	}
}

func TestFtsOrderClause_titleAsc(t *testing.T) {
	got := ftsOrderClause(SearchQuery{Sort: "title", Order: "asc"})
	if got != "ORDER BY d.title ASC" {
		t.Errorf("title asc: got %q", got)
	}
}

// --- fallbackOrderClause ---

func TestFallbackOrderClause_default(t *testing.T) {
	got := fallbackOrderClause(SearchQuery{})
	if got != "ORDER BY d.updated_at DESC" {
		t.Errorf("default: got %q", got)
	}
}

func TestFallbackOrderClause_titleAsc(t *testing.T) {
	got := fallbackOrderClause(SearchQuery{Sort: "title", Order: "ASC"})
	if got != "ORDER BY d.title ASC" {
		t.Errorf("title asc: got %q", got)
	}
}

func TestFallbackOrderClause_createdAtDesc(t *testing.T) {
	got := fallbackOrderClause(SearchQuery{Sort: "created_at"})
	if got != "ORDER BY d.created_at DESC" {
		t.Errorf("created_at desc: got %q", got)
	}
}

func TestFallbackOrderClause_createdAtAsc(t *testing.T) {
	got := fallbackOrderClause(SearchQuery{Sort: "created_at", Order: "ASC"})
	if got != "ORDER BY d.created_at ASC" {
		t.Errorf("created_at asc: got %q", got)
	}
}
