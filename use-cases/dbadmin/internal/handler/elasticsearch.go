package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	Connections         *connection.Store
	ESManager           *esmanager.Manager
	History             *eshistory.Store
	Registry            *OperationRegistry
	Logger              plumelog.StructuredLogger
	QueryTimeoutSeconds int
}

func (h ElasticsearchHandler) registerOperation(cancel context.CancelFunc, kind, connID, resource, summary string) string {
	if h.Registry == nil {
		return ""
	}
	return h.Registry.Register(OperationInfo{
		Driver:   string(connection.DriverElasticsearch),
		Kind:     kind,
		ConnID:   connID,
		Resource: resource,
		Summary:  summary,
	}, cancel)
}

func (h ElasticsearchHandler) unregisterOperation(operationID string) {
	if h.Registry != nil && operationID != "" {
		h.Registry.Unregister(operationID)
	}
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
		Type(contract.TypeForbidden).
		Message("connection is read-only").
		Detail("code", "READONLY_VIOLATION").
		Build()))
}

func (h ElasticsearchHandler) timeout(ctx context.Context) (context.Context, context.CancelFunc) {
	seconds := h.QueryTimeoutSeconds
	if seconds <= 0 {
		seconds = 30
	}
	return context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
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

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	res, err := client.Info(client.Info.WithContext(ctx))
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
	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	res, err := req.Do(ctx, client)
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
	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	res, err := req.Do(ctx, client)
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
	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	res, err := req.Do(ctx, client)
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
	if !decodeJSONLimited(w, r, h.Logger, &req) {
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
		if sizeFloat > MaxESSearchSize {
			dslMap["size"] = MaxESSearchSize
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
	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "search", connID, req.Index, "search DSL")
	defer h.unregisterOperation(operationID)

	res, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(req.Index),
		client.Search.WithBody(bytes.NewReader(enforcedDSL)),
	)
	duration := time.Since(start).Milliseconds()
	h.logSlowQuery(connID, req.Index, string(enforcedDSL), duration)

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

	// Parse result to extract hits count and apply truncation
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

		// Apply truncation to hits
		if hitsArray, ok := hits["hits"].([]any); ok {
			limits := DefaultPreviewLimits()
			anyTruncated := false
			for _, hit := range hitsArray {
				if hitMap, ok := hit.(map[string]any); ok {
					if source, ok := hitMap["_source"].(map[string]any); ok {
						_, wasTruncated := limits.TruncateJSONValue(source)
						if wasTruncated {
							anyTruncated = true
						}
					}
				}
			}
			result["_truncated"] = anyTruncated
		}
	}

	h.History.Add(entry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
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
	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	res, err := req.Do(ctx, client)
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

	// Apply truncation to document _source
	limits := DefaultPreviewLimits()
	if source, ok := doc["_source"].(map[string]any); ok {
		_, wasTruncated := limits.TruncateJSONValue(source)
		doc["_truncated"] = wasTruncated
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
	if !decodeJSONLimited(w, r, h.Logger, &req) {
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
	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "delete", connID, req.Index, "delete document")
	defer h.unregisterOperation(operationID)

	res, err := delReq.Do(ctx, client)
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

// ExportDocuments exports matching documents from an index as JSON or NDJSON.
func (h ElasticsearchHandler) ExportDocuments(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	indexName := r.URL.Query().Get("index")
	format := r.URL.Query().Get("format")
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")

	if err := validateIndexName(indexName, false); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "ndjson" {
		h.badRequest(w, r, "format must be json or ndjson")
		return
	}
	limit := parseBoundedLimit(limitStr, DefaultExportRows, MaxESExportRows)

	var dsl map[string]any
	if query != "" {
		if err := json.Unmarshal([]byte(query), &dsl); err != nil {
			h.badRequest(w, r, "query must be valid JSON")
			return
		}
	} else {
		dsl = map[string]any{"query": map[string]any{"match_all": map[string]any{}}}
	}
	dsl["size"] = limit + 1
	body, err := json.Marshal(dsl)
	if err != nil {
		h.internalErr(w, r, "marshal export query", err)
		return
	}

	client, _ := h.openClient(w, r, connID)
	if client == nil {
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "export", connID, indexName, format+" export")
	defer h.unregisterOperation(operationID)

	res, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(indexName),
		client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		h.internalErr(w, r, "export search", err)
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		h.internalErr(w, r, "export search", fmt.Errorf("status: %s", res.Status()))
		return
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string         `json:"_id"`
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		h.internalErr(w, r, "decode export result", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_%s.%s\"", indexName, timestamp, format))
	w.Header().Set("X-Export-Row-Limit", strconv.Itoa(limit))

	if format == "ndjson" {
		for i, hit := range result.Hits.Hits {
			if i >= limit {
				fmt.Fprintf(w, "{\"_dbadmin_export_truncated\":true,\"limit\":%d,\"max_limit\":%d}\n", limit, MaxESExportRows)
				break
			}
			doc := hit.Source
			if doc == nil {
				doc = map[string]any{}
			}
			doc["_id"] = hit.ID
			line, _ := json.Marshal(doc)
			fmt.Fprintf(w, "%s\n", line)
		}
		return
	}

	docs := make([]map[string]any, 0, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		if i >= limit {
			docs = append(docs, map[string]any{
				"_dbadmin_export_truncated": true,
				"limit":                     limit,
				"max_limit":                 MaxESExportRows,
			})
			break
		}
		doc := hit.Source
		if doc == nil {
			doc = map[string]any{}
		}
		doc["_id"] = hit.ID
		docs = append(docs, doc)
	}
	logWriteErr(h.Logger, json.NewEncoder(w).Encode(docs))
}

// ImportDocuments imports JSON documents into an index via the Bulk API.
func (h ElasticsearchHandler) ImportDocuments(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	var req struct {
		Index     string           `json:"index"`
		Documents []map[string]any `json:"documents"`
		Confirm   bool             `json:"confirm"`
	}
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if err := validateIndexName(req.Index, false); err != nil {
		h.badRequest(w, r, err.Error())
		return
	}
	if !req.Confirm {
		h.badRequest(w, r, "confirm=true is required for import operations")
		return
	}
	if len(req.Documents) == 0 {
		h.badRequest(w, r, "no documents to import")
		return
	}
	if len(req.Documents) > MaxBulkImportDocuments {
		h.badRequest(w, r, "import is limited to "+strconv.Itoa(MaxBulkImportDocuments)+" documents")
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

	var bulk bytes.Buffer
	for _, doc := range req.Documents {
		meta := map[string]any{}
		source := make(map[string]any, len(doc))
		for k, v := range doc {
			if k == "_id" {
				if id, ok := v.(string); ok && id != "" {
					meta["_id"] = id
				}
				continue
			}
			source[k] = v
		}
		action, _ := json.Marshal(map[string]any{"index": meta})
		docBody, _ := json.Marshal(source)
		bulk.Write(action)
		bulk.WriteByte('\n')
		bulk.Write(docBody)
		bulk.WriteByte('\n')
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "import", connID, req.Index, "bulk import")
	defer h.unregisterOperation(operationID)

	bulkReq := esapi.BulkRequest{
		Index: req.Index,
		Body:  bytes.NewReader(bulk.Bytes()),
	}
	res, err := bulkReq.Do(ctx, client)
	if err != nil {
		h.internalErr(w, r, "bulk import", err)
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		h.internalErr(w, r, "bulk import", fmt.Errorf("status: %s", res.Status()))
		return
	}

	var result struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Status int            `json:"status"`
			Error  map[string]any `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		h.internalErr(w, r, "decode bulk response", err)
		return
	}
	errorsCount := 0
	for _, item := range result.Items {
		for _, op := range item {
			if op.Status >= 300 || op.Error != nil {
				errorsCount++
			}
		}
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"imported_count": len(req.Documents) - errorsCount,
		"errors":         errorsCount,
	}, nil))
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

// logSlowQuery logs queries that exceed the slow query threshold.
func (h ElasticsearchHandler) logSlowQuery(connID, index, dsl string, durationMs int64) {
	if durationMs > SlowQueryThresholdMs {
		h.Logger.Warn("slow Elasticsearch query detected", plumelog.Fields{
			"connection_id": connID,
			"index":         index,
			"duration_ms":   durationMs,
			"threshold_ms":  SlowQueryThresholdMs,
			"dsl":           truncateQueryForLog(dsl, 200),
		})
	}
}
