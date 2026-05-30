package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/redismanager"
)

// RedisHandler handles all Redis-specific endpoints.
type RedisHandler struct {
	Connections  *connection.Store
	RedisManager *redismanager.Manager
	Logger       plumelog.StructuredLogger
}

// --- write commands that must be blocked in readonly mode --------------------

var redisWriteCommands = map[string]bool{
	"SET": true, "MSET": true, "MSETNX": true, "SETEX": true, "PSETEX": true,
	"SETNX": true, "GETSET": true, "GETDEL": true, "GETEX": true,
	"DEL": true, "UNLINK": true,
	"EXPIRE": true, "EXPIREAT": true, "PEXPIRE": true, "PEXPIREAT": true,
	"PERSIST": true, "EXPIRETIME": true, "PEXPIRETIME": true,
	"HSET": true, "HMSET": true, "HSETNX": true, "HDEL": true,
	"LPUSH": true, "RPUSH": true, "LPUSHX": true, "RPUSHX": true,
	"LPOP": true, "RPOP": true, "LSET": true, "LINSERT": true,
	"LREM": true, "LTRIM": true,
	"SADD": true, "SREM": true, "SMOVE": true, "SPOP": true,
	"ZADD": true, "ZREM": true, "ZINCRBY": true, "ZRANGESTORE": true,
	"ZPOPMIN": true, "ZPOPMAX": true,
	"FLUSHDB": true, "FLUSHALL": true,
	"RENAME": true, "RENAMENX": true,
	"COPY": true, "MOVE": true,
	"INCR": true, "INCRBY": true, "INCRBYFLOAT": true,
	"DECR": true, "DECRBY": true,
	"APPEND":   true,
	"SETRANGE": true,
	"XADD":     true, "XTRIM": true, "XDEL": true,
}

// forbiddenCommands are never allowed regardless of readonly status.
var forbiddenCommands = map[string]bool{
	"KEYS":      true, // O(N) scan — must use SCAN
	"SHUTDOWN":  true,
	"DEBUG":     true,
	"CONFIG":    true,
	"SLAVEOF":   true,
	"REPLICAOF": true,
}

// --- helpers -----------------------------------------------------------------

func (h RedisHandler) openClient(connID string, dbIndex int) (*connection.Connection, *redis.Client, error) {
	conn, err := h.Connections.Get(connID)
	if err != nil {
		return nil, nil, err
	}
	cl, err := h.RedisManager.Open(conn, dbIndex)
	if err != nil {
		return conn, nil, err
	}
	return conn, cl, nil
}

func (h RedisHandler) connNotFound(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotFound).Message("connection not found").Build()))
}

func (h RedisHandler) internalErr(w http.ResponseWriter, r *http.Request, err error) {
	h.Logger.Error("redis handler error", plumelog.Fields{"error": err.Error()})
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).Message("internal error").Build()))
}

func dbIndexParam(r *http.Request) (int, error) {
	s := router.Param(r, "dbIndex")
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 15 {
		return 0, fmt.Errorf("invalid dbIndex: must be 0-15")
	}
	return n, nil
}

// --- ListDBs -----------------------------------------------------------------

type redisDB struct {
	Index int   `json:"index"`
	Keys  int64 `json:"keys"`
}

