package pipeline

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/examples/sms-gateway/internal/message"
	"github.com/spcent/plumego/examples/sms-gateway/internal/routing"
	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
	"github.com/spcent/plumego/net/mq"
)

// Processor wires queue tasks to provider sending + state transitions.
type Processor struct {
	Repo      message.Store
	Router    *routing.PolicyRouter
	Providers map[string]ProviderSender
	Hooks     []message.TransitionHook
	OnFailure func(ctx context.Context, msg message.Message, task mq.Task, err error)
	OnDLQ     func(ctx context.Context, msg message.Message, task mq.Task, err error)
}

// Handle processes a queued send task.
func (p *Processor) Handle(ctx context.Context, task mq.Task) error {
	ctx = tasks.ContextWithTrace(ctx, task)
	payload, err := tasks.DecodeSendTask(task.Payload)
	if err != nil {
		return err
	}
	if payload.MessageID == "" {
		return fmt.Errorf("missing message id")
	}

	msg, found, err := p.Repo.Get(ctx, payload.MessageID)
	if err != nil {
		return err
	}
	if !found {
		return message.ErrMessageNotFound
	}

	if recorder, ok := p.Repo.(message.AttemptRecorder); ok {
		_ = recorder.RecordAttempt(ctx, msg.ID, task.Attempts)
	}
	msg.Attempts = task.Attempts

	if payload.Provider == "" && p.Router != nil {
		selected, _, err := p.Router.SelectProvider(ctx, msg.TenantID)
		if err != nil {
			return err
		}
		payload.Provider = selected
		if updater, ok := p.Repo.(message.ProviderUpdater); ok {
			_ = updater.UpdateProvider(ctx, msg.ID, selected)
		}
	}

	if payload.Provider == "" {
		return ErrProviderUnavailable
	}
	sender, ok := p.Providers[payload.Provider]
	if !ok {
		return fmt.Errorf("unknown provider %s", payload.Provider)
	}

	if err := p.ensureQueued(ctx, &msg); err != nil {
		return err
	}
	if msg.Status == message.StatusSent || msg.Status == message.StatusDelivered {
		return nil
	}

	if err := message.ApplyAndPersist(ctx, p.Repo, &msg, message.StatusSending, message.Reason{}, p.Hooks...); err != nil {
		return err
	}

	if err := sender.Send(ctx, payload); err != nil {
		reason := message.Reason{
			Code:      message.ReasonProviderTransient,
			Detail:    err.Error(),
			Retryable: true,
		}
		_ = message.ApplyAndPersist(ctx, p.Repo, &msg, message.StatusFailed, reason, p.Hooks...)
		if p.OnFailure != nil {
			p.OnFailure(ctx, msg, task, err)
		}
		if task.MaxAttempts > 0 && task.Attempts >= task.MaxAttempts {
			dlqReason := message.Reason{
				Code:   message.ReasonDLQ,
				Detail: err.Error(),
			}
			if recorder, ok := p.Repo.(message.DLQRecorder); ok {
				_ = recorder.RecordDLQ(ctx, msg.ID, dlqReason)
			}
			if p.OnDLQ != nil {
				p.OnDLQ(ctx, msg, task, err)
			}
		}
		return err
	}

	return message.ApplyAndPersist(ctx, p.Repo, &msg, message.StatusSent, message.Reason{}, p.Hooks...)
}

func (p *Processor) ensureQueued(ctx context.Context, msg *message.Message) error {
	switch msg.Status {
	case message.StatusAccepted:
		return message.ApplyAndPersist(ctx, p.Repo, msg, message.StatusQueued, message.Reason{}, p.Hooks...)
	case message.StatusFailed:
		return message.ApplyAndPersist(ctx, p.Repo, msg, message.StatusQueued, message.Reason{}, p.Hooks...)
	case message.StatusQueued, message.StatusSending, message.StatusSent, message.StatusDelivered:
		return nil
	default:
		return fmt.Errorf("unsupported status %s", msg.Status)
	}
}
