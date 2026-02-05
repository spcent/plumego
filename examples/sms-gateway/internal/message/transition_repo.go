package message

import "context"

// ApplyAndPersist applies a transition and persists the status update.
func ApplyAndPersist(ctx context.Context, repo Repository, msg *Message, to Status, reason Reason, hooks ...TransitionHook) error {
	if msg == nil {
		return ErrInvalidTransition
	}
	from := msg.Status
	msg.ReasonCode = reason.Code
	msg.ReasonDetail = reason.Detail
	if err := ApplyTransition(ctx, msg, to, hooks...); err != nil {
		return err
	}
	if repo == nil {
		return nil
	}
	return repo.UpdateStatus(ctx, msg.ID, from, to, reason)
}
