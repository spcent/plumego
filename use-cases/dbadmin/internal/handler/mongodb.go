package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/domain/mongohistory"
	"dbadmin/internal/mongomanager"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBHandler handles all MongoDB-specific endpoints.
type MongoDBHandler struct {
	Connections         *connection.Store
	MongoManager        *mongomanager.Manager
	History             *mongohistory.Store
	Registry            *OperationRegistry
	Logger              plumelog.StructuredLogger
	QueryTimeoutSeconds int
}

func (h MongoDBHandler) registerOperation(cancel context.CancelFunc, kind, connID, resource, summary string) string {
	if h.Registry == nil {
		return ""
	}
	return h.Registry.Register(OperationInfo{
		Driver:   string(connection.DriverMongoDB),
		Kind:     kind,
		ConnID:   connID,
		Resource: resource,
		Summary:  summary,
	}, cancel)
}

func (h MongoDBHandler) unregisterOperation(operationID string) {
	if h.Registry != nil && operationID != "" {
		h.Registry.Unregister(operationID)
	}
}

// --- helpers -----------------------------------------------------------------

func (h MongoDBHandler) openClient(connID string) (*connection.Connection, error) {
	conn, err := h.Connections.Get(connID)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (h MongoDBHandler) connNotFound(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotFound).Message("connection not found").Build()))
}

func (h MongoDBHandler) internalErr(w http.ResponseWriter, r *http.Request, err error) {
	h.Logger.Error("mongodb handler error", plumelog.Fields{"error": err.Error()})
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).Message("internal error").Build()))
}

func (h MongoDBHandler) badRequest(w http.ResponseWriter, r *http.Request, msg string) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeBadRequest).Message(msg).Build()))
}

func (h MongoDBHandler) readonly(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeForbidden).
		Message("connection is read-only").
		Detail("code", "READONLY_VIOLATION").
		Build()))
}

func (h MongoDBHandler) timeout(ctx context.Context) (context.Context, context.CancelFunc) {
	seconds := h.QueryTimeoutSeconds
	if seconds <= 0 {
		seconds = 30
	}
	return context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
}

// validateName checks if a database or collection name is valid.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 64 {
		return fmt.Errorf("name too long")
	}
	for _, c := range name {
		if c == 0 || c == '/' || c == '\\' || c == '.' || c == ' ' || c == '$' {
			return fmt.Errorf("name contains invalid character: %c", c)
		}
	}
	return nil
}

// parseID accepts only ObjectID hex strings or plain string IDs.
// This prevents user-controlled JSON values from being interpreted as other BSON scalar types.
func parseID(idStr string) (any, error) {
	if idStr == "" {
		return nil, fmt.Errorf("_id is required")
	}
	if len(idStr) == 24 {
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			return oid, nil
		}
	}
	return idStr, nil
}

// convertObjectIDs recursively converts {"$oid": "..."} to ObjectID.
func convertObjectIDs(doc bson.M) {
	for key, val := range doc {
		switch v := val.(type) {
		case map[string]any:
			if oid, ok := v["$oid"]; ok {
				if oidStr, ok := oid.(string); ok {
					if parsed, err := primitive.ObjectIDFromHex(oidStr); err == nil {
						doc[key] = parsed
					}
				}
			} else {
				convertObjectIDs(v)
			}
		case []any:
			for i, item := range v {
				if subDoc, ok := item.(map[string]any); ok {
					if oid, ok := subDoc["$oid"]; ok {
						if oidStr, ok := oid.(string); ok {
							if parsed, err := primitive.ObjectIDFromHex(oidStr); err == nil {
								v[i] = parsed
							}
						}
					} else {
						convertObjectIDs(subDoc)
					}
				}
			}
		}
	}
}

