package email

import (
	"fmt"
	"net/smtp"
	"os"
)

type Sender interface {
	Send(to, subject, body string) error
}

type NoopSender struct{}

func (NoopSender) Send(to, subject, body string) error {
	return nil
}

type SMTPSender struct {
	host string
	user string
	pass string
	from string
}

func NewFromEnv() Sender {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return NoopSender{}
	}
	return &SMTPSender{
		host: host,
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: os.Getenv("SMTP_FROM"),
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	from := s.from
	if from == "" {
		from = s.user
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	addr := s.host + ":587"
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
