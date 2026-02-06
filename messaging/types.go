package messaging

import (
	"errors"
	"time"
)

// Channel represents the delivery channel for a message.
type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
)

// Priority maps to mq.MessagePriority for task ordering.
type Priority int

const (
	PriorityLow      Priority = 10 // marketing, non-urgent
	PriorityNormal   Priority = 20 // notifications
	PriorityHigh     Priority = 30 // transactional
	PriorityCritical Priority = 40 // OTP, alerts
)

// SendRequest is the unified request for sending SMS or email.
type SendRequest struct {
	ID          string            `json:"id"`
	Channel     Channel           `json:"channel"`
	Priority    Priority          `json:"priority"`
	To          string            `json:"to"`
	Subject     string            `json:"subject,omitempty"`
	Template    string            `json:"template,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	Body        string            `json:"body,omitempty"`
	HTML        bool              `json:"html,omitempty"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	DedupeKey   string            `json:"dedupe_key,omitempty"`
	MaxRetries  int               `json:"max_retries,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`
	TTL         time.Duration     `json:"ttl,omitempty"`
}

// SendResult is the result of a send attempt.
type SendResult struct {
	RequestID  string    `json:"request_id"`
	Channel    Channel   `json:"channel"`
	Status     string    `json:"status"`
	ProviderID string    `json:"provider_id,omitempty"`
	Error      string    `json:"error,omitempty"`
	SentAt     time.Time `json:"sent_at,omitempty"`
	Attempts   int       `json:"attempts"`
}

// BatchRequest groups multiple send requests.
type BatchRequest struct {
	Requests []SendRequest `json:"requests"`
}

// BatchResult groups the results of a batch send.
type BatchResult struct {
	Total    int      `json:"total"`
	Accepted int      `json:"accepted"`
	Rejected int      `json:"rejected"`
	Errors   []string `json:"errors,omitempty"`
}

// ServiceStats exposes queue and delivery statistics.
type ServiceStats struct {
	Queued      int64 `json:"queued"`
	InFlight    int64 `json:"in_flight"`
	Dead        int64 `json:"dead"`
	Expired     int64 `json:"expired"`
	TotalSent   int64 `json:"total_sent"`
	TotalFailed int64 `json:"total_failed"`
}

// topicFor returns the mq topic for a channel.
func topicFor(ch Channel) string {
	return string(ch) + ".send"
}

// Sentinel errors for the messaging package.
var (
	ErrInvalidChannel   = errors.New("messaging: invalid channel")
	ErrInvalidEmail     = errors.New("messaging: invalid email address")
	ErrInvalidPhone     = errors.New("messaging: invalid phone number")
	ErrMissingBody      = errors.New("messaging: body or template is required")
	ErrMissingRecipient = errors.New("messaging: recipient (to) is required")
	ErrMissingSubject   = errors.New("messaging: subject is required for email")
	ErrProviderFailure  = errors.New("messaging: provider send failed")
	ErrTemplateRender   = errors.New("messaging: template rendering failed")
)
