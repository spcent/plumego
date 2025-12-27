package pubsub

import "errors"

var (
	ErrClosed       = errors.New("pubsub is closed")
	ErrInvalidTopic = errors.New("invalid topic")
	ErrInvalidOpts  = errors.New("invalid subscribe options")
)