// ListDBs lists Redis databases (0-15) with their key counts.
// Uses CONFIG GET databases to determine the configured count, falls back to 16.
func (h RedisHandler) ListDBs(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	conn, err := h.Connections.Get(connID)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}
	// Open on DB 0 to query server info.
	cl, err := h.RedisManager.Open(conn, 0)
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	dbCount := 16
	cfgRes, err := cl.ConfigGet(ctx, "databases").Result()
	if err == nil {
		if v, ok := cfgRes["databases"]; ok {
			if n, err2 := strconv.Atoi(v); err2 == nil && n > 0 {
				dbCount = n
			}
		}
	}

	// Get key counts via INFO keyspace.
	keyCounts := map[int]int64{}
	info, err := cl.Info(ctx, "keyspace").Result()
	if err == nil {
		for _, line := range strings.Split(info, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "db") {
				continue
			}
			colon := strings.Index(line, ":")
			if colon < 0 {
				continue
			}
			idxStr := line[2:colon]
			idx, err2 := strconv.Atoi(idxStr)
			if err2 != nil {
				continue
			}
			// line format: db0:keys=12345,expires=0,avg_ttl=0
			fields := strings.Split(line[colon+1:], ",")
			for _, f := range fields {
				if strings.HasPrefix(f, "keys=") {
					if n, err3 := strconv.ParseInt(f[5:], 10, 64); err3 == nil {
						keyCounts[idx] = n
					}
				}
			}
		}
	}

	dbs := make([]redisDB, dbCount)
	for i := range dbs {
		dbs[i] = redisDB{Index: i, Keys: keyCounts[i]}
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"databases": dbs}, nil))
}

// --- ListKeys ----------------------------------------------------------------

type keyEntry struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	TTL    int64  `json:"ttl"`              // seconds; -1 = no expiry
	Memory int64  `json:"memory,omitempty"` // bytes; 0 if unavailable
	IsBig  bool   `json:"isBig,omitempty"`  // true if memory > 1MB
}

type listKeysResponse struct {
	Keys       []keyEntry `json:"keys"`
	NextCursor uint64     `json:"nextCursor"`
	Done       bool       `json:"done"`
}

// ListKeys paginates keys via SCAN (never KEYS). Returns up to `count` keys
// matching the pattern, starting from cursor.
func (h RedisHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)

	q := r.URL.Query()
	pattern := q.Get("pattern")
	if pattern == "" {
		pattern = "*"
	}
	cursorStr := q.Get("cursor")
	cursor, _ := strconv.ParseUint(cursorStr, 10, 64)
	countStr := q.Get("count")
	count, _ := strconv.ParseInt(countStr, 10, 64)
	if count <= 0 || count > 500 {
		count = 100
	}

	conn, cl, err := h.openClient(connID, dbIndex)
	_ = conn
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	keys, nextCursor, err := cl.Scan(ctx, cursor, pattern, count).Result()
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	// Fetch types, TTLs, and memory usage in pipeline.
	entries := make([]keyEntry, len(keys))
	if len(keys) > 0 {
		pipe := cl.Pipeline()
		typeCmds := make([]*redis.StatusCmd, len(keys))
		ttlCmds := make([]*redis.DurationCmd, len(keys))
		memCmds := make([]*redis.IntCmd, len(keys))
		for i, k := range keys {
			typeCmds[i] = pipe.Type(ctx, k)
			ttlCmds[i] = pipe.TTL(ctx, k)
			memCmds[i] = pipe.MemoryUsage(ctx, k, 0) // 0 = default samples
		}
		pipe.Exec(ctx) //nolint:errcheck — individual cmd errors handled below
		for i, k := range keys {
			t := "unknown"
			if typeCmds[i].Err() == nil {
				t = typeCmds[i].Val()
			}
			ttlSec := int64(-1)
			if ttlCmds[i].Err() == nil {
				d := ttlCmds[i].Val()
				if d == -1*time.Second {
					ttlSec = -1
				} else if d == -2*time.Second {
					ttlSec = -2 // key not found / expired
				} else {
					ttlSec = int64(d.Seconds())
				}
			}
			var memory int64
			if memCmds[i].Err() == nil {
				memory = memCmds[i].Val()
			}
			// Mark as big key if > 1MB
			isBig := memory > 1024*1024
			entries[i] = keyEntry{Key: k, Type: t, TTL: ttlSec, Memory: memory, IsBig: isBig}
		}
	}

	resp := listKeysResponse{
		Keys:       entries,
		NextCursor: nextCursor,
		Done:       nextCursor == 0,
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, resp, nil))
}

// --- GetKey ------------------------------------------------------------------

