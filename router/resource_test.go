package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

type mockController struct {
	indexCalled  int
	showCalled   int
	createCalled int
	updateCalled int
	deleteCalled int
	patchCalled  int
	lastParams   map[string]string
}

func (m *mockController) Index(w http.ResponseWriter, r *http.Request) {
	m.indexCalled++
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusOK)
}

func (m *mockController) Show(w http.ResponseWriter, r *http.Request) {
	m.showCalled++
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusOK)
}

func (m *mockController) Create(w http.ResponseWriter, r *http.Request) {
	m.createCalled++
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusCreated)
}

func (m *mockController) Update(w http.ResponseWriter, r *http.Request) {
	m.updateCalled++
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusOK)
}

func (m *mockController) Delete(w http.ResponseWriter, r *http.Request) {
	m.deleteCalled++
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockController) Patch(w http.ResponseWriter, r *http.Request) {
	m.patchCalled++
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusOK)
}

func (m *mockController) Options(w http.ResponseWriter, r *http.Request) {
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockController) Head(w http.ResponseWriter, r *http.Request) {
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusOK)
}

func (m *mockController) BatchCreate(w http.ResponseWriter, r *http.Request) {
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusCreated)
}

func (m *mockController) BatchDelete(w http.ResponseWriter, r *http.Request) {
	m.lastParams = contract.RequestContextFrom(r.Context()).Params
	w.WriteHeader(http.StatusNoContent)
}

func TestResource_RouteRegistration(t *testing.T) {
	r := NewRouter()
	if err := r.Resource("/users", &mockController{}); err != nil {
		t.Fatalf("Resource registration failed: %v", err)
	}

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/users/123"},
		{"PUT", "/users/123"},
		{"DELETE", "/users/123"},
		{"PATCH", "/users/123"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code == http.StatusNotFound {
				t.Errorf("Route not registered: %s %s", tc.method, tc.path)
			}
		})
	}
}

func TestResource_ControllerInvocation(t *testing.T) {
	r := NewRouter()
	ctrl := &mockController{}
	if err := r.Resource("/posts", ctrl); err != nil {
		t.Fatalf("Resource registration failed: %v", err)
	}

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
		wantCall *int
		wantID   string
	}{
		{"Index", "GET", "/posts", http.StatusOK, &ctrl.indexCalled, ""},
		{"Create", "POST", "/posts", http.StatusCreated, &ctrl.createCalled, ""},
		{"Show", "GET", "/posts/456", http.StatusOK, &ctrl.showCalled, "456"},
		{"Update", "PUT", "/posts/789", http.StatusOK, &ctrl.updateCalled, "789"},
		{"Delete", "DELETE", "/posts/000", http.StatusNoContent, &ctrl.deleteCalled, "000"},
		{"Patch", "PATCH", "/posts/111", http.StatusOK, &ctrl.patchCalled, "111"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("Expected status code %d, got %d", tt.wantCode, rec.Code)
			}

			if *tt.wantCall != 1 {
				t.Errorf("Handler was not called, call count %d", *tt.wantCall)
			}

			if tt.wantID != "" && ctrl.lastParams["id"] != tt.wantID {
				t.Errorf("Expected param id=%s, got %s", tt.wantID, ctrl.lastParams["id"])
			}
		})
	}
}

func TestBaseResourceController_DefaultImplementation(t *testing.T) {
	ctrl := &BaseResourceController{}
	tests := []struct {
		name   string
		method string
		path   string
		call   func(w http.ResponseWriter, r *http.Request)
	}{
		{"Index", "GET", "/posts", func(w http.ResponseWriter, r *http.Request) { ctrl.Index(w, r) }},
		{"Show", "GET", "/posts/456", func(w http.ResponseWriter, r *http.Request) { ctrl.Show(w, r) }},
		{"Create", "POST", "/posts", func(w http.ResponseWriter, r *http.Request) { ctrl.Create(w, r) }},
		{"Update", "PUT", "/posts/789", func(w http.ResponseWriter, r *http.Request) { ctrl.Update(w, r) }},
		{"Delete", "DELETE", "/posts/000", func(w http.ResponseWriter, r *http.Request) { ctrl.Delete(w, r) }},
		{"Patch", "PATCH", "/posts/111", func(w http.ResponseWriter, r *http.Request) { ctrl.Patch(w, r) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			tt.call(rec, req)
			if rec.Code != http.StatusNotImplemented {
				t.Errorf("%s method expected 501, got %d", tt.name, rec.Code)
			}
		})
	}
}

