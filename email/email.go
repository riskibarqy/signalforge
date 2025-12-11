package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
	To   []string
}

func Send(ctx context.Context, cfg SMTPConfig, subject, body string) error {
	if cfg.Host == "" || cfg.From == "" || len(cfg.To) == 0 {
		return fmt.Errorf("smtp config incomplete")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	msg := buildMessage(cfg.From, cfg.To, subject, body)

	ch := make(chan error, 1)
	go func() {
		ch <- smtp.SendMail(addr, auth, cfg.From, cfg.To, []byte(msg))
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func buildMessage(from string, to []string, subject, body string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", strings.Join(to, ",")),
		fmt.Sprintf("Subject: %s", subject),
		"Mime-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"",
	}
	return strings.Join(headers, "\r\n") + "\r\n" + body
}
