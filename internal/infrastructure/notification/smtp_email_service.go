package notification

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"panda-pocket/internal/domain/notification"
)

// SMTPEmailService implements EmailService using SMTP
type SMTPEmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
	fromName string
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService() *SMTPEmailService {
	return &SMTPEmailService{
		host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		port:     getEnv("SMTP_PORT", "587"),
		username: getEnv("SMTP_USERNAME", ""),
		password: getEnv("SMTP_PASSWORD", ""),
		from:     getEnv("SMTP_FROM_EMAIL", ""),
		fromName: getEnv("SMTP_FROM_NAME", "PandaPocket"),
	}
}

// SendEmail sends an email using SMTP
func (s *SMTPEmailService) SendEmail(ctx context.Context, message notification.EmailMessage) error {
	// If credentials are not set (e.g. dev mode without real creds), log and skip
	if s.username == "" || s.password == "" || s.username == "your_email@gmail.com" {
		fmt.Printf("[MOCK EMAIL] To: %s, Subject: %s\n%s\n", message.To, message.Subject, message.Body)
		return nil
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.fromName, s.from)
	headers["To"] = message.To
	headers["Subject"] = message.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	headerString := ""
	for k, v := range headers {
		headerString += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	msg := []byte(headerString + "\r\n" + message.Body)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	if err := smtp.SendMail(addr, auth, s.from, []string{message.To}, msg); err != nil {
		return err
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