// bsonToJSON converts BSON types to JSON-friendly types.
func bsonToJSON(val any) any {
	switch v := val.(type) {
	case primitive.ObjectID:
		return map[string]any{"$oid": v.Hex()}
	case primitive.DateTime:
		return v.Time().Format(time.RFC3339)
	case primitive.Timestamp:
		return map[string]any{"$timestamp": map[string]any{"t": v.T, "i": v.I}}
	case primitive.Binary:
		return map[string]any{"$binary": map[string]any{"base64": v.Data, "subType": v.Subtype}}
	case primitive.Regex:
		return map[string]any{"$regex": v.Pattern, "$options": v.Options}
	case primitive.D:
		result := make(map[string]any)
		for _, elem := range v {
			result[elem.Key] = bsonToJSON(elem.Value)
		}
		return result
	case primitive.A:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = bsonToJSON(item)
		}
		return result
	case bson.M:
		result := make(map[string]any)
		for k, item := range v {
			result[k] = bsonToJSON(item)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = bsonToJSON(item)
		}
		return result
	case map[string]any:
		result := make(map[string]any)
		for k, item := range v {
			result[k] = bsonToJSON(item)
		}
		return result
	default:
		return val
	}
}

// --- ListDatabases -----------------------------------------------------------

// ListDatabases returns all databases in a MongoDB connection.
// GET /api/connections/:id/mongo/databases
func (h MongoDBHandler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	result, err := client.ListDatabases(ctx, bson.M{})
	if err != nil {
		h.Logger.Error("list databases failed", plumelog.Fields{
			"connection_id": connID,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	type Database struct {
		Name       string `json:"name"`
		SizeOnDisk int64  `json:"size_on_disk"`
		Empty      bool   `json:"empty"`
	}

	databases := make([]Database, 0, len(result.Databases))
	for _, db := range result.Databases {
		databases = append(databases, Database{
			Name:       db.Name,
			SizeOnDisk: db.SizeOnDisk,
			Empty:      db.Empty,
		})
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"databases": databases,
	}, nil))
}

// --- ListCollections ---------------------------------------------------------

// ListCollections returns all collections in a MongoDB database.
// GET /api/connections/:id/mongo/collections?database=xxx
func (h MongoDBHandler) ListCollections(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := r.URL.Query().Get("database")

	if err := validateName(dbName); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	cursor, err := client.Database(dbName).ListCollections(ctx, bson.M{})
	if err != nil {
		h.Logger.Error("list collections failed", plumelog.Fields{
			"connection_id": connID,
			"database":      dbName,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}
	defer cursor.Close(ctx)

	type Collection struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Options any    `json:"options,omitempty"`
	}

	collections := make([]Collection, 0)
	for cursor.Next(ctx) {
		var result struct {
			Name    string `bson:"name"`
			Type    string `bson:"type"`
			Options any    `bson:"options"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		collections = append(collections, Collection{
			Name:    result.Name,
			Type:    result.Type,
			Options: bsonToJSON(result.Options),
		})
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"collections": collections,
	}, nil))
}

// --- QueryDocuments ----------------------------------------------------------

// QueryDocuments queries documents with filter, projection, sort, limit, skip.
// POST /api/connections/:id/mongo/documents/query
func (h MongoDBHandler) QueryDocuments(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Filter     string `json:"filter"`
		Projection string `json:"projection"`
		Sort       string `json:"sort"`
		Limit      int64  `json:"limit"`
		Skip       int64  `json:"skip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "query", connID, req.Database+"."+req.Collection, "find documents")
	defer h.unregisterOperation(operationID)

	var filter bson.M
	if req.Filter != "" {
		if err := json.Unmarshal([]byte(req.Filter), &filter); err != nil {
			h.badRequest(w, r, "invalid filter JSON: "+err.Error())
			return
		}
	} else {
		filter = bson.M{}
	}

	var projection bson.M
	if req.Projection != "" {
		if err := json.Unmarshal([]byte(req.Projection), &projection); err != nil {
			h.badRequest(w, r, "invalid projection JSON: "+err.Error())
			return
		}
	}

	var sort bson.M
	if req.Sort != "" {
		if err := json.Unmarshal([]byte(req.Sort), &sort); err != nil {
			h.badRequest(w, r, "invalid sort JSON: "+err.Error())
			return
		}
	}

	opts := options.Find().
		SetLimit(req.Limit).
		SetSkip(req.Skip)

	if projection != nil {
		opts.SetProjection(projection)
	}
	if sort != nil {
		opts.SetSort(sort)
	}

	start := time.Now()
	cursor, err := client.Database(req.Database).Collection(req.Collection).Find(ctx, filter, opts)
	if err != nil {
		h.Logger.Error("query documents failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}
	defer cursor.Close(ctx)

	limits := DefaultPreviewLimits()
	documents := make([]any, 0)
	anyTruncated := false
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		jsonDoc := bsonToJSON(doc)
		truncated, wasTruncated := limits.TruncateJSONValue(jsonDoc)
		documents = append(documents, truncated)
		if wasTruncated {
			anyTruncated = true
		}
	}

	durationMs := time.Since(start).Milliseconds()
	h.logSlowQuery(connID, req.Database, req.Collection, "Find", req.Filter, durationMs)

	total, err := client.Database(req.Database).Collection(req.Collection).CountDocuments(ctx, filter)
	if err != nil {
		h.Logger.Error("count documents failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"documents": documents,
		"total":     total,
		"limit":     req.Limit,
		"skip":      req.Skip,
		"truncated": anyTruncated,
	}, nil))
}

// --- ListIndexes -------------------------------------------------------------

// ListIndexes returns all indexes for a collection.
// GET /api/connections/:id/mongo/indexes?database=xxx&collection=xxx
func (h MongoDBHandler) ListIndexes(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := r.URL.Query().Get("database")
	collName := r.URL.Query().Get("collection")

	if err := validateName(dbName); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(collName); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	cursor, err := client.Database(dbName).Collection(collName).Indexes().List(ctx)
	if err != nil {
		h.Logger.Error("list indexes failed", plumelog.Fields{
			"connection_id": connID,
			"database":      dbName,
			"collection":    collName,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}
	defer cursor.Close(ctx)

	indexes := make([]any, 0)
	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			continue
		}
		indexes = append(indexes, bsonToJSON(idx))
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"indexes": indexes,
	}, nil))
}

// --- InsertDocument ----------------------------------------------------------

// InsertDocument inserts a new document.
// POST /api/connections/:id/mongo/documents
func (h MongoDBHandler) InsertDocument(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Document   string `json:"document"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	if conn.Readonly {
		h.readonly(w, r)
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	var doc bson.M
	if err := json.Unmarshal([]byte(req.Document), &doc); err != nil {
		h.badRequest(w, r, "invalid document JSON: "+err.Error())
		return
	}

	convertObjectIDs(doc)

	result, err := client.Database(req.Database).Collection(req.Collection).InsertOne(ctx, doc)
	if err != nil {
		h.Logger.Error("insert document failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, map[string]any{
		"inserted_id": bsonToJSON(result.InsertedID),
	}, nil))
}

// --- UpdateDocument ----------------------------------------------------------

// UpdateDocument updates a document by _id.
// PATCH /api/connections/:id/mongo/documents
func (h MongoDBHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		ID         string `json:"id"`
		Document   string `json:"document"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if req.ID == "" {
		h.badRequest(w, r, "document _id is required")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	if conn.Readonly {
		h.readonly(w, r)
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	var doc bson.M
	if err := json.Unmarshal([]byte(req.Document), &doc); err != nil {
		h.badRequest(w, r, "invalid document JSON: "+err.Error())
		return
	}

	convertObjectIDs(doc)

	filterID, err := parseID(req.ID)
	if err != nil {
		h.badRequest(w, r, "invalid _id: "+err.Error())
		return
	}

	result, err := client.Database(req.Database).Collection(req.Collection).UpdateOne(
		ctx,
		bson.M{"_id": filterID},
		bson.M{"$set": doc},
	)
	if err != nil {
		h.Logger.Error("update document failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"matched_count":  result.MatchedCount,
		"modified_count": result.ModifiedCount,
	}, nil))
}

// --- DeleteDocument ----------------------------------------------------------

// DeleteDocument deletes a document by _id.
// DELETE /api/connections/:id/mongo/documents
func (h MongoDBHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		ID         string `json:"id"`
		Confirm    bool   `json:"confirm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if !req.Confirm {
		h.badRequest(w, r, "confirm=true is required")
		return
	}

	if req.ID == "" {
		h.badRequest(w, r, "document _id is required")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	if conn.Readonly {
		h.readonly(w, r)
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	filterID, err := parseID(req.ID)
	if err != nil {
		h.badRequest(w, r, "invalid _id: "+err.Error())
		return
	}

	result, err := client.Database(req.Database).Collection(req.Collection).DeleteOne(
		ctx,
		bson.M{"_id": filterID},
	)
	if err != nil {
		h.Logger.Error("delete document failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"deleted_count": result.DeletedCount,
	}, nil))
}

// Aggregate executes a MongoDB aggregation pipeline.
// POST /api/connections/:id/mongo/aggregate
func (h MongoDBHandler) Aggregate(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Pipeline   string `json:"pipeline"` // JSON array string
		Limit      int    `json:"limit"`
		Confirm    bool   `json:"confirm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	// Default limit: 100, max: 1000
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	// Parse pipeline JSON array
	var pipelineArray []any
	if err := json.Unmarshal([]byte(req.Pipeline), &pipelineArray); err != nil {
		h.badRequest(w, r, "invalid pipeline JSON: "+err.Error())
		return
	}

	// Check for dangerous stages ($out, $merge)
	hasDangerous := false
	for _, stage := range pipelineArray {
		if stageMap, ok := stage.(map[string]any); ok {
			if _, hasOut := stageMap["$out"]; hasOut {
				hasDangerous = true
				break
			}
			if _, hasMerge := stageMap["$merge"]; hasMerge {
				hasDangerous = true
				break
			}
		}
	}

	// Block dangerous operations in readonly mode
	if conn.Readonly && hasDangerous {
		h.readonly(w, r)
		return
	}

	// Require confirmation for dangerous operations
	if hasDangerous && !req.Confirm {
		h.badRequest(w, r, "pipeline contains dangerous stages ($out or $merge); confirm=true required")
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "aggregate", connID, req.Database+"."+req.Collection, "aggregation pipeline")
	defer h.unregisterOperation(operationID)

	// Convert pipeline to bson.A
	pipeline := bson.A{}
	for _, stage := range pipelineArray {
		pipeline = append(pipeline, stage)
	}

	startTime := time.Now()
	cursor, err := client.Database(req.Database).Collection(req.Collection).Aggregate(ctx, pipeline)
	if err != nil {
		h.Logger.Error("aggregate failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		// Save to history with error
		duration := time.Since(startTime).Milliseconds()
		h.savePipelineHistory(connID, req.Database, req.Collection, req.Pipeline, duration, 0, err.Error())
		h.internalErr(w, r, err)
		return
	}
	defer cursor.Close(ctx)

	limits := DefaultPreviewLimits()
	var documents []any
	anyTruncated := false
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		jsonDoc := bsonToJSON(doc)
		truncated, wasTruncated := limits.TruncateJSONValue(jsonDoc)
		documents = append(documents, truncated)
		if wasTruncated {
			anyTruncated = true
		}
		if len(documents) >= req.Limit {
			break
		}
	}

	duration := time.Since(startTime).Milliseconds()
	h.savePipelineHistory(connID, req.Database, req.Collection, req.Pipeline, duration, len(documents), "")
	h.logSlowQuery(connID, req.Database, req.Collection, "Aggregate", req.Pipeline, duration)

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"documents":   documents,
		"count":       len(documents),
		"duration_ms": duration,
		"truncated":   anyTruncated,
	}, nil))
}

