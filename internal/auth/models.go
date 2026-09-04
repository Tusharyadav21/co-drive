package auth

import (
	"time"

	"co-drive/internal/session"
)

// Session lives in internal/session so that handler packages can read it off a
// request context without importing internal/auth. Aliased here because the
// session store and lifecycle still belong to this package.
type Session = session.Session

type VerificationType string

const (
	VerificationTypeOTP           VerificationType = "otp"
	VerificationTypePasswordReset VerificationType = "password_reset"
)

type Verification struct {
	ID        string           `json:"id"`
	Email     string           `json:"email"`
	Type      VerificationType `json:"type"`
	CodeHash  string           `json:"-"`
	ExpiresAt time.Time        `json:"expires_at"`
	Used      bool             `json:"used"`
	CreatedAt time.Time        `json:"created_at"`
}