func TestResource_PathTrimSuffix(t *testing.T) {
	r := NewRouter()
	ctrl := &mockController{}
	if err := r.Resource("/products/", ctrl); err != nil {
		t.Fatalf("Resource registration failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/products", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if ctrl.indexCalled != 1 {
		t.Error("Path with trailing slash was not trimmed correctly, Index handler was not called")
	}
}

// ================================================
// Enhanced CRUD Framework Tests
// ================================================

func TestQueryBuilder_Parse_Pagination(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected QueryParams
	}{
		{
			name:  "default pagination",
			query: "",
			expected: QueryParams{
				Page:     1,
				PageSize: 20,
				Limit:    20,
				Offset:   0,
				Filters:  map[string]string{},
			},
		},
		{
			name:  "custom page and page_size",
			query: "page=2&page_size=50",
			expected: QueryParams{
				Page:     2,
				PageSize: 50,
				Limit:    50,
				Offset:   50,
				Filters:  map[string]string{},
			},
		},
		{
			name:  "page_size exceeds max",
			query: "page_size=200",
			expected: QueryParams{
				Page:     1,
				PageSize: 100,
				Limit:    100,
				Offset:   0,
				Filters:  map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			qb := NewQueryBuilder()
			params := qb.Parse(req)

			if params.Page != tt.expected.Page {
				t.Errorf("Page = %d, want %d", params.Page, tt.expected.Page)
			}
			if params.PageSize != tt.expected.PageSize {
				t.Errorf("PageSize = %d, want %d", params.PageSize, tt.expected.PageSize)
			}
			if params.Limit != tt.expected.Limit {
				t.Errorf("Limit = %d, want %d", params.Limit, tt.expected.Limit)
			}
			if params.Offset != tt.expected.Offset {
				t.Errorf("Offset = %d, want %d", params.Offset, tt.expected.Offset)
			}
		})
	}
}

func TestQueryBuilder_Parse_Sorting(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []SortField
	}{
		{
			name:     "single ascending sort",
			query:    "sort=name",
			expected: []SortField{{Field: "name", Desc: false}},
		},
		{
			name:     "single descending sort",
			query:    "sort=-created_at",
			expected: []SortField{{Field: "created_at", Desc: true}},
		},
		{
			name:  "multiple sort fields",
			query: "sort=name,-created_at,email",
			expected: []SortField{
				{Field: "name", Desc: false},
				{Field: "created_at", Desc: true},
				{Field: "email", Desc: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			qb := NewQueryBuilder()
			params := qb.Parse(req)

			if len(params.Sort) != len(tt.expected) {
				t.Fatalf("Sort length = %d, want %d", len(params.Sort), len(tt.expected))
			}

			for i, sort := range params.Sort {
				if sort.Field != tt.expected[i].Field {
					t.Errorf("Sort[%d].Field = %s, want %s", i, sort.Field, tt.expected[i].Field)
				}
				if sort.Desc != tt.expected[i].Desc {
					t.Errorf("Sort[%d].Desc = %v, want %v", i, sort.Desc, tt.expected[i].Desc)
				}
			}
		})
	}
}

func TestNewPaginationMeta(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		totalItems int64
		expected   PaginationMeta
	}{
		{
			name:       "first page",
			page:       1,
			pageSize:   20,
			totalItems: 100,
			expected: PaginationMeta{
				Page:       1,
				PageSize:   20,
				TotalItems: 100,
				TotalPages: 5,
				HasNext:    true,
				HasPrev:    false,
			},
		},
		{
			name:       "last page",
			page:       5,
			pageSize:   20,
			totalItems: 100,
			expected: PaginationMeta{
				Page:       5,
				PageSize:   20,
				TotalItems: 100,
				TotalPages: 5,
				HasNext:    false,
				HasPrev:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewPaginationMeta(tt.page, tt.pageSize, tt.totalItems)

			if meta.Page != tt.expected.Page {
				t.Errorf("Page = %d, want %d", meta.Page, tt.expected.Page)
			}
			if meta.TotalPages != tt.expected.TotalPages {
				t.Errorf("TotalPages = %d, want %d", meta.TotalPages, tt.expected.TotalPages)
			}
			if meta.HasNext != tt.expected.HasNext {
				t.Errorf("HasNext = %v, want %v", meta.HasNext, tt.expected.HasNext)
			}
			if meta.HasPrev != tt.expected.HasPrev {
				t.Errorf("HasPrev = %v, want %v", meta.HasPrev, tt.expected.HasPrev)
			}
		})
	}
}

func TestParamExtractor_GetQueryInt(t *testing.T) {
	pe := NewParamExtractor()

	tests := []struct {
		name         string
		query        string
		paramName    string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			query:        "limit=50",
			paramName:    "limit",
			defaultValue: 10,
			expected:     50,
		},
		{
			name:         "missing parameter",
			query:        "",
			paramName:    "limit",
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			result := pe.GetQueryInt(req, tt.paramName, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("GetQueryInt() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestBatchProcessor_Process(t *testing.T) {
	processor := NewBatchProcessor(10)

	t.Run("successful batch", func(t *testing.T) {
		items := []any{1, 2, 3}
		result := processor.Process(nil, items, func(ctx context.Context, item any) error {
			return nil
		})

		if result.Successful != 3 {
			t.Errorf("Successful = %d, want 3", result.Successful)
		}
		if result.Failed != 0 {
			t.Errorf("Failed = %d, want 0", result.Failed)
		}
	})

	t.Run("batch too large", func(t *testing.T) {
		items := make([]any, 20)
		result := processor.Process(nil, items, func(ctx context.Context, item any) error {
			return nil
		})

		if result.Failed != 20 {
			t.Errorf("Failed = %d, want 20", result.Failed)
		}
		if len(result.Errors) != 1 {
			t.Errorf("Errors length = %d, want 1", len(result.Errors))
		}
	})
}
