package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

// Sender email отправитель
type Sender struct {
	Host      string
	Port      int
	Username  string
	Password  string
	From      string
	UseTLS    bool
	SkipVerify bool
}

// Config конфигурация SMTP
type Config struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	UseTLS     bool
	SkipVerify bool
}

// NewSender создает новый SMTP отправитель
func NewSender(config *Config) *Sender {
	return &Sender{
		Host:      config.Host,
		Port:      config.Port,
		Username:  config.Username,
		Password:  config.Password,
		From:     config.From,
		UseTLS:   config.UseTLS,
		SkipVerify: config.SkipVerify,
	}
}

// Send отправляет email
func (s *Sender) Send(to, subject, body, htmlBody string) error {
	var auth smtp.Auth
	if s.Username != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	// Определяем адрес сервера
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	// Настраиваем TLS если нужно
	var conn net.Conn
	var err error
	
	if s.UseTLS {
		tlsConfig := &tls.Config{
			ServerName:         s.Host,
			InsecureSkipVerify: s.SkipVerify,
		}
		conn, err = tls.Dial("tcp", addr, tlsConfig)
	} else {
		conn, err = net.DialTimeout("tcp", addr, 30*time.Second)
	}
	
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Создаем клиент
	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Аутентификация
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	// Отправитель
	if err := client.Mail(s.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Получатель
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Тело письма
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer w.Close()

	// Формируем письмо
	message := s.buildMessage(to, subject, body, htmlBody)
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}

	return nil
}

// buildMessage формирует MIME сообщение
func (s *Sender) buildMessage(to, subject, body, htmlBody string) string {
	message := fmt.Sprintf("From: %s\r\n", s.From)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	
	if htmlBody != "" {
		message += "Content-Type: multipart/alternative; boundary=boundary\r\n\r\n"
		message += "--boundary\r\n"
		message += "Content-Type: text/plain; charset=utf-8\r\n\r\n"
		message += body + "\r\n\r\n"
		message += "--boundary\r\n"
		message += "Content-Type: text/html; charset=utf-8\r\n\r\n"
		message += htmlBody + "\r\n\r\n"
		message += "--boundary--\r\n"
	} else {
		message += "Content-Type: text/plain; charset=utf-8\r\n\r\n"
		message += body
	}

	return message
}

// SendBatch отправляет несколько писем
func (s *Sender) SendBatch(recipients []string, subject, body, htmlBody string) map[string]error {
	results := make(map[string]error)
	
	for _, to := range recipients {
		err := s.Send(to, subject, body, htmlBody)
		if err != nil {
			results[to] = err
		}
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}
	
	return results
}
