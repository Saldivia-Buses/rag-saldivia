package service

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

// SMTPMailer sends emails via SMTP.
type SMTPMailer struct {
	host string
	port string
	from string
}

// NewSMTPMailer creates an SMTP mailer.
func NewSMTPMailer(host, port, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, from: from}
}

// Send sends a plain text email.
func (m *SMTPMailer) Send(_ context.Context, to, subject, body string) error {
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid recipient address %q: %w", to, err)
	}

	addr := net.JoinHostPort(m.host, m.port)
	sanitizedSubject := mime.QEncoding.Encode("utf-8", sanitizeHeader(subject))
	sanitizedBody := strings.ReplaceAll(body, "\r\n.\r\n", "\r\n..\r\n")

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		m.from, to, sanitizedSubject, sanitizedBody)

	if err := smtp.SendMail(addr, nil, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send email to %s: %w", to, err)
	}
	return nil
}

// sanitizeHeader strips CR/LF to prevent SMTP header injection.
func sanitizeHeader(s string) string {
	r := strings.NewReplacer("\r", "", "\n", "")
	return r.Replace(s)
}
