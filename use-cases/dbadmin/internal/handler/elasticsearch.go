package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/domain/eshistory"
	"dbadmin/internal/esmanager"
)

// ElasticsearchHandler handles Elasticsearch-specific endpoints.
type ElasticsearchHandler struct {
	Connections *connection.Store
	ESManager   *esmanager.Manager
	History     *eshistory.Store
	Logger      plumelog.StructuredLogger
}

// --- helpers -----------------------------------------------------------------

func (h ElasticsearchHandler) openClient(w http.ResponseWriter, r *http.Request, connID string) (*elasticsearch.Client, *connection.Connection) {
	conn, err := h.Connections.Get(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return nil, nil
		}
		h.internalErr(w, r, "get connection", err)
		return nil, nil
	}

	client, err := h.ESManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, "open ES client", err)
		return nil, nil
	}

	return client, conn
}

func (h ElasticsearchHandler) connNotFound(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotFound).Message("connection not found").Build()))
}

func (h ElasticsearchHandler) internalErr(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.Logger.Error("elasticsearch handler error", plumelog.Fields{"msg": msg, "error": err.Error()})
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).Message("internal error").Build()))
}

func (h ElasticsearchHandler) badRequest(w http.ResponseWriter, r *http.Request, msg string) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeBadRequest).Message(msg).Build()))
}

func (h ElasticsearchHandler) readonly(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeForbidden).Message("connection is read-only").Build()))
}

// validateIndexName checks if an index name is valid.
func validateIndexName(name string, allowWildcard bool) error {
	if name == "" {
		return fmt.Errorf("index name is required")
	}
	if len(name) > 255 {
		return fmt.Errorf("index name too long")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("index name cannot contain '..'")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("index name cannot start with '-'")
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("index name cannot contain spaces")
	}
	if allowWildcard && name == "*" {
		return nil
	}
	if strings.Contains(name, "*") && name != "*" {
		return fmt.Errorf("wildcard only allowed as '*' alone")
	}
	return nil
}

// --- endpoints ---------------------------------------------------------------

// ClusterInfo returns cluster information.
func (h ElasticsearchHandler) ClusterInfo(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	res, err := client.Info()
	if err != nil {
		h.internalErr(w, r, "cluster info", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		h.internalErr(w, r, "cluster info", fmt.Errorf("status: %s", res.Status()))
		return
	}

	var info map[string]any
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		h.internalErr(w, r, "decode cluster info", err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, info, nil))
}

// ListIndices returns a list of all indices.
func (h ElasticsearchHandler) ListIndices(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	req := esapi.CatIndicesRequest{
		Format: "json",
		H:      []string{"index", "health", "status", "docs.count", "store.size", "pri", "rep"},
	}
	res, err := req.Do(r.Context(), client)
	if err != nil {
		h.internalErr(w, r, "list indices", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		h.internalErr(w, r, "list indices", fmt.Errorf("status: %s", res.Status()))
		return
	}

	var indices []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		h.internalErr(w, r, "decode indices", err)
		return
	}

	// Filter out system indices (starting with '.')
	var filtered []map[string]any
	for _, idx := range indices {
		if name, ok := idx["index"].(string); ok && !strings.HasPrefix(name, ".") {
			filtered = append(filtered, idx)
		}
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"indices": filtered,
	}, map[string]any{"count": len(filtered)}))
}

// GetMapping returns the mapping for an index.
func (h ElasticsearchHandler) GetMapping(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	indexName := r.URL.Query().Get("index")

	if err := validateIndexName(indexName, false); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}

	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	req := esapi.IndicesGetMappingRequest{
		Index: []string{indexName},
	}
	res, err := req.Do(r.Context(), client)
	if err != nil {
		h.internalErr(w, r, "get mapping", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		h.internalErr(w, r, "get mapping", fmt.Errorf("status: %s", res.Status()))
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		h.internalErr(w, r, "read mapping", err)
		return
	}

	// Pretty-print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		h.internalErr(w, r, "format mapping", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(prettyJSON.Bytes())
}

// GetSettings returns the settings for an index.
func (h ElasticsearchHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	indexName := r.URL.Query().Get("index")

	if err := validateIndexName(indexName, false); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}

	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	req := esapi.IndicesGetSettingsRequest{
		Index: []string{indexName},
	}
	res, err := req.Do(r.Context(), client)
	if err != nil {
		h.internalErr(w, r, "get settings", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		h.internalErr(w, r, "get settings", fmt.Errorf("status: %s", res.Status()))
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		h.internalErr(w, r, "read settings", err)
		return
	}

	// Pretty-print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		h.internalErr(w, r, "format settings", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(prettyJSON.Bytes())
}

