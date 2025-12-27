package webhookout

import "time"

type Target struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Secret  string            `json:"secret,omitempty"`
	Events  []string          `json:"events"`
	Enabled bool              `json:"enabled"`
	Headers map[string]string `json:"headers,omitempty"`

	TimeoutMs     int   `json:"timeout_ms,omitempty"`
	MaxRetries    int   `json:"max_retries,omitempty"`
	BackoffBaseMs int   `json:"backoff_base_ms,omitempty"`
	BackoffMaxMs  int   `json:"backoff_max_ms,omitempty"`
	RetryOn429    *bool `json:"retry_on_429,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TargetPatch struct {
	Name          *string            `json:"name,omitempty"`
	URL           *string            `json:"url,omitempty"`
	Secret        *string            `json:"secret,omitempty"`
	Events        *[]string          `json:"events,omitempty"`
	Enabled       *bool              `json:"enabled,omitempty"`
	Headers       *map[string]string `json:"headers,omitempty"`
	TimeoutMs     *int               `json:"timeout_ms,omitempty"`
	MaxRetries    *int               `json:"max_retries,omitempty"`
	BackoffBaseMs *int               `json:"backoff_base_ms,omitempty"`
	BackoffMaxMs  *int               `json:"backoff_max_ms,omitempty"`
	RetryOn429    *bool              `json:"retry_on_429,omitempty"`
}

type TargetFilter struct {
	Enabled *bool
	Event   string
}

type Event struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	OccurredAt time.Time      `json:"occurred_at"`
	Data       map[string]any `json:"data"`
	Meta       map[string]any `json:"meta,omitempty"`
}

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliveryRetry   DeliveryStatus = "retrying"
	DeliverySuccess DeliveryStatus = "success"
	DeliveryFailed  DeliveryStatus = "failed"
	DeliveryDead    DeliveryStatus = "dead"
)

type Delivery struct {
	ID        string `json:"id"`
	TargetID  string `json:"target_id"`
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`

	// IMPORTANT: replayable payload (raw JSON) - stored per delivery
	// NOTE: kept as []byte in store; handlers can choose to omit it in list output if desired.
	PayloadJSON []byte `json:"payload_json,omitempty"`

	Attempt int            `json:"attempt"`
	Status  DeliveryStatus `json:"status"`
	NextAt  *time.Time     `json:"next_at,omitempty"`

	LastHTTPStatus  int    `json:"last_http_status,omitempty"`
	LastError       string `json:"last_error,omitempty"`
	LastDurationMs  int    `json:"last_duration_ms,omitempty"`
	LastRespSnippet string `json:"last_resp_snippet,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeliveryFilter struct {
	TargetID *string
	Event    *string
	Status   *DeliveryStatus
	Limit    int
	Cursor   string
	From     *time.Time
	To       *time.Time
}
