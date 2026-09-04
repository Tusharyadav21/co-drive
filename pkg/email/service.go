package email

type Service interface {
	SendEmail(to string, subject string, htmlBody string) error
	SendOTP(to string, otp string, ttlMinutes int) error
	SendPasswordReset(to string, token string, resetURL string, ttlMinutes int) error
}