type keyDetail struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	TTL      int64  `json:"ttl"` // seconds; -1 = no expiry
	Encoding string `json:"encoding,omitempty"`
	// type-specific payloads
	StringVal *string           `json:"string,omitempty"`
	HashVal   map[string]string `json:"hash,omitempty"`
	ListVal   []string          `json:"list,omitempty"`
	SetVal    []string          `json:"set,omitempty"`
	ZSetVal   []zsetMember      `json:"zset,omitempty"`
	StreamVal *streamInfo       `json:"stream,omitempty"`
	// truncation flags
	StringTruncated bool `json:"stringTruncated,omitempty"`
	ValueTruncated  bool `json:"valueTruncated,omitempty"`
}

type zsetMember struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

type streamMessage struct {
	ID     string            `json:"id"`
	Values map[string]string `json:"values"`
}

type streamInfo struct {
	Length   int64           `json:"length"`
	Groups   int64           `json:"groups"`
	Messages []streamMessage `json:"messages,omitempty"`
}

// GetKey returns full details for a single key including type-appropriate value.
func (h RedisHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)
	key := r.URL.Query().Get("key")
	if key == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("key is required").Build()))
		return
	}

	_, cl, err := h.openClient(connID, dbIndex)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	keyType, err := cl.Type(ctx, key).Result()
	if err != nil {
		h.internalErr(w, r, err)
		return
	}
	if keyType == "none" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("key not found").Build()))
		return
	}

	ttlSec := int64(-1)
	ttl, err := cl.TTL(ctx, key).Result()
	if err == nil {
		if ttl == -1*time.Second {
			ttlSec = -1
		} else if ttl >= 0 {
			ttlSec = int64(ttl.Seconds())
		}
	}

	detail := keyDetail{Key: key, Type: keyType, TTL: ttlSec}
	limits := DefaultPreviewLimits()

	switch keyType {
	case "string":
		val, err := cl.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			h.internalErr(w, r, err)
			return
		}
		// Truncate large strings
		result := limits.PreviewValue(val)
		truncated := result.Value.(string)
		detail.StringVal = &truncated
		detail.StringTruncated = result.Truncated

	case "hash":
		vals, err := cl.HGetAll(ctx, key).Result()
		if err != nil {
			h.internalErr(w, r, err)
			return
		}
		// Truncate large hash values
		truncatedVals := make(map[string]string)
		for k, v := range vals {
			result := limits.PreviewValue(v)
			truncatedVals[k] = result.Value.(string)
			if result.Truncated {
				detail.ValueTruncated = true
			}
		}
		detail.HashVal = truncatedVals

	case "list":
		vals, err := cl.LRange(ctx, key, 0, 999).Result()
		if err != nil {
			h.internalErr(w, r, err)
			return
		}
		// Truncate large list values
		truncatedVals := make([]string, len(vals))
		for i, v := range vals {
			result := limits.PreviewValue(v)
			truncatedVals[i] = result.Value.(string)
			if result.Truncated {
				detail.ValueTruncated = true
			}
		}
		detail.ListVal = truncatedVals

	case "set":
		vals, err := cl.SMembers(ctx, key).Result()
		if err != nil {
			h.internalErr(w, r, err)
			return
		}
		// Truncate large set values
		truncatedVals := make([]string, len(vals))
		for i, v := range vals {
			result := limits.PreviewValue(v)
			truncatedVals[i] = result.Value.(string)
			if result.Truncated {
				detail.ValueTruncated = true
			}
		}
		detail.SetVal = truncatedVals

	case "zset":
		members, err := cl.ZRangeWithScores(ctx, key, 0, 999).Result()
		if err != nil {
			h.internalErr(w, r, err)
			return
		}
		zm := make([]zsetMember, len(members))
		for i, m := range members {
			memberStr := fmt.Sprintf("%v", m.Member)
			result := limits.PreviewValue(memberStr)
			zm[i] = zsetMember{Member: result.Value.(string), Score: m.Score}
			if result.Truncated {
				detail.ValueTruncated = true
			}
		}
		detail.ZSetVal = zm

	case "stream":
		length, _ := cl.XLen(ctx, key).Result()
		groups, _ := cl.XInfoGroups(ctx, key).Result()
		// Fetch first 100 messages for read-only viewing
		messages, _ := cl.XRangeN(ctx, key, "-", "+", 100).Result()
		msgs := make([]streamMessage, len(messages))
		for i, m := range messages {
			vals := make(map[string]string)
			for k, v := range m.Values {
				valStr := fmt.Sprintf("%v", v)
				result := limits.PreviewValue(valStr)
				vals[k] = result.Value.(string)
				if result.Truncated {
					detail.ValueTruncated = true
				}
			}
			msgs[i] = streamMessage{ID: m.ID, Values: vals}
		}
		detail.StreamVal = &streamInfo{Length: length, Groups: int64(len(groups)), Messages: msgs}
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, detail, nil))
}

