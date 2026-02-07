package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/net/mq"
	"github.com/spcent/plumego/net/mq/store"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/scheduler"
	"github.com/spcent/plumego/security/input"
)

// Service is the central message-sending coordinator.
// It validates requests, enqueues tasks into mq.TaskQueue, and starts
// workers that consume tasks and dispatch them via providers.
type Service struct {
	queue     *mq.TaskQueue
	worker    *mq.Worker
	store     mq.TaskStore
	scheduler *scheduler.Scheduler
	bus       *pubsub.InProcPubSub
	templates *TemplateEngine
	logger    log.StructuredLogger

	sms   SMSProvider
	email EmailProvider

	receipts ReceiptStore
	quota    *QuotaChecker
	monitor  *ChannelMonitor
	metrics  *metricsWrapper
	webhook  *WebhookNotifier

	totalSent   atomic.Int64
	totalFailed atomic.Int64
}

// Config holds all dependencies required by the Service.
type Config struct {
	// TaskStore is the backing store for the task queue.
	// Use store.NewMemory for dev or store.NewSQL for prod.
	TaskStore mq.TaskStore

	// Scheduler is used for periodic jobs (DLQ retry, cleanup, etc.).
	Scheduler *scheduler.Scheduler

	// Bus is the in-process pub/sub for result events.
	Bus *pubsub.InProcPubSub

	// SMS provider; nil disables SMS channel.
	SMS SMSProvider

	// Email provider; nil disables email channel.
	Email EmailProvider

	// Logger; optional.
	Logger log.StructuredLogger

	// ReceiptStore for persisting delivery receipts; nil uses in-memory default.
	Receipts ReceiptStore

	// QuotaChecker for per-tenant send limits; nil disables quota enforcement.
	Quota *QuotaChecker

	// MetricsCollector for observability; nil disables metrics.
	MetricsCollector metrics.MetricsCollector

	// Webhook is the outbound webhook service for delivery notifications.
	// When set together with Bus, a WebhookNotifier is created automatically.
	Webhook *webhookout.Service

	// Worker tuning.
	WorkerConcurrency int
	WorkerMaxInflight int
	ConsumerID        string
}

// New creates a Service wired to the given dependencies.
func New(cfg Config) *Service {
	if cfg.TaskStore == nil {
		cfg.TaskStore = store.NewMemory(store.DefaultMemConfig())
	}
	if cfg.WorkerConcurrency <= 0 {
		cfg.WorkerConcurrency = 4
	}
	if cfg.WorkerMaxInflight <= 0 {
		cfg.WorkerMaxInflight = cfg.WorkerConcurrency * 2
	}
	if cfg.Receipts == nil {
		cfg.Receipts = NewMemReceiptStore(0)
	}

	queue := mq.NewTaskQueue(cfg.TaskStore)

	workerCfg := mq.WorkerConfig{
		ConsumerID:          cfg.ConsumerID,
		Concurrency:         cfg.WorkerConcurrency,
		MaxInflight:         cfg.WorkerMaxInflight,
		LeaseDuration:       30 * time.Second,
		LeaseExtendInterval: 15 * time.Second,
		PollInterval:        200 * time.Millisecond,
		ShutdownTimeout:     30 * time.Second,
		RetryPolicy: mq.ExponentialBackoff{
			Base:   2 * time.Second,
			Max:    5 * time.Minute,
			Factor: 2,
			Jitter: 0.2,
		},
	}
	worker := mq.NewWorker(queue, workerCfg)

	monitor := NewChannelMonitor(cfg.SMS, cfg.Email, cfg.Logger)

	var wh *WebhookNotifier
	if cfg.Bus != nil && cfg.Webhook != nil {
		wh = NewWebhookNotifier(cfg.Bus, cfg.Webhook, cfg.Logger)
	}

	svc := &Service{
		queue:     queue,
		worker:    worker,
		store:     cfg.TaskStore,
		scheduler: cfg.Scheduler,
		bus:       cfg.Bus,
		templates: NewTemplateEngine(),
		logger:    cfg.Logger,
		sms:       cfg.SMS,
		email:     cfg.Email,
		receipts:  cfg.Receipts,
		quota:     cfg.Quota,
		monitor:   monitor,
		metrics:   newMetrics(cfg.MetricsCollector),
		webhook:   wh,
	}

	if cfg.SMS != nil {
		worker.Register(topicFor(ChannelSMS), svc.handleSMS)
	}
	if cfg.Email != nil {
		worker.Register(topicFor(ChannelEmail), svc.handleEmail)
	}

	return svc
}

