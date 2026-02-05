package notification

import "context"

// EmailMessage represents an email message
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// EmailService defines the contract for sending emails
type EmailService interface {
	SendEmail(ctx context.Context, message EmailMessage) error
}