// savePipelineHistory saves an aggregation pipeline execution to history.
func (h MongoDBHandler) savePipelineHistory(connID, database, collection, pipeline string, duration int64, resultCount int, errMsg string) {
	if h.History == nil {
		return
	}
	entry := &mongohistory.Entry{
		ID:          generateID(),
		ConnID:      connID,
		Database:    database,
		Collection:  collection,
		Pipeline:    pipeline,
		Duration:    duration,
		ResultCount: resultCount,
		Error:       errMsg,
		CreatedAt:   time.Now(),
	}
	if err := h.History.Add(entry); err != nil {
		h.Logger.Error("save pipeline history failed", plumelog.Fields{"error": err.Error()})
	}
}

// generateID generates a unique ID for history entries.
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// logSlowQuery logs queries that exceed the slow query threshold.
func (h MongoDBHandler) logSlowQuery(connID, database, collection, operation, query string, durationMs int64) {
	if durationMs > SlowQueryThresholdMs {
		h.Logger.Warn("slow MongoDB query detected", plumelog.Fields{
			"connection_id": connID,
			"database":      database,
			"collection":    collection,
			"operation":     operation,
			"duration_ms":   durationMs,
			"threshold_ms":  SlowQueryThresholdMs,
			"query":         truncateQueryForLog(query, 200),
		})
	}
}

