package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

// Sender defines minimal interface for sending emails
type Sender interface {
	Send(to, subject, bodyHTML, bodyText string) error
}

// SMTPSender implements Sender using net/smtp
type SMTPSender struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPSenderFromEnv() (*SMTPSender, error) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("FROM_EMAIL")
	if host == "" || port == "" || user == "" || pass == "" || from == "" {
		return nil, fmt.Errorf("SMTP env vars missing")
	}
	return &SMTPSender{host: host, port: port, user: user, pass: pass, from: from}, nil
}

func (s *SMTPSender) Send(to, subject, bodyHTML, bodyText string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	// Prefer STARTTLS
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", s.from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: multipart/alternative; boundary=boundary42\r\n\r\n"
	msg += "--boundary42\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n"
	msg += bodyText + "\r\n"
	msg += "--boundary42\r\nContent-Type: text/html; charset=utf-8\r\n\r\n"
	msg += bodyHTML + "\r\n"
	msg += "--boundary42--\r\n"

	// Attempt TLS if port 465
	if s.port == "465" {
		c, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.host})
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(c, s.host)
		if err != nil {
			return err
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return err
		}
		if err := client.Mail(s.from); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return client.Quit()
	}

	// STARTTLS / plain send
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}
