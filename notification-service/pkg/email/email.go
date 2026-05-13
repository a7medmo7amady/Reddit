package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type EmailService interface {
	SendNotification(to string, subject string, body string) error
}

type smtpEmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService() EmailService {
	return &smtpEmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *smtpEmailService) SendNotification(to string, subject string, body string) error {
	if s.host == "" {
		log.Printf("SMTP not configured. Mocking email to %s: [%s] %s", to, subject, body)
		return nil
	}

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
