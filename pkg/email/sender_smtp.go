package email

import (
	"fmt"
	"log"
	"net/smtp"

	"co-drive/config"
)

type SMTPService struct {
	cfg config.SMTPConfig
}

func NewSMTPService(cfg config.SMTPConfig) *SMTPService {
	return &SMTPService{cfg: cfg}
}

func (s *SMTPService) SendEmail(to string, subject string, htmlBody string) error {
	if s.cfg.Host == "" {
		log.Printf("[DEV EMAIL LOG] To: %s | Subject: %s | Body: %s", to, subject, htmlBody)
		return nil
	}

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	headers := make(map[string]string)
	headers["From"] = s.cfg.FromEmail
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	if err := smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email via smtp: %w", err)
	}

	return nil
}

func (s *SMTPService) SendOTP(to string, otp string, ttlMinutes int) error {
	htmlBody, err := RenderOTPEmail(otp, ttlMinutes)
	if err != nil {
		return err
	}
	return s.SendEmail(to, "Your Login Verification Code", htmlBody)
}

func (s *SMTPService) SendPasswordReset(to string, token string, resetURL string, ttlMinutes int) error {
	htmlBody, err := RenderPasswordResetEmail(token, resetURL, ttlMinutes)
	if err != nil {
		return err
	}
	return s.SendEmail(to, "Reset Your Password", htmlBody)
}
