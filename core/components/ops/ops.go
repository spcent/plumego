package ops

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core/internal/contractio"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/router"
)

const DefaultBasePath = "/ops"

// Component exposes protected operations endpoints for queue/receipt/tenant management.
type Component struct {
	cfg        Options
	logger     log.StructuredLogger
	routesOnce sync.Once
}

// Options configures the ops component.
type Options struct {
	Enabled  bool
	BasePath string
	Auth     AuthConfig
	Hooks    Hooks
	Logger   log.StructuredLogger
}

// AuthConfig configures authentication for ops endpoints.
// When AllowInsecure is false (default), missing auth configuration will deny all requests.
type AuthConfig struct {
	Token         string
	Middleware    middleware.Middleware
	AllowInsecure bool
}

// Hooks define optional operations hooks.
type Hooks struct {
	QueueStats   func(ctx context.Context, queue string) (QueueStats, error)
	QueueList    func(ctx context.Context) ([]string, error)
	QueueReplay  func(ctx context.Context, req QueueReplayRequest) (QueueReplayResult, error)
	ReceiptLookup func(ctx context.Context, messageID string) (ReceiptRecord, error)
	ChannelHealth func(ctx context.Context, provider string) (ChannelHealth, error)
	ChannelList   func(ctx context.Context) ([]string, error)
	TenantQuota   func(ctx context.Context, tenantID string) (TenantQuotaSnapshot, error)
}

// QueueStats reports current queue counters.
type QueueStats struct {
	Queue     string         `json:"queue"`
	Queued    int64          `json:"queued"`
	Leased    int64          `json:"leased"`
	Dead      int64          `json:"dead"`
	Expired   int64          `json:"expired"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// QueueReplayRequest triggers DLQ/task replay.
type QueueReplayRequest struct {
	Queue  string `json:"queue" validate:"required"`
	Max    int    `json:"max,omitempty" validate:"min=1"`
	Reason string `json:"reason,omitempty"`
}

// QueueReplayResult reports replay outcome.
type QueueReplayResult struct {
	Queue     string `json:"queue"`
	Requested int    `json:"requested,omitempty"`
	Replayed  int    `json:"replayed"`
	Remaining int    `json:"remaining,omitempty"`
}

// ReceiptRecord describes a delivery receipt record.
type ReceiptRecord struct {
	MessageID   string         `json:"message_id"`
	Status      string         `json:"status"`
	Provider    string         `json:"provider,omitempty"`
	DeliveredAt time.Time      `json:"delivered_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
}

// ChannelHealth reports provider status.
type ChannelHealth struct {
	Provider  string         `json:"provider"`
	Status    string         `json:"status"`
	Message   string         `json:"message,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// TenantQuotaSnapshot reports tenant quota usage and limits.
type TenantQuotaSnapshot struct {
	TenantID  string         `json:"tenant_id"`
	Limits    []QuotaLimit   `json:"limits,omitempty"`
	Usage     []QuotaUsage   `json:"usage,omitempty"`
	Frozen    bool           `json:"frozen,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// QuotaLimit is a per-window limit configuration.
type QuotaLimit struct {
	Window   string `json:"window"`
	Requests int    `json:"requests,omitempty"`
	Tokens   int    `json:"tokens,omitempty"`
}

// QuotaUsage reports usage in a quota window.
type QuotaUsage struct {
	Window      string    `json:"window"`
	WindowStart time.Time `json:"window_start,omitempty"`
	WindowEnd   time.Time `json:"window_end,omitempty"`
	Requests    int       `json:"requests,omitempty"`
	Tokens      int       `json:"tokens,omitempty"`
}

// NewComponent constructs an ops component.
func NewComponent(opts Options) *Component {
	logger := opts.Logger
	if logger == nil {
		logger = log.NewGLogger()
	}
	return &Component{
		cfg:    opts,
		logger: logger,
	}
}

func (c *Component) RegisterRoutes(r *router.Router) {
	if !c.cfg.Enabled {
		return
	}

	c.routesOnce.Do(func() {
		base := normalizeBasePath(c.cfg.BasePath)
		group := r.Group(base)
		for _, mw := range c.authMiddlewares() {
			group.Use(mw)
		}

		group.GetCtx("", c.handleSummary)
		group.GetCtx("/queue", c.handleQueueStats)
		group.PostCtx("/queue/replay", c.handleQueueReplay)
		group.GetCtx("/receipts", c.handleReceiptLookup)
		group.GetCtx("/channels", c.handleChannelHealth)
		group.GetCtx("/tenants/quota", c.handleTenantQuota)
	})
}

func (c *Component) RegisterMiddleware(_ *middleware.Registry) {}

func (c *Component) Start(_ context.Context) error { return nil }

func (c *Component) Stop(_ context.Context) error { return nil }

func (c *Component) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{
		Status:  health.StatusHealthy,
		Details: map[string]any{"enabled": c.cfg.Enabled},
	}

	if !c.cfg.Enabled {
		status.Status = health.StatusDegraded
		status.Message = "component disabled"
		return "ops", status
	}

	if !c.cfg.Auth.AllowInsecure && !c.hasAuthConfigured() {
		status.Status = health.StatusDegraded
		status.Message = "ops auth not configured"
	}

	return "ops", status
}

