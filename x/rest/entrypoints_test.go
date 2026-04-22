package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// --- RegisterResourceRoutes nil-arg and edge cases ---

func TestRegisterResourceRoutes_NilRouter_NoError(t *testing.T) {
	if err := RegisterResourceRoutes(nil, "/items", NewBaseResourceController("items"), DefaultRouteOptions()); err != nil {
		t.Errorf("nil router: unexpected error: %v", err)
	}
}

func TestRegisterResourceRoutes_NilController_NoError(t *testing.T) {
	r := router.NewRouter()
	if err := RegisterResourceRoutes(r, "/items", nil, DefaultRouteOptions()); err != nil {
		t.Errorf("nil controller: unexpected error: %v", err)
	}
}

func TestRegisterResourceRoutes_EmptyPrefix_NormalizesToSlash(t *testing.T) {
	r := router.NewRouter()
	if err := RegisterResourceRoutes(r, "", NewBaseResourceController("things"), RouteOptions{}); err != nil {
		t.Fatalf("empty prefix: %v", err)
	}
	routes := r.Routes()
	for _, route := range routes {
		if route.Path == "" {
			t.Errorf("route path should not be empty after normalization")
		}
	}
}

func TestDefaultRouteOptions_EnablesAll(t *testing.T) {
	opts := DefaultRouteOptions()
	if !opts.EnableBatch {
		t.Error("expected EnableBatch = true")
	}
	if !opts.EnableHead {
		t.Error("expected EnableHead = true")
	}
	if !opts.EnableOptions {
		t.Error("expected EnableOptions = true")
	}
}

// --- BaseResourceController negative-path coverage ---

func TestBaseResourceController_AllMethodsReturn501(t *testing.T) {
	c := NewBaseResourceController("orders")
	methods := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"Show", c.Show},
		{"Create", c.Create},
		{"Update", c.Update},
		{"Delete", c.Delete},
		{"Patch", c.Patch},
		{"BatchCreate", c.BatchCreate},
		{"BatchDelete", c.BatchDelete},
	}

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/orders", nil)
			m.fn(rec, req)

			if rec.Code != http.StatusNotImplemented {
				t.Errorf("%s: status = %d, want 501", m.name, rec.Code)
			}
			var resp contract.ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("%s: decode: %v", m.name, err)
			}
			if resp.Error.Code != contract.CodeNotImplemented {
				t.Errorf("%s: code = %q, want %q", m.name, resp.Error.Code, contract.CodeNotImplemented)
			}
		})
	}
}

func TestBaseResourceController_Options_SetsHeaders(t *testing.T) {
	c := NewBaseResourceController("items")
	rec := httptest.NewRecorder()
	c.Options(rec, httptest.NewRequest(http.MethodOptions, "/items", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("Options: status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Options: missing Access-Control-Allow-Methods header")
	}
}

func TestBaseResourceController_Head_Returns200(t *testing.T) {
	c := NewBaseResourceController("items")
	rec := httptest.NewRecorder()
	c.Head(rec, httptest.NewRequest(http.MethodHead, "/items", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("Head: status = %d, want 200", rec.Code)
	}
}

func TestBaseResourceController_NilReceiver_ResourceName(t *testing.T) {
	// resourceName() must not panic on nil or empty ResourceName.
	c := &BaseResourceController{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Index(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("empty ResourceName: status = %d, want 501", rec.Code)
	}
}

// --- NewPaginationMeta boundary cases ---

func TestNewPaginationMeta_FirstPage(t *testing.T) {
	meta := NewPaginationMeta(1, 10, 25)
	if meta.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", meta.TotalPages)
	}
	if meta.HasPrev {
		t.Error("HasPrev should be false on page 1")
	}
	if !meta.HasNext {
		t.Error("HasNext should be true on page 1 of 3")
	}
}

func TestNewPaginationMeta_LastPage(t *testing.T) {
	meta := NewPaginationMeta(3, 10, 25)
	if meta.HasNext {
		t.Error("HasNext should be false on last page")
	}
	if !meta.HasPrev {
		t.Error("HasPrev should be true on page 3")
	}
}

func TestNewPaginationMeta_ZeroItems(t *testing.T) {
	meta := NewPaginationMeta(1, 10, 0)
	if meta.TotalPages != 0 {
		t.Errorf("TotalPages = %d, want 0", meta.TotalPages)
	}
	if meta.HasNext || meta.HasPrev {
		t.Error("no navigation expected for zero items")
	}
}

// --- QueryBuilder negative-path cases ---

func TestQueryBuilder_PageSizeExceedsMax_Clamped(t *testing.T) {
	qb := NewQueryBuilder().WithPageSize(20, 50)
	req := httptest.NewRequest(http.MethodGet, "/?page_size=9999", nil)
	params := qb.Parse(req)
	if params.PageSize > 50 {
		t.Errorf("PageSize = %d, want ≤ 50", params.PageSize)
	}
}

func TestQueryBuilder_InvalidPage_UsesDefault(t *testing.T) {
	qb := NewQueryBuilder()
	req := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)
	params := qb.Parse(req)
	if params.Page != 1 {
		t.Errorf("Page = %d, want 1 for invalid input", params.Page)
	}
}

func TestQueryBuilder_UnknownSortField_Ignored(t *testing.T) {
	qb := NewQueryBuilder().WithAllowedSorts("name", "created_at")
	req := httptest.NewRequest(http.MethodGet, "/?sort=evil,name", nil)
	params := qb.Parse(req)
	for _, sf := range params.Sort {
		if sf.Field == "evil" {
			t.Error("disallowed sort field 'evil' was not filtered out")
		}
	}
}

func TestQueryBuilder_UnknownFilterField_Ignored(t *testing.T) {
	qb := NewQueryBuilder().WithAllowedFilters("status")
	req := httptest.NewRequest(http.MethodGet, "/?status=active&evil=injected", nil)
	params := qb.Parse(req)
	if _, ok := params.Filters["evil"]; ok {
		t.Error("disallowed filter 'evil' was not filtered out")
	}
	if params.Filters["status"] != "active" {
		t.Errorf("Filters[status] = %q, want active", params.Filters["status"])
	}
}

func TestQueryBuilder_NegativePageZero_UsesDefault(t *testing.T) {
	qb := NewQueryBuilder()
	req := httptest.NewRequest(http.MethodGet, "/?page=0", nil)
	params := qb.Parse(req)
	if params.Page != 1 {
		t.Errorf("Page = %d, want 1 for page=0", params.Page)
	}
}

func TestQueryBuilder_DescendingSort(t *testing.T) {
	qb := NewQueryBuilder().WithAllowedSorts("name", "created_at")
	req := httptest.NewRequest(http.MethodGet, "/?sort=-created_at,name", nil)
	params := qb.Parse(req)
	if len(params.Sort) != 2 {
		t.Fatalf("Sort len = %d, want 2", len(params.Sort))
	}
	if !params.Sort[0].Desc || params.Sort[0].Field != "created_at" {
		t.Errorf("first sort = {%v %v}, want {created_at desc}", params.Sort[0].Field, params.Sort[0].Desc)
	}
}
