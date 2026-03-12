package mq

import "errors"

var (
	ErrDuplicateTask = errors.New("mq: duplicate task")
	ErrTaskNotFound  = errors.New("mq: task not found")
	ErrLeaseLost     = errors.New("mq: lease lost")
	ErrTaskExpired   = errors.New("mq: task expired")
)