func (c *Component) Dependencies() []reflect.Type { return nil }

func (c *Component) handleSummary(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	data := map[string]any{
		"base_path": normalizeBasePath(c.cfg.BasePath),
		"auth": map[string]any{
			"required": !c.cfg.Auth.AllowInsecure,
			"enabled":  c.hasAuthConfigured(),
		},
		"features": map[string]any{
			"queue_stats":    c.cfg.Hooks.QueueStats != nil,
			"queue_replay":   c.cfg.Hooks.QueueReplay != nil,
			"receipt_lookup": c.cfg.Hooks.ReceiptLookup != nil,
			"channel_health": c.cfg.Hooks.ChannelHealth != nil,
			"tenant_quota":   c.cfg.Hooks.TenantQuota != nil,
		},
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, data)
}

func (c *Component) handleQueueStats(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.QueueStats == nil {
		writeNotImplemented(ctx, "queue_stats_not_configured", "queue stats hook not configured")
		return
	}

	queue := strings.TrimSpace(ctx.Query.Get("queue"))
	var stats []QueueStats

	if queue == "" {
		if c.cfg.Hooks.QueueList == nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("queue", "queue parameter required"))
			return
		}
		queues, err := c.cfg.Hooks.QueueList(ctx.R.Context())
		if err != nil {
			c.writeHookError(ctx, "queue_list_failed", err)
			return
		}
		stats = make([]QueueStats, 0, len(queues))
		for _, q := range queues {
			snapshot, err := c.cfg.Hooks.QueueStats(ctx.R.Context(), q)
			if err != nil {
				c.writeHookError(ctx, "queue_stats_failed", err)
				return
			}
			stats = append(stats, snapshot)
		}
	} else {
		snapshot, err := c.cfg.Hooks.QueueStats(ctx.R.Context(), queue)
		if err != nil {
			c.writeHookError(ctx, "queue_stats_failed", err)
			return
		}
		stats = []QueueStats{snapshot}
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"queues": stats,
	})
}

func (c *Component) handleQueueReplay(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.QueueReplay == nil {
		writeNotImplemented(ctx, "queue_replay_not_configured", "queue replay hook not configured")
		return
	}

	var req QueueReplayRequest
	if err := ctx.BindAndValidateJSON(&req); err != nil {
		contract.WriteBindError(ctx.W, ctx.R, err)
		return
	}

	result, err := c.cfg.Hooks.QueueReplay(ctx.R.Context(), req)
	if err != nil {
		c.writeHookError(ctx, "queue_replay_failed", err)
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"replay": result,
	})
}

func (c *Component) handleReceiptLookup(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.ReceiptLookup == nil {
		writeNotImplemented(ctx, "receipt_lookup_not_configured", "receipt lookup hook not configured")
		return
	}

	messageID := strings.TrimSpace(ctx.Query.Get("message_id"))
	if messageID == "" {
		contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("message_id", "message_id parameter required"))
		return
	}

	receipt, err := c.cfg.Hooks.ReceiptLookup(ctx.R.Context(), messageID)
	if err != nil {
		c.writeHookError(ctx, "receipt_lookup_failed", err)
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"receipt": receipt,
	})
}

