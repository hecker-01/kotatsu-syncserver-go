// Package services - mail.go provides email sending capabilities.
// Supports multiple providers: console (for development) and SMTP (for production).
package services

import (
	"fmt"
	"net/smtp"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// MailService defines the interface for sending emails.
type MailService interface {
	// Send sends an email with the given parameters.
	// Returns an error if the email could not be sent.
	Send(to, subject, textBody, htmlBody string) error
}

// ConsoleMailService prints emails to the console (for development/testing).
type ConsoleMailService struct{}

// NewConsoleMailService creates a new console mail service instance.
func NewConsoleMailService() *ConsoleMailService {
	return &ConsoleMailService{}
}

// Send prints the email to the console instead of sending it.
func (s *ConsoleMailService) Send(to, subject, textBody, htmlBody string) error {
	logger.Info("=== CONSOLE MAIL SERVICE ===")
	logger.Info("To: " + to)
	logger.Info("Subject: " + subject)
	logger.Info("Text Body:")
	logger.Info(textBody)
	if htmlBody != "" {
		logger.Info("HTML Body:")
		logger.Info(htmlBody)
	}
	logger.Info("=== END MAIL ===")
	return nil
}

// SMTPMailService sends emails via SMTP.
type SMTPMailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPMailService creates a new SMTP mail service instance.
func NewSMTPMailService(host string, port int, username, password, from string) *SMTPMailService {
	return &SMTPMailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Send sends an email via SMTP.
func (s *SMTPMailService) Send(to, subject, textBody, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Build email headers and body
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n", s.from, to, subject)

	var body string
	if htmlBody != "" {
		// Multipart message with both text and HTML
		boundary := "==BOUNDARY=="
		headers += fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary)
		body = fmt.Sprintf("--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n\r\n--%s--",
			boundary, textBody, boundary, htmlBody, boundary)
	} else {
		headers += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"
		body = textBody
	}

	msg := []byte(headers + body)

	// Set up authentication
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	err := smtp.SendMail(addr, auth, s.from, []string{to}, msg)
	if err != nil {
		logger.Error("failed to send email via SMTP", "to", to, "error", err)
		return err
	}

	logger.Info("email sent via SMTP", "to", to, "subject", subject)
	return nil
}

// NewMailService creates a mail service based on configuration.
// Returns ConsoleMailService for MAIL_PROVIDER=console, SMTPMailService for MAIL_PROVIDER=smtp.
func NewMailService(cfg *utils.Config) MailService {
	switch cfg.MailProvider {
	case "smtp":
		return NewSMTPMailService(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFrom,
		)
	default:
		return NewConsoleMailService()
	}
}