// Templates returns the template engine for registration.
func (s *Service) Templates() *TemplateEngine { return s.templates }

// Receipts returns the receipt store for direct queries.
func (s *Service) Receipts() ReceiptStore { return s.receipts }

// Monitor returns the channel health monitor.
func (s *Service) Monitor() *ChannelMonitor { return s.monitor }

// Start launches the worker pool, registers scheduled jobs,
// and starts the webhook notifier (if configured).
func (s *Service) Start(ctx context.Context) error {
	s.worker.Start(ctx)
	s.registerScheduledJobs()
	if s.monitor != nil {
		s.monitor.RegisterJobs(s.scheduler)
	}
	if s.webhook != nil {
		if err := s.webhook.Start(ctx); err != nil && s.logger != nil {
			s.logger.Warn("webhook notifier start failed", log.Fields{"error": err.Error()})
		}
	}
	return nil
}

// Stop drains the worker pool and stops the webhook notifier.
func (s *Service) Stop(ctx context.Context) error {
	if s.webhook != nil {
		s.webhook.Stop()
	}
	return s.worker.Stop(ctx)
}

// Send validates, checks quota, and enqueues a single message.
func (s *Service) Send(ctx context.Context, req SendRequest) error {
	if err := s.validate(req); err != nil {
		if s.metrics.enabled() {
			s.metrics.ObserveValidation(ctx, req.Channel)
		}
		return err
	}
	if err := s.checkQuota(ctx, req); err != nil {
		return err
	}
	start := time.Now()
	err := s.enqueue(ctx, req)
	if s.metrics.enabled() {
		s.metrics.ObserveEnqueue(ctx, req.Channel, time.Since(start), err)
	}
	if err == nil {
		s.saveReceipt(req, "queued", "", "")
	}
	return err
}

// SendBatch validates and enqueues multiple messages, returning a summary.
func (s *Service) SendBatch(ctx context.Context, batch BatchRequest) BatchResult {
	result := BatchResult{Total: len(batch.Requests)}
	for _, req := range batch.Requests {
		if err := s.validate(req); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", req.ID, err))
			continue
		}
		if err := s.checkQuota(ctx, req); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", req.ID, err))
			continue
		}
		if err := s.enqueue(ctx, req); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", req.ID, err))
			continue
		}
		s.saveReceipt(req, "queued", "", "")
		result.Accepted++
	}
	return result
}

// Stats returns current queue statistics.
func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	qs, err := s.queue.Stats(ctx)
	if err != nil {
		return ServiceStats{}, err
	}
	return ServiceStats{
		Queued:      qs.Queued,
		InFlight:    qs.Leased,
		Dead:        qs.Dead,
		Expired:     qs.Expired,
		TotalSent:   s.totalSent.Load(),
		TotalFailed: s.totalFailed.Load(),
	}, nil
}

// --- internal ---

func (s *Service) checkQuota(ctx context.Context, req SendRequest) error {
	if s.quota == nil {
		return nil
	}
	return s.quota.Allow(ctx, req.TenantID)
}

func (s *Service) validate(req SendRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return ErrMissingRecipient
	}
	switch req.Channel {
	case ChannelSMS:
		if s.sms == nil {
			return fmt.Errorf("%w: sms provider not configured", ErrInvalidChannel)
		}
		if !input.ValidatePhone(req.To) {
			return ErrInvalidPhone
		}
		if req.Body == "" && req.Template == "" {
			return ErrMissingBody
		}
	case ChannelEmail:
		if s.email == nil {
			return fmt.Errorf("%w: email provider not configured", ErrInvalidChannel)
		}
		if !input.ValidateEmail(req.To) {
			return ErrInvalidEmail
		}
		if req.Subject == "" {
			return ErrMissingSubject
		}
		if req.Body == "" && req.Template == "" {
			return ErrMissingBody
		}
	default:
		return fmt.Errorf("%w: %q", ErrInvalidChannel, req.Channel)
	}
	return nil
}

func (s *Service) enqueue(ctx context.Context, req SendRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("messaging: marshal request: %w", err)
	}

	pri := mq.MessagePriority(req.Priority)
	opts := mq.EnqueueOptions{
		DedupeKey:   req.DedupeKey,
		Priority:    &pri,
		MaxAttempts: req.MaxRetries,
	}
	if req.ScheduledAt != nil {
		opts.AvailableAt = *req.ScheduledAt
	}
	if req.TTL > 0 {
		opts.ExpiresAt = time.Now().Add(req.TTL)
	}

	task := mq.Task{
		ID:       req.ID,
		Topic:    topicFor(req.Channel),
		TenantID: req.TenantID,
		Payload:  payload,
	}
	return s.queue.Enqueue(ctx, task, opts)
}

