// Package audit provides write-ahead audit logging for data mutation events.
package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"cloud-vault/internal/idgen"
)

// Action constants.
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// Resource type constants.
const (
	ResourceDocument   = "document"
	ResourceTag        = "tag"
	ResourceCollection = "collection"
)

// Event is a single immutable audit log entry.
type Event struct {
	ID           string         `json:"id"`
	ActorID      string         `json:"actor_id,omitempty"`
	ActorIP      string         `json:"actor_ip,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Detail       map[string]any `json:"detail,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// Logger writes audit events to the database.
// All methods are safe for concurrent use. Log never blocks the caller —
// audit write failures are swallowed to avoid aborting the primary operation.
type Logger struct {
	db *sql.DB
}

// NewLogger creates an audit Logger backed by db.
func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// Log records an audit event. It never returns an error; failures are silently dropped.
func (l *Logger) Log(ctx context.Context, actorID, actorIP, action, resourceType, resourceID string, detail map[string]any) {
	if l == nil || l.db == nil {
		return
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339)
	var detailJSON *string
	if len(detail) > 0 {
		if b, err := json.Marshal(detail); err == nil {
			s := string(b)
			detailJSON = &s
		}
	}
	_, _ = l.db.ExecContext(ctx,
		`INSERT INTO audit_events (id, actor_id, actor_ip, action, resource_type, resource_id, detail_json, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, nullableStr(actorID), nullableStr(actorIP), action, resourceType, resourceID, detailJSON, now,
	)
}

// List returns audit events newest-first, optionally filtered by resource type/ID.
func (l *Logger) List(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]Event, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args := []any{}
	where := "1=1"
	if resourceType != "" {
		where += " AND resource_type = ?"
		args = append(args, resourceType)
	}
	if resourceID != "" {
		where += " AND resource_id = ?"
		args = append(args, resourceID)
	}

	var total int
	if err := l.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit events: %w", err)
	}

	queryArgs := append(args, limit, offset)
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, actor_id, actor_ip, action, resource_type, resource_id, detail_json, created_at
		 FROM audit_events WHERE `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var actorID, actorIP, detailJSON *string
		var createdAt string
		if err := rows.Scan(&e.ID, &actorID, &actorIP, &e.Action, &e.ResourceType, &e.ResourceID, &detailJSON, &createdAt); err != nil {
			return nil, 0, err
		}
		if actorID != nil {
			e.ActorID = *actorID
		}
		if actorIP != nil {
			e.ActorIP = *actorIP
		}
		if detailJSON != nil {
			_ = json.Unmarshal([]byte(*detailJSON), &e.Detail)
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		events = append(events, e)
	}
	return events, total, rows.Err()
}

// ClientIP extracts the originating client IP from r, preferring proxy
// forwarding headers and falling back to the raw remote address.
func ClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if idx := strings.Index(fwd, ","); idx > 0 {
			return strings.TrimSpace(fwd[:idx])
		}
		return strings.TrimSpace(fwd)
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