// --- SetTTL ------------------------------------------------------------------

type setTTLRequest struct {
	Key string `json:"key"`
	TTL int64  `json:"ttl"` // seconds; -1 = remove expiry
}

// SetTTL sets or removes the TTL on a key.
func (h RedisHandler) SetTTL(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)

	var req setTTLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if req.Key == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("key is required").Build()))
		return
	}

	conn, cl, err := h.openClient(connID, dbIndex)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if req.TTL < 0 {
		if err := cl.Persist(ctx, req.Key).Err(); err != nil {
			h.internalErr(w, r, err)
			return
		}
	} else {
		if err := cl.Expire(ctx, req.Key, time.Duration(req.TTL)*time.Second).Err(); err != nil {
			h.internalErr(w, r, err)
			return
		}
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"ok": true}, nil))
}

// --- DeleteKey ---------------------------------------------------------------

type deleteKeyRequest struct {
	Key     string `json:"key"`
	Confirm bool   `json:"confirm"`
}

// DeleteKey deletes a single key. Requires confirm=true in the request body.
func (h RedisHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)

	var req deleteKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if req.Key == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("key is required").Build()))
		return
	}
	if !req.Confirm {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required to delete a key").Build()))
		return
	}

	conn, cl, err := h.openClient(connID, dbIndex)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := cl.Del(ctx, req.Key).Err(); err != nil {
		h.internalErr(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"ok": true}, nil))
}

// --- BatchPreview ------------------------------------------------------------

type batchPreviewRequest struct {
	Pattern string `json:"pattern"`
	MaxKeys int    `json:"maxKeys"` // default 100, max 1000
}

type batchPreviewResponse struct {
	Keys      []keyEntry `json:"keys"`
	TotalKeys int        `json:"totalKeys"`
	Truncated bool       `json:"truncated"` // true if more keys match than maxKeys
}

// BatchPreview scans keys matching a pattern and returns them without deleting.
// This allows the user to preview what will be deleted before confirming.
func (h RedisHandler) BatchPreview(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)

	var req batchPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if req.Pattern == "" {
		req.Pattern = "*"
	}
	if req.MaxKeys <= 0 {
		req.MaxKeys = 100
	} else if req.MaxKeys > 1000 {
		req.MaxKeys = 1000
	}

	_, cl, err := h.openClient(connID, dbIndex)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var allKeys []keyEntry
	var cursor uint64
	truncated := false

	// Scan in batches to collect keys
	for {
		keys, nextCursor, err := cl.Scan(ctx, cursor, req.Pattern, 100).Result()
		if err != nil {
			h.internalErr(w, r, err)
			return
		}

		// Fetch metadata for this batch
		if len(keys) > 0 {
			pipe := cl.Pipeline()
			typeCmds := make([]*redis.StatusCmd, len(keys))
			ttlCmds := make([]*redis.DurationCmd, len(keys))
			memCmds := make([]*redis.IntCmd, len(keys))
			for i, k := range keys {
				typeCmds[i] = pipe.Type(ctx, k)
				ttlCmds[i] = pipe.TTL(ctx, k)
				memCmds[i] = pipe.MemoryUsage(ctx, k, 0)
			}
			pipe.Exec(ctx) //nolint:errcheck

			for i, k := range keys {
				if len(allKeys) >= req.MaxKeys {
					truncated = true
					break
				}

				t := "unknown"
				if typeCmds[i].Err() == nil {
					t = typeCmds[i].Val()
				}
				ttlSec := int64(-1)
				if ttlCmds[i].Err() == nil {
					d := ttlCmds[i].Val()
					if d == -1*time.Second {
						ttlSec = -1
					} else if d == -2*time.Second {
						ttlSec = -2
					} else {
						ttlSec = int64(d.Seconds())
					}
				}
				var memory int64
				if memCmds[i].Err() == nil {
					memory = memCmds[i].Val()
				}
				isBig := memory > 1024*1024
				allKeys = append(allKeys, keyEntry{Key: k, Type: t, TTL: ttlSec, Memory: memory, IsBig: isBig})
			}
		}

		if truncated || nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	resp := batchPreviewResponse{
		Keys:      allKeys,
		TotalKeys: len(allKeys),
		Truncated: truncated,
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, resp, nil))
}

