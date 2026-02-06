package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/net/mq"
	"github.com/spcent/plumego/net/mq/store"
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

// Start launches the worker pool and registers scheduled jobs.
func (s *Service) Start(ctx context.Context) error {
	s.worker.Start(ctx)
	s.registerScheduledJobs()
	return nil
}

// Stop drains the worker pool.
func (s *Service) Stop(ctx context.Context) error {
	return s.worker.Stop(ctx)
}

// Send validates and enqueues a single message.
func (s *Service) Send(ctx context.Context, req SendRequest) error {
	if err := s.validate(req); err != nil {
		return err
	}
	return s.enqueue(ctx, req)
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
		if err := s.enqueue(ctx, req); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", req.ID, err))
			continue
		}
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

func (s *Service) handleSMS(ctx context.Context, task mq.Task) error {
	var req SendRequest
	if err := json.Unmarshal(task.Payload, &req); err != nil {
		return fmt.Errorf("messaging: unmarshal sms task: %w", err)
	}

	body, err := s.renderBody(req)
	if err != nil {
		return err
	}

	result, err := s.sms.Send(ctx, SMSMessage{
		To:   req.To,
		Body: body,
	})
	if err != nil {
		s.totalFailed.Add(1)
		return fmt.Errorf("%w: %v", ErrProviderFailure, err)
	}

	s.totalSent.Add(1)
	s.publishResult(SendResult{
		RequestID:  req.ID,
		Channel:    ChannelSMS,
		Status:     "sent",
		ProviderID: result.ProviderID,
		SentAt:     time.Now(),
		Attempts:   task.Attempts,
	})
	return nil
}

func (s *Service) handleEmail(ctx context.Context, task mq.Task) error {
	var req SendRequest
	if err := json.Unmarshal(task.Payload, &req); err != nil {
		return fmt.Errorf("messaging: unmarshal email task: %w", err)
	}

	body, err := s.renderBody(req)
	if err != nil {
		return err
	}

	result, err := s.email.Send(ctx, EmailMessage{
		To:      req.To,
		Subject: req.Subject,
		Body:    body,
		HTML:    req.HTML,
	})
	if err != nil {
		s.totalFailed.Add(1)
		return fmt.Errorf("%w: %v", ErrProviderFailure, err)
	}

	s.totalSent.Add(1)
	s.publishResult(SendResult{
		RequestID:  req.ID,
		Channel:    ChannelEmail,
		Status:     "sent",
		ProviderID: result.MessageID,
		SentAt:     time.Now(),
		Attempts:   task.Attempts,
	})
	return nil
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
