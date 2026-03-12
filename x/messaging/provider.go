package messaging

import "context"

// SMSProvider sends SMS messages via an external gateway.
type SMSProvider interface {
	Send(ctx context.Context, msg SMSMessage) (*SMSResult, error)
	Name() string
}

// SMSMessage is the payload for a single SMS send.
type SMSMessage struct {
	To   string
	Body string
}

// SMSResult is returned by the provider after a successful send.
type SMSResult struct {
	ProviderID string
}

// EmailProvider sends email messages via SMTP or an HTTP API.
type EmailProvider interface {
	Send(ctx context.Context, msg EmailMessage) (*EmailResult, error)
	Name() string
}

// EmailMessage is the payload for a single email send.
type EmailMessage struct {
	To      string
	Subject string
	Body    string
	HTML    bool
}

// EmailResult is returned by the provider after a successful send.
type EmailResult struct {
	MessageID string
}
