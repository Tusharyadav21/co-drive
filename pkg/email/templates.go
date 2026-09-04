package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var TemplatesFS embed.FS

type OTPData struct {
	OTP        string
	TTLMinutes int
}

type PasswordResetData struct {
	Token      string
	ResetURL   string
	TTLMinutes int
}

func RenderOTPEmail(otp string, ttlMinutes int) (string, error) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/otp.html")
	if err != nil {
		return "", fmt.Errorf("failed to parse otp template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, OTPData{OTP: otp, TTLMinutes: ttlMinutes}); err != nil {
		return "", fmt.Errorf("failed to render otp template: %w", err)
	}

	return buf.String(), nil
}

func RenderPasswordResetEmail(token string, resetURL string, ttlMinutes int) (string, error) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/password_reset.html")
	if err != nil {
		return "", fmt.Errorf("failed to parse password reset template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, PasswordResetData{Token: token, ResetURL: resetURL, TTLMinutes: ttlMinutes}); err != nil {
		return "", fmt.Errorf("failed to render password reset template: %w", err)
	}

	return buf.String(), nil
}
