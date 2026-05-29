package search

// SearchQuery carries all search parameters from the HTTP layer to the engine.
type SearchQuery struct {
	Q            string
	TagID        string
	Status       string
	ReviewStatus string
	SourceType   string
	ImportJobID  string
	IsFavorite   *bool
	From         string // ISO datetime lower bound on updated_at
	To           string // ISO datetime upper bound on updated_at
	Sort         string // "relevance" | "updated_at" | "title"
	Order        string // "asc" | "desc"
	Limit        int
	Offset       int
}

// SearchResultItem is one document in a search response.
type SearchResultItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary,omitempty"`
	Highlights   []string `json:"highlights,omitempty"`
	Score        float64  `json:"score"`
	Tags         []string `json:"tags,omitempty"`
	OriginalPath string   `json:"original_path,omitempty"`
	SourceType   string   `json:"source_type"`
	IsFavorite   bool     `json:"is_favorite"`
	UpdatedAt    string   `json:"updated_at"`
}

// SearchResult is the full paginated search response.
type SearchResult struct {
	Items  []SearchResultItem `json:"items"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// ReindexRequest specifies which documents to re-index.
type ReindexRequest struct {
	Scope      string `json:"scope"`       // all | failed | stale | document
	DocumentID string `json:"document_id,omitempty"`
}

// IndexStatusResponse is the HTTP response for GET /search/index-status.
type IndexStatusResponse struct {
	TotalDocuments int64  `json:"total_documents"`
	Indexed        int64  `json:"indexed"`
	Pending        int64  `json:"pending"`
	Failed         int64  `json:"failed"`
	Stale          int64  `json:"stale"`
	LastIndexedAt  string `json:"last_indexed_at,omitempty"`
}

// HistoryItem is one entry in the search history response.
type HistoryItem struct {
	ID          string `json:"id"`
	Query       string `json:"query"`
	ResultCount int    `json:"result_count"`
	CreatedAt   string `json:"created_at"`
}

// HistoryResponse is the HTTP response for GET /search/history.
type HistoryResponse struct {
	Items []HistoryItem `json:"items"`
}
