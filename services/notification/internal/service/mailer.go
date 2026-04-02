package service

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
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
	addr := net.JoinHostPort(m.host, m.port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		m.from, to, subject, body)

	if err := smtp.SendMail(addr, nil, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send email to %s: %w", to, err)
	}
	return nil
}