// --- BatchDelete -------------------------------------------------------------

type batchDeleteRequest struct {
	Keys    []string `json:"keys"`
	Confirm bool     `json:"confirm"`
}

type batchDeleteResponse struct {
	DeletedCount int64 `json:"deletedCount"`
}

// BatchDelete deletes multiple keys. Requires confirm=true in the request body.
func (h RedisHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)

	var req batchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if len(req.Keys) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("no keys specified").Build()))
		return
	}
	if !req.Confirm {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required to delete keys").Build()))
		return
	}
	// Limit batch size to prevent accidental mass deletion
	if len(req.Keys) > 1000 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("batch delete limited to 1000 keys at a time").Build()))
		return
	}

	conn, cl, err := h.openClient(connID, dbIndex)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	deleted, err := cl.Del(ctx, req.Keys...).Result()
	if err != nil {
		h.internalErr(w, r, err)
		return
	}

	resp := batchDeleteResponse{DeletedCount: deleted}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, resp, nil))
}

// --- Command -----------------------------------------------------------------

type commandRequest struct {
	Command string `json:"command"`
}

type commandResponse struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
	TimeMs int64  `json:"timeMs"`
}

// Command executes an arbitrary Redis command via the command console.
// KEYS and other forbidden commands are rejected. Write commands are blocked
// when the connection is readonly.
func (h RedisHandler) Command(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbIndexStr := router.Param(r, "dbIndex")
	dbIndex, _ := strconv.Atoi(dbIndexStr)

	var req commandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}

	parts := parseCommand(req.Command)
	if len(parts) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("empty command").Build()))
		return
	}

	verb := strings.ToUpper(parts[0])

	if forbiddenCommands[verb] {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeForbidden).
			Message(fmt.Sprintf("%s is not allowed — use SCAN instead of KEYS", verb)).
			Build()))
		return
	}

	conn, cl, err := h.openClient(connID, dbIndex)
	if err != nil {
		if err == connection.ErrNotFound {
			h.connNotFound(w, r)
			return
		}
		h.internalErr(w, r, err)
		return
	}

	if conn.Readonly && redisWriteCommands[verb] {
		if guardReadonly(conn, w, r, h.Logger) {
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	args := make([]any, len(parts))
	for i, p := range parts {
		args[i] = p
	}

	start := time.Now()
	res, err := cl.Do(ctx, args...).Result()
	elapsed := time.Since(start).Milliseconds()

	resp := commandResponse{TimeMs: elapsed}
	if err != nil && err != redis.Nil {
		resp.Error = err.Error()
	} else {
		resp.Result = formatRedisResult(res)
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, resp, nil))
}

// parseCommand splits a Redis command line respecting quoted strings.
func parseCommand(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}
	var parts []string
	var cur strings.Builder
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				cur.WriteByte(c)
			}
		} else if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
		} else if c == ' ' || c == '\t' {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
		} else {
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// formatRedisResult converts a redis Do result to a JSON-friendly value.
func formatRedisResult(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = formatRedisResult(item)
		}
		return out
	case []byte:
		return string(val)
	default:
		return val
	}
}