func (c *Component) handleChannelHealth(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.ChannelHealth == nil {
		writeNotImplemented(ctx, "channel_health_not_configured", "channel health hook not configured")
		return
	}

	provider := strings.TrimSpace(ctx.Query.Get("provider"))
	var channels []ChannelHealth

	if provider == "" {
		if c.cfg.Hooks.ChannelList == nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("provider", "provider parameter required"))
			return
		}
		list, err := c.cfg.Hooks.ChannelList(ctx.R.Context())
		if err != nil {
			c.writeHookError(ctx, "channel_list_failed", err)
			return
		}
		channels = make([]ChannelHealth, 0, len(list))
		for _, name := range list {
			status, err := c.cfg.Hooks.ChannelHealth(ctx.R.Context(), name)
			if err != nil {
				c.writeHookError(ctx, "channel_health_failed", err)
				return
			}
			channels = append(channels, status)
		}
	} else {
		status, err := c.cfg.Hooks.ChannelHealth(ctx.R.Context(), provider)
		if err != nil {
			c.writeHookError(ctx, "channel_health_failed", err)
			return
		}
		channels = []ChannelHealth{status}
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"channels": channels,
	})
}

func (c *Component) handleTenantQuota(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.TenantQuota == nil {
		writeNotImplemented(ctx, "tenant_quota_not_configured", "tenant quota hook not configured")
		return
	}

	tenantID := strings.TrimSpace(ctx.Query.Get("tenant_id"))
	if tenantID == "" {
		contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("tenant_id", "tenant_id parameter required"))
		return
	}

	snapshot, err := c.cfg.Hooks.TenantQuota(ctx.R.Context(), tenantID)
	if err != nil {
		c.writeHookError(ctx, "tenant_quota_failed", err)
		return
	}

	contractio.WriteContractResponse(ctx, http.StatusOK, map[string]any{
		"quota": snapshot,
	})
}

func (c *Component) authMiddlewares() []middleware.Middleware {
	var middlewares []middleware.Middleware

	if c.cfg.Auth.Middleware != nil {
		middlewares = append(middlewares, c.cfg.Auth.Middleware)
	}

	token := strings.TrimSpace(c.cfg.Auth.Token)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("AUTH_TOKEN"))
	}
	if token != "" {
		middlewares = append(middlewares, auth.FromAuthMiddleware(auth.NewSimpleAuthMiddleware(token)))
	}

	if !c.cfg.Auth.AllowInsecure && len(middlewares) == 0 {
		middlewares = append(middlewares, denyAllMiddleware())
		if c.logger != nil {
			c.logger.Warn("ops auth not configured; denying all requests", log.Fields{})
		}
	}

	return middlewares
}

func (c *Component) hasAuthConfigured() bool {
	if c.cfg.Auth.Middleware != nil {
		return true
	}
	if strings.TrimSpace(c.cfg.Auth.Token) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("AUTH_TOKEN")) != "" {
		return true
	}
	return false
}

func (c *Component) writeHookError(ctx *contract.Ctx, code string, err error) {
	if ctx == nil {
		return
	}
	if c.logger != nil && err != nil {
		c.logger.ErrorCtx(ctx.R.Context(), "ops hook failed", log.Fields{
			"code":  code,
			"error": err.Error(),
			"path":  ctx.R.URL.Path,
		})
	}
	contract.WriteError(ctx.W, ctx.R, contract.APIError{
		Status:   http.StatusInternalServerError,
		Code:     code,
		Message:  "internal error",
		Category: contract.CategoryServer,
	})
}

func writeNotImplemented(ctx *contract.Ctx, code, message string) {
	if ctx == nil {
		return
	}
	contract.WriteError(ctx.W, ctx.R, contract.APIError{
		Status:   http.StatusNotImplemented,
		Code:     code,
		Message:  message,
		Category: contract.CategoryServer,
	})
}

func denyAllMiddleware() middleware.Middleware {
	return func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contract.WriteError(w, r, contract.NewUnauthorizedError("ops auth required"))
		})
	}
}

func normalizeBasePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return DefaultBasePath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(path, "/")
}
