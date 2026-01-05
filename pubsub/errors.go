package pubsub

import "errors"

var (
	// ErrClosed is returned when the pubsub system is already closed
	ErrClosed = errors.New("pubsub is closed")

	// ErrInvalidTopic is returned for empty or whitespace-only topics
	ErrInvalidTopic = errors.New("invalid topic")

	// ErrInvalidOpts is returned for invalid subscription options
	ErrInvalidOpts = errors.New("invalid subscribe options")

	// ErrBufferTooSmall is returned when buffer size is less than 1
	ErrBufferTooSmall = errors.New("buffer size must be at least 1")

	// ErrPublishToClosed is returned when publishing to a closed pubsub
	ErrPublishToClosed = errors.New("cannot publish to closed pubsub")

	// ErrSubscribeToClosed is returned when subscribing to a closed pubsub
	ErrSubscribeToClosed = errors.New("cannot subscribe to closed pubsub")
)

// IsClosedError checks if error is related to closed pubsub
func IsClosedError(err error) bool {
	return err == ErrClosed || err == ErrPublishToClosed || err == ErrSubscribeToClosed
}