// truncateQueryForLog truncates query for logging to prevent excessively long log entries.
func truncateQueryForLog(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "... (truncated)"
}

// ExplainQuery shows the query execution plan for a find operation.
// POST /api/connections/:id/mongo/explain
func (h MongoDBHandler) ExplainQuery(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Filter     string `json:"filter"`
		Sort       string `json:"sort"`
		Projection string `json:"projection"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	// Build find command for explain
	filter := bson.M{}
	if req.Filter != "" {
		if err := json.Unmarshal([]byte(req.Filter), &filter); err != nil {
			h.badRequest(w, r, "invalid filter JSON: "+err.Error())
			return
		}
	}

	var sort bson.M
	if req.Sort != "" {
		if err := json.Unmarshal([]byte(req.Sort), &sort); err != nil {
			h.badRequest(w, r, "invalid sort JSON: "+err.Error())
			return
		}
	}

	var projection bson.M
	if req.Projection != "" {
		if err := json.Unmarshal([]byte(req.Projection), &projection); err != nil {
			h.badRequest(w, r, "invalid projection JSON: "+err.Error())
			return
		}
	}

	// Use aggregation with $explain or runCommand
	// Build the find command
	findCmd := bson.D{
		{Key: "find", Value: req.Collection},
		{Key: "filter", Value: filter},
		{Key: "limit", Value: 1},
	}

	if sort != nil {
		findCmd = append(findCmd, bson.E{Key: "sort", Value: sort})
	}
	if projection != nil {
		findCmd = append(findCmd, bson.E{Key: "projection", Value: projection})
	}

	explainCmd := bson.D{
		{Key: "explain", Value: findCmd},
		{Key: "verbosity", Value: "executionStats"},
	}

	var result bson.M
	if err := client.Database(req.Database).RunCommand(ctx, explainCmd).Decode(&result); err != nil {
		h.Logger.Error("explain query failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"explain": bsonToJSON(result),
	}, nil))
}

// SchemaSample analyzes the schema of a collection by sampling documents.
// GET /api/connections/:id/mongo/schema?database=X&collection=Y&sample=N
func (h MongoDBHandler) SchemaSample(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := r.URL.Query().Get("database")
	collName := r.URL.Query().Get("collection")
	sampleStr := r.URL.Query().Get("sample")

	if err := validateName(dbName); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(collName); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	sample := 100
	if sampleStr != "" {
		if s, err := strconv.Atoi(sampleStr); err == nil && s > 0 {
			sample = s
		}
	}
	if sample > 1000 {
		sample = 1000
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	// Use $sample aggregation to get random documents
	pipeline := bson.A{
		bson.M{"$sample": bson.M{"size": sample}},
	}

	cursor, err := client.Database(dbName).Collection(collName).Aggregate(ctx, pipeline)
	if err != nil {
		h.Logger.Error("schema sample failed", plumelog.Fields{
			"connection_id": connID,
			"database":      dbName,
			"collection":    collName,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}
	defer cursor.Close(ctx)

	// Analyze field types
	fieldStats := make(map[string]map[string]int) // field -> type -> count
	fieldSamples := make(map[string][]any)        // field -> sample values

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		analyzeDocument(doc, "", fieldStats, fieldSamples)
	}

	// Build response
	var fields []map[string]any
	for fieldName, typeCounts := range fieldStats {
		var types []map[string]any
		for typeName, count := range typeCounts {
			types = append(types, map[string]any{
				"type":  typeName,
				"count": count,
			})
		}

		samples := fieldSamples[fieldName]
		if len(samples) > 5 {
			samples = samples[:5]
		}

		fields = append(fields, map[string]any{
			"name":          fieldName,
			"types":         types,
			"sample_values": samples,
		})
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"fields":        fields,
		"sampled_count": sample,
	}, nil))
}

// analyzeDocument recursively analyzes document fields.
func analyzeDocument(doc bson.M, prefix string, fieldStats map[string]map[string]int, fieldSamples map[string][]any) {
	for key, value := range doc {
		fieldName := key
		if prefix != "" {
			fieldName = prefix + "." + key
		}

		typeName := getTypeName(value)

		if fieldStats[fieldName] == nil {
			fieldStats[fieldName] = make(map[string]int)
		}
		fieldStats[fieldName][typeName]++

		// Store sample values (limit to 10)
		if len(fieldSamples[fieldName]) < 10 {
			fieldSamples[fieldName] = append(fieldSamples[fieldName], bsonToJSON(value))
		}

		// Recurse into nested documents
		if nested, ok := value.(bson.M); ok {
			analyzeDocument(nested, fieldName, fieldStats, fieldSamples)
		}
	}
}

// getTypeName returns the type name of a BSON value.
func getTypeName(value any) string {
	switch value.(type) {
	case primitive.ObjectID:
		return "ObjectId"
	case string:
		return "String"
	case int32:
		return "Int32"
	case int64:
		return "Int64"
	case float64:
		return "Double"
	case bool:
		return "Boolean"
	case primitive.DateTime:
		return "Date"
	case primitive.Timestamp:
		return "Timestamp"
	case primitive.Regex:
		return "Regex"
	case primitive.Binary:
		return "Binary"
	case bson.M:
		return "Object"
	case bson.A:
		return "Array"
	case nil:
		return "Null"
	default:
		return fmt.Sprintf("%T", value)
	}
}

// CollectionStats returns statistics for a collection.
// GET /api/connections/:id/mongo/stats?database=X&collection=Y
func (h MongoDBHandler) CollectionStats(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := r.URL.Query().Get("database")
	collName := r.URL.Query().Get("collection")

	if err := validateName(dbName); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(collName); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()

	// Run collStats command
	cmd := bson.D{
		{Key: "collStats", Value: collName},
		{Key: "scale", Value: 1},
	}

	var result bson.M
	if err := client.Database(dbName).RunCommand(ctx, cmd).Decode(&result); err != nil {
		h.Logger.Error("collection stats failed", plumelog.Fields{
			"connection_id": connID,
			"database":      dbName,
			"collection":    collName,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	// Extract relevant stats
	count := int64(0)
	size := int64(0)
	avgObjSize := int64(0)
	storageSize := int64(0)
	totalIndexSize := int64(0)

	if v, ok := result["count"].(int32); ok {
		count = int64(v)
	} else if v, ok := result["count"].(int64); ok {
		count = v
	}
	if v, ok := result["size"].(int64); ok {
		size = v
	}
	if v, ok := result["avgObjSize"].(int64); ok {
		avgObjSize = v
	}
	if v, ok := result["storageSize"].(int64); ok {
		storageSize = v
	}
	if v, ok := result["totalIndexSize"].(int64); ok {
		totalIndexSize = v
	}

	// Extract index info
	var indexes []map[string]any
	if indexSizes, ok := result["indexSizes"].(bson.M); ok {
		for name, sizeVal := range indexSizes {
			sizeNum := int64(0)
			if v, ok := sizeVal.(int64); ok {
				sizeNum = v
			}
			indexes = append(indexes, map[string]any{
				"name": name,
				"size": sizeNum,
			})
		}
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"count":          count,
		"size":           size,
		"avgObjSize":     avgObjSize,
		"storageSize":    storageSize,
		"totalIndexSize": totalIndexSize,
		"indexes":        indexes,
	}, nil))
}

// ExportDocuments exports documents in JSON, NDJSON, or CSV format.
// GET /api/connections/:id/mongo/export?database=X&collection=Y&format=json|ndjson|csv&filter=...&limit=N
func (h MongoDBHandler) ExportDocuments(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := r.URL.Query().Get("database")
	collName := r.URL.Query().Get("collection")
	format := r.URL.Query().Get("format")
	filterStr := r.URL.Query().Get("filter")
	limitStr := r.URL.Query().Get("limit")

	if err := validateName(dbName); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(collName); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	if format == "" {
		format = "json"
	}
	if format != "json" && format != "ndjson" && format != "csv" {
		h.badRequest(w, r, "format must be json, ndjson, or csv")
		return
	}

	limit := 10000
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 100000 {
		limit = 100000
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "export", connID, dbName+"."+collName, format+" export")
	defer h.unregisterOperation(operationID)

	filter := bson.M{}
	if filterStr != "" {
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
			h.badRequest(w, r, "invalid filter JSON: "+err.Error())
			return
		}
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := client.Database(dbName).Collection(collName).Find(ctx, filter, opts)
	if err != nil {
		h.Logger.Error("export documents failed", plumelog.Fields{
			"connection_id": connID,
			"database":      dbName,
			"collection":    collName,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}
	defer cursor.Close(ctx)

	// Set response headers for download
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.%s", collName, timestamp, format)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	switch format {
	case "json":
		h.exportJSON(ctx, w, cursor)
	case "ndjson":
		h.exportNDJSON(ctx, w, cursor)
	case "csv":
		h.exportCSV(ctx, w, cursor)
	}
}

func (h MongoDBHandler) exportJSON(ctx context.Context, w http.ResponseWriter, cursor *mongo.Cursor) {
	fmt.Fprint(w, "[\n")
	first := true
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		if !first {
			fmt.Fprint(w, ",\n")
		}
		first = false
		jsonBytes, _ := json.Marshal(bsonToJSON(doc))
		fmt.Fprintf(w, "  %s", string(jsonBytes))
	}
	fmt.Fprint(w, "\n]\n")
}

func (h MongoDBHandler) exportNDJSON(ctx context.Context, w http.ResponseWriter, cursor *mongo.Cursor) {
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		jsonBytes, _ := json.Marshal(bsonToJSON(doc))
		fmt.Fprintf(w, "%s\n", string(jsonBytes))
	}
}

func (h MongoDBHandler) exportCSV(ctx context.Context, w http.ResponseWriter, cursor *mongo.Cursor) {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	headerWritten := false
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// Write header on first document
		if !headerWritten {
			var headers []string
			for key := range doc {
				headers = append(headers, key)
			}
			sort.Strings(headers)
			writer.Write(headers)
			headerWritten = true
		}

		// Build CSV row
		var row []string
		for _, key := range getSortedKeys(doc) {
			value := doc[key]
			row = append(row, formatCSVValue(value))
		}
		writer.Write(row)
	}
}

func getSortedKeys(doc bson.M) []string {
	keys := make([]string, 0, len(doc))
	for key := range doc {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func formatCSVValue(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case primitive.ObjectID:
		return v.Hex()
	case primitive.DateTime:
		return v.Time().Format(time.RFC3339)
	default:
		// For complex types, marshal to JSON
		jsonBytes, _ := json.Marshal(bsonToJSON(value))
		return string(jsonBytes)
	}
}

// ImportDocuments imports documents from JSON or NDJSON.
// POST /api/connections/:id/mongo/import
func (h MongoDBHandler) ImportDocuments(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	var req struct {
		Database   string `json:"database"`
		Collection string `json:"collection"`
		Documents  []any  `json:"documents"`
		Confirm    bool   `json:"confirm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, r, "invalid request body")
		return
	}

	if err := validateName(req.Database); err != nil {
		h.badRequest(w, r, "invalid database name: "+err.Error())
		return
	}

	if err := validateName(req.Collection); err != nil {
		h.badRequest(w, r, "invalid collection name: "+err.Error())
		return
	}

	if !req.Confirm {
		h.badRequest(w, r, "confirm=true required")
		return
	}

	if len(req.Documents) == 0 {
		h.badRequest(w, r, "no documents to import")
		return
	}

	conn, err := h.openClient(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
		} else {
			h.internalErr(w, r, err)
		}
		return
	}

	if conn.Driver != connection.DriverMongoDB {
		h.badRequest(w, r, "connection is not MongoDB")
		return
	}

	if conn.Readonly {
		h.readonly(w, r)
		return
	}

	client, err := h.MongoManager.Open(conn)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := h.timeout(r.Context())
	defer cancel()
	operationID := h.registerOperation(cancel, "import", connID, req.Database+"."+req.Collection, "import documents")
	defer h.unregisterOperation(operationID)

	// Convert documents to bson.M
	docs := make([]any, len(req.Documents))
	for i, doc := range req.Documents {
		if docMap, ok := doc.(map[string]any); ok {
			bsonDoc := bson.M{}
			for k, v := range docMap {
				bsonDoc[k] = v
			}
			convertObjectIDs(bsonDoc)
			docs[i] = bsonDoc
		} else {
			h.badRequest(w, r, "invalid document format at index "+strconv.Itoa(i))
			return
		}
	}

	result, err := client.Database(req.Database).Collection(req.Collection).InsertMany(ctx, docs)
	if err != nil {
		h.Logger.Error("import documents failed", plumelog.Fields{
			"connection_id": connID,
			"database":      req.Database,
			"collection":    req.Collection,
			"error":         err.Error(),
		})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"inserted_count": len(result.InsertedIDs),
	}, nil))
}

