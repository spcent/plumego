package message

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusAccepted  Status = "accepted"
	StatusQueued    Status = "queued"
	StatusSending   Status = "sending"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
	StatusExpired   Status = "expired"
	StatusCanceled  Status = "canceled"
)

var ErrInvalidTransition = errors.New("message: invalid status transition")

type Message struct {
	ID             string
	TenantID       string
	Status         Status
	ReasonCode     ReasonCode
	ReasonDetail   string
	Attempts       int
	MaxAttempts    int
	NextAttemptAt  time.Time
	ProviderMsgID  string
	IdempotencyKey string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TransitionHook interface {
	Before(ctx context.Context, from Status, to Status, msg Message) error
	After(ctx context.Context, from Status, to Status, msg Message) error
	OnInvalid(ctx context.Context, from Status, to Status, msg Message, err error)
}

var allowedTransitions = map[Status]map[Status]struct{}{
	StatusAccepted: {StatusQueued: {}},
	StatusQueued: {
		StatusSending:  {},
		StatusFailed:   {},
		StatusExpired:  {},
		StatusCanceled: {},
	},
	StatusSending: {
		StatusSent:     {},
		StatusFailed:   {},
		StatusExpired:  {},
		StatusQueued:   {},
		StatusCanceled: {},
	},
	StatusSent: {
		StatusDelivered: {},
		StatusFailed:    {},
		StatusExpired:   {},
	},
	StatusFailed: {
		StatusQueued: {},
	},
}

func CanTransition(from Status, to Status) bool {
	if next, ok := allowedTransitions[from]; ok {
		_, allowed := next[to]
		return allowed
	}
	return false
}

func ApplyTransition(ctx context.Context, msg *Message, to Status, hooks ...TransitionHook) error {
	if msg == nil {
		return ErrInvalidTransition
	}
	from := msg.Status
	if !CanTransition(from, to) {
		err := ErrInvalidTransition
		for _, hook := range hooks {
			hook.OnInvalid(ctx, from, to, *msg, err)
		}
		return err
	}

	for _, hook := range hooks {
		if err := hook.Before(ctx, from, to, *msg); err != nil {
			return err
		}
	}

	msg.Status = to
	msg.UpdatedAt = time.Now()
	msg.Version++

	for _, hook := range hooks {
		if err := hook.After(ctx, from, to, *msg); err != nil {
			return err
		}
	}

	return nil
}