// Search executes a search DSL query.
func (h ElasticsearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Index string          `json:"index"`
		DSL   json.RawMessage `json:"dsl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if err := validateIndexName(req.Index, true); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}

	if len(req.DSL) == 0 {
		h.badRequest(w, r, "dsl is required")
		return
	}

	if !json.Valid(req.DSL) {
		h.badRequest(w, r, "dsl must be valid JSON")
		return
	}

	// Parse DSL to enforce size limits
	var dslMap map[string]any
	if err := json.Unmarshal(req.DSL, &dslMap); err != nil {
		h.badRequest(w, r, "invalid DSL format")
		return
	}

	// Enforce size: default 50, max 500
	sizeVal, hasSize := dslMap["size"]
	if !hasSize {
		dslMap["size"] = 50
	} else if sizeFloat, ok := sizeVal.(float64); ok {
		if sizeFloat > 500 {
			dslMap["size"] = 500
		}
	}

	// Re-marshal with enforced size
	enforcedDSL, err := json.Marshal(dslMap)
	if err != nil {
		h.internalErr(w, r, "marshal DSL", err)
		return
	}

	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	start := time.Now()
	res, err := client.Search(
		client.Search.WithContext(r.Context()),
		client.Search.WithIndex(req.Index),
		client.Search.WithBody(bytes.NewReader(enforcedDSL)),
	)
	duration := time.Since(start).Milliseconds()

	// Record history
	entry := &eshistory.Entry{
		ID:        uuid.New().String(),
		ConnID:    connID,
		Index:     req.Index,
		DSL:       string(enforcedDSL),
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	if err != nil {
		entry.Error = err.Error()
		h.History.Add(entry)
		h.internalErr(w, r, "execute search", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		entry.Error = string(body)
		h.History.Add(entry)
		h.internalErr(w, r, "search error", fmt.Errorf("status: %s, body: %s", res.Status(), string(body)))
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		entry.Error = err.Error()
		h.History.Add(entry)
		h.internalErr(w, r, "read search result", err)
		return
	}

	// Parse result to extract hits count
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		entry.Error = err.Error()
		h.History.Add(entry)
		h.internalErr(w, r, "decode search result", err)
		return
	}

	if hits, ok := result["hits"].(map[string]any); ok {
		if total, ok := hits["total"].(map[string]any); ok {
			if value, ok := total["value"].(float64); ok {
				entry.ResultCount = int(value)
			}
		}
	}

	h.History.Add(entry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// GetDocument retrieves a document by ID.
func (h ElasticsearchHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	indexName := r.URL.Query().Get("index")
	docID := r.URL.Query().Get("id")

	if err := validateIndexName(indexName, false); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}

	if docID == "" {
		h.badRequest(w, r, "document id is required")
		return
	}

	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	req := esapi.GetRequest{
		Index:      indexName,
		DocumentID: docID,
	}
	res, err := req.Do(r.Context(), client)
	if err != nil {
		h.internalErr(w, r, "get document", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			h.badRequest(w, r, "document not found")
			return
		}
		h.internalErr(w, r, "get document", fmt.Errorf("status: %s", res.Status()))
		return
	}

	var doc map[string]any
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		h.internalErr(w, r, "decode document", err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, doc, nil))
}

// DeleteDocument deletes a document by ID.
func (h ElasticsearchHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Index   string `json:"index"`
		ID      string `json:"id"`
		Confirm bool   `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if !req.Confirm {
		h.badRequest(w, r, "confirm=true is required for delete operations")
		return
	}

	if err := validateIndexName(req.Index, false); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}

	if req.ID == "" {
		h.badRequest(w, r, "document id is required")
		return
	}

	client, conn := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	if conn.Readonly {
		h.readonly(w, r)
		return
	}

	delReq := esapi.DeleteRequest{
		Index:      req.Index,
		DocumentID: req.ID,
	}
	res, err := delReq.Do(r.Context(), client)
	if err != nil {
		h.internalErr(w, r, "delete document", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			h.badRequest(w, r, "document not found")
			return
		}
		h.internalErr(w, r, "delete document", fmt.Errorf("status: %s", res.Status()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListHistory returns search history for a connection.
func (h ElasticsearchHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	entries, err := h.History.List(connID)
	if err != nil {
		h.internalErr(w, r, "list history", err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, entries, map[string]any{
		"count": len(entries),
	}))
}

// DeleteHistoryEntry deletes a single history entry.
func (h ElasticsearchHandler) DeleteHistoryEntry(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	entryID := router.Param(r, "entryId")

	if err := h.History.Delete(connID, entryID); err != nil {
		h.internalErr(w, r, "delete history entry", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearHistory clears all search history for a connection.
func (h ElasticsearchHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	if err := h.History.Clear(connID); err != nil {
		h.internalErr(w, r, "clear history", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