// ParseObjectId parses an ObjectId and returns its components.
// GET /api/mongo/objectid/:id/parse
func (h MongoDBHandler) ParseObjectId(w http.ResponseWriter, r *http.Request) {
	idStr := router.Param(r, "id")

	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		h.badRequest(w, r, "invalid ObjectId: "+err.Error())
		return
	}

	// Extract timestamp from ObjectId
	timestamp := oid.Timestamp()

	// Extract counter (last 3 bytes)
	counter := int(oid[9])<<16 | int(oid[10])<<8 | int(oid[11])

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"hex":       oid.Hex(),
		"timestamp": timestamp.Format(time.RFC3339),
		"counter":   counter,
	}, nil))
}

// ListHistory returns the aggregation pipeline history for a connection.
// GET /api/connections/:id/mongo/history
func (h MongoDBHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	if h.History == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, []any{}, nil))
		return
	}

	entries, err := h.History.List(connID)
	if err != nil {
		h.Logger.Error("list mongo history failed", plumelog.Fields{"error": err.Error()})
		h.internalErr(w, r, err)
		return
	}

	if entries == nil {
		entries = []mongohistory.Entry{}
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, entries, nil))
}

// DeleteHistoryEntry removes a single history entry.
// DELETE /api/connections/:id/mongo/history/:entryId
func (h MongoDBHandler) DeleteHistoryEntry(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	entryID := router.Param(r, "entryId")

	if h.History == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusNoContent, nil, nil))
		return
	}

	if err := h.History.Delete(connID, entryID); err != nil {
		h.Logger.Error("delete mongo history entry failed", plumelog.Fields{"error": err.Error()})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusNoContent, nil, nil))
}

// ClearHistory removes all history entries for a connection.
// DELETE /api/connections/:id/mongo/history
func (h MongoDBHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")

	if h.History == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusNoContent, nil, nil))
		return
	}

	if err := h.History.Clear(connID); err != nil {
		h.Logger.Error("clear mongo history failed", plumelog.Fields{"error": err.Error()})
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusNoContent, nil, nil))
}
