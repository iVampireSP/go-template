package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// Config SMTP 邮件发送配置。
type Config struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	Encryption string // "ssl", "tls", "none"
}

// Content 已渲染的邮件成品内容。
type Content struct {
	Subject string
	HTML    string
}

// Mailable 可发送邮件契约：实现方负责渲染，返回成品。
type Mailable interface {
	Content(ctx context.Context) (Content, error)
}

// Email 邮件发送服务。
type Email struct {
	cfg Config
}

// NewEmail 创建邮件发送服务。
func NewEmail(cfg Config) *Email {
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.Encryption == "" {
		cfg.Encryption = "tls"
	}
	return &Email{cfg: cfg}
}

// Send 发送模板邮件。
func (e *Email) Send(ctx context.Context, to []string, mail Mailable) error {
	content, err := mail.Content(ctx)
	if err != nil {
		return fmt.Errorf("failed to build mail content: %w", err)
	}
	return e.sendSMTP(ctx, to, content.Subject, content.HTML)
}

// SendRaw 发送原始邮件。
func (e *Email) SendRaw(ctx context.Context, to []string, subject, body string) error {
	return e.sendSMTP(ctx, to, subject, body)
}

func (e *Email) sendSMTP(_ context.Context, to []string, subject, body string) error {
	if e.cfg.Host == "" {
		return fmt.Errorf("email service is not configured")
	}

	msg := buildMessage(e.cfg.From, to, subject, body)
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)

	switch e.cfg.Encryption {
	case "ssl":
		return e.sendSSL(addr, to, msg)
	case "none":
		return e.sendPlain(addr, to, msg)
	default:
		return e.sendSTARTTLS(addr, to, msg)
	}
}

func (e *Email) sendSTARTTLS(addr string, to []string, msg []byte) error {
	var auth smtp.Auth
	if e.cfg.Username != "" && e.cfg.Password != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}
	return smtp.SendMail(addr, auth, e.cfg.From, to, msg)
}

func (e *Email) sendSSL(addr string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: e.cfg.Host})
	if err != nil {
		return fmt.Errorf("ssl dial: %w", err)
	}
	defer conn.Close()
	return e.sendViaClient(conn, to, msg)
}

func (e *Email) sendPlain(addr string, to []string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("plain dial: %w", err)
	}
	defer conn.Close()
	return e.sendViaClient(conn, to, msg)
}

func (e *Email) sendViaClient(conn net.Conn, to []string, msg []byte) error {
	c, err := smtp.NewClient(conn, e.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if e.cfg.Username != "" && e.cfg.Password != "" {
		if err := c.Auth(smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := c.Mail(e.cfg.From); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", addr, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return c.Quit()
}

func buildMessage(from string, to []string, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