func (s *Service) renderBody(req SendRequest) (string, error) {
	if req.Template != "" && s.templates.Has(req.Template) {
		return s.templates.Render(req.Template, req.Params)
	}
	if req.Body != "" {
		return req.Body, nil
	}
	return "", ErrMissingBody
}

// deliverTask is the unified handler for both SMS and email task processing.
// It unmarshals the request, renders the body, calls the channel-specific
// sendFn, and records metrics, receipts, and results.
func (s *Service) deliverTask(
	ctx context.Context,
	task mq.Task,
	channel Channel,
	providerName string,
	sendFn func(ctx context.Context, req SendRequest, body string) (providerID string, err error),
) error {
	var req SendRequest
	if err := json.Unmarshal(task.Payload, &req); err != nil {
		return fmt.Errorf("messaging: unmarshal %s task: %w", channel, err)
	}

	body, err := s.renderBody(req)
	if err != nil {
		return err
	}

	start := time.Now()
	providerID, err := sendFn(ctx, req, body)
	elapsed := time.Since(start)

	if err != nil {
		s.totalFailed.Add(1)
		s.monitor.RecordFailure(channel, err)
		s.metrics.ObserveSend(ctx, channel, providerName, elapsed, err)
		s.updateReceipt(req.ID, "failed", "", providerName, err.Error(), task.Attempts)
		return fmt.Errorf("%w: %v", ErrProviderFailure, err)
	}

	s.totalSent.Add(1)
	s.monitor.RecordSuccess(channel, elapsed)
	s.metrics.ObserveSend(ctx, channel, providerName, elapsed, nil)
	s.updateReceipt(req.ID, "sent", providerID, providerName, "", task.Attempts)
	s.publishResult(SendResult{
		RequestID:  req.ID,
		Channel:    channel,
		Status:     "sent",
		ProviderID: providerID,
		SentAt:     time.Now(),
		Attempts:   task.Attempts,
	})
	return nil
}

func (s *Service) handleSMS(ctx context.Context, task mq.Task) error {
	return s.deliverTask(ctx, task, ChannelSMS, s.sms.Name(),
		func(ctx context.Context, req SendRequest, body string) (string, error) {
			result, err := s.sms.Send(ctx, SMSMessage{To: req.To, Body: body})
			if err != nil {
				return "", err
			}
			return result.ProviderID, nil
		})
}

func (s *Service) handleEmail(ctx context.Context, task mq.Task) error {
	return s.deliverTask(ctx, task, ChannelEmail, s.email.Name(),
		func(ctx context.Context, req SendRequest, body string) (string, error) {
			result, err := s.email.Send(ctx, EmailMessage{
				To: req.To, Subject: req.Subject, Body: body, HTML: req.HTML,
			})
			if err != nil {
				return "", err
			}
			return result.MessageID, nil
		})
}

// providerNameFor returns the provider name for a channel.
func (s *Service) providerNameFor(ch Channel) string {
	switch ch {
	case ChannelSMS:
		if s.sms != nil {
			return s.sms.Name()
		}
	case ChannelEmail:
		if s.email != nil {
			return s.email.Name()
		}
	}
	return ""
}

func (s *Service) saveReceipt(req SendRequest, status, providerID, errMsg string) {
	if s.receipts == nil {
		return
	}
	_ = s.receipts.Save(Receipt{
		ID:         req.ID,
		Channel:    req.Channel,
		To:         req.To,
		Status:     status,
		ProviderID: providerID,
		Provider:   s.providerNameFor(req.Channel),
		Error:      errMsg,
		TenantID:   req.TenantID,
		QueuedAt:   time.Now(),
	})
}

func (s *Service) updateReceipt(id, status, providerID, provider, errMsg string, attempts int) {
	if s.receipts == nil {
		return
	}
	r, ok := s.receipts.Get(id)
	if !ok {
		return
	}
	r.Status = status
	r.Attempts = attempts
	r.Provider = provider
	if providerID != "" {
		r.ProviderID = providerID
	}
	if errMsg != "" {
		r.Error = errMsg
	}
	if status == "sent" {
		r.SentAt = time.Now()
	}
	_ = s.receipts.Save(r)
}

func (s *Service) publishResult(result SendResult) {
	if s.bus == nil {
		return
	}
	data, _ := json.Marshal(result)
	_ = s.bus.Publish("messaging.result", pubsub.Message{
		Topic: "messaging.result",
		Type:  string(result.Channel),
		Time:  time.Now(),
		Data:  string(data),
	})
}
