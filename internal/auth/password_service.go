package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const MinPasswordLength = 8

// ValidatePassword is the single length gate for every path that sets a
// password (register, reset). bcrypt's own max (72 bytes) is already
// enforced by HashPassword returning an error, so only a minimum belongs here.
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
