package ops

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/router"
)

const DefaultBasePath = "/ops"

// Handler exposes protected operations endpoints for queue/receipt/tenant management.
type Handler struct {
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
	QueueStats    func(ctx context.Context, queue string) (QueueStats, error)
	QueueList     func(ctx context.Context) ([]string, error)
	QueueReplay   func(ctx context.Context, req QueueReplayRequest) (QueueReplayResult, error)
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

// New constructs an ops handler.
func New(opts Options) *Handler {
	logger := opts.Logger
	if logger == nil {
		logger = log.NewNoOpLogger()
	}
	return &Handler{
		cfg:    opts,
		logger: logger,
	}
}

func (c *Handler) RegisterRoutes(r *router.Router) {
	if !c.cfg.Enabled {
		return
	}

	c.routesOnce.Do(func() {
		base := normalizeBasePath(c.cfg.BasePath)
		group := r.Group(base)
		for _, mw := range c.authMiddlewares() {
			group.Use(mw)
		}

		group.Get("", contract.AdaptCtxHandler(c.handleSummary))
		group.Get("/queue", contract.AdaptCtxHandler(c.handleQueueStats))
		group.Post("/queue/replay", contract.AdaptCtxHandler(c.handleQueueReplay))
		group.Get("/receipts", contract.AdaptCtxHandler(c.handleReceiptLookup))
		group.Get("/channels", contract.AdaptCtxHandler(c.handleChannelHealth))
		group.Get("/tenants/quota", contract.AdaptCtxHandler(c.handleTenantQuota))
	})
}

func (c *Handler) handleSummary(ctx *contract.Ctx) {
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

	_ = ctx.Response(http.StatusOK, data, nil)
}

func (c *Handler) handleQueueStats(ctx *contract.Ctx) {
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
			contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Status(http.StatusBadRequest).
				Category(contract.CategoryValidation).
				Type(contract.TypeValidation).
				Code(contract.CodeValidationError).
				Message("validation failed for field 'queue': queue parameter required").
				Detail("field", "queue").
				Detail("validation_message", "queue parameter required").
				Build())
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

	_ = ctx.Response(http.StatusOK, map[string]any{
		"queues": stats,
	}, nil)
}

func (c *Handler) handleQueueReplay(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.QueueReplay == nil {
		writeNotImplemented(ctx, "queue_replay_not_configured", "queue replay hook not configured")
		return
	}

	var req QueueReplayRequest
	if err := ctx.BindJSON(&req, nil); err != nil {
		_ = contract.WriteBindError(ctx.W, ctx.R, err)
		return
	}
	if err := contract.ValidateStruct(&req); err != nil {
		_ = contract.WriteBindError(ctx.W, ctx.R, err)
		return
	}

	result, err := c.cfg.Hooks.QueueReplay(ctx.R.Context(), req)
	if err != nil {
		c.writeHookError(ctx, "queue_replay_failed", err)
		return
	}

	_ = ctx.Response(http.StatusOK, map[string]any{
		"replay": result,
	}, nil)
}

func (c *Handler) handleReceiptLookup(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.ReceiptLookup == nil {
		writeNotImplemented(ctx, "receipt_lookup_not_configured", "receipt lookup hook not configured")
		return
	}

	messageID := strings.TrimSpace(ctx.Query.Get("message_id"))
	if messageID == "" {
		contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Type(contract.TypeValidation).
			Code(contract.CodeValidationError).
			Message("validation failed for field 'message_id': message_id parameter required").
			Detail("field", "message_id").
			Detail("validation_message", "message_id parameter required").
			Build())
		return
	}

	receipt, err := c.cfg.Hooks.ReceiptLookup(ctx.R.Context(), messageID)
	if err != nil {
		c.writeHookError(ctx, "receipt_lookup_failed", err)
		return
	}

	_ = ctx.Response(http.StatusOK, map[string]any{
		"receipt": receipt,
	}, nil)
}

func (c *Handler) handleChannelHealth(ctx *contract.Ctx) {
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
			contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Status(http.StatusBadRequest).
				Category(contract.CategoryValidation).
				Type(contract.TypeValidation).
				Code(contract.CodeValidationError).
				Message("validation failed for field 'provider': provider parameter required").
				Detail("field", "provider").
				Detail("validation_message", "provider parameter required").
				Build())
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

	_ = ctx.Response(http.StatusOK, map[string]any{
		"channels": channels,
	}, nil)
}

func (c *Handler) handleTenantQuota(ctx *contract.Ctx) {
	if ctx == nil {
		return
	}

	if c.cfg.Hooks.TenantQuota == nil {
		writeNotImplemented(ctx, "tenant_quota_not_configured", "tenant quota hook not configured")
		return
	}

	tenantID := strings.TrimSpace(ctx.Query.Get("tenant_id"))
	if tenantID == "" {
		contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Type(contract.TypeValidation).
			Code(contract.CodeValidationError).
			Message("validation failed for field 'tenant_id': tenant_id parameter required").
			Detail("field", "tenant_id").
			Detail("validation_message", "tenant_id parameter required").
			Build())
		return
	}

	snapshot, err := c.cfg.Hooks.TenantQuota(ctx.R.Context(), tenantID)
	if err != nil {
		c.writeHookError(ctx, "tenant_quota_failed", err)
		return
	}

	_ = ctx.Response(http.StatusOK, map[string]any{
		"quota": snapshot,
	}, nil)
}

func (c *Handler) authMiddlewares() []middleware.Middleware {
	var middlewares []middleware.Middleware

	if c.cfg.Auth.Middleware != nil {
		middlewares = append(middlewares, c.cfg.Auth.Middleware)
	}

	token := strings.TrimSpace(c.cfg.Auth.Token)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("AUTH_TOKEN"))
	}
	if token != "" {
		middlewares = append(middlewares, auth.SimpleAuth(token))
	}

	if !c.cfg.Auth.AllowInsecure && len(middlewares) == 0 {
		middlewares = append(middlewares, denyAllMiddleware())
		if c.logger != nil {
			c.logger.Warn("ops auth not configured; denying all requests", log.Fields{})
		}
	}

	return middlewares
}

func (c *Handler) hasAuthConfigured() bool {
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

func (c *Handler) writeHookError(ctx *contract.Ctx, code string, err error) {
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
	contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
		Status(http.StatusInternalServerError).
		Code(code).
		Message("internal error").
		Category(contract.CategoryServer).
		Build())
}

func writeNotImplemented(ctx *contract.Ctx, code, message string) {
	if ctx == nil {
		return
	}
	contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
		Status(http.StatusNotImplemented).
		Code(code).
		Message(message).
		Category(contract.CategoryServer).
		Build())
}

func denyAllMiddleware() middleware.Middleware {
	return func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contract.WriteError(w, r, contract.NewErrorBuilder().
				Status(http.StatusUnauthorized).
				Category(contract.CategoryAuth).
				Type(contract.TypeUnauthorized).
				Code(contract.CodeUnauthorized).
				Message("ops auth required").
				Build())
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
