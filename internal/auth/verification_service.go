package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"co-drive/pkg/email"
)

var (
	ErrCooldownActive      = errors.New("cooldown active")
	ErrAccountLocked       = errors.New("account temporarily locked")
	ErrMaxAttemptsExceeded = errors.New("maximum verification attempts exceeded")
	ErrInvalidOrExpiredOTP = errors.New("invalid or expired OTP code")
)

type CooldownError struct {
	RemainingSeconds int
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("please wait %d seconds before requesting another code", e.RemainingSeconds)
}

type LockoutError struct {
	RemainingSeconds int
}

func (e *LockoutError) Error() string {
	return fmt.Sprintf("too many failed attempts. Try again in %d seconds", e.RemainingSeconds)
}

type emailTracker struct {
	lastRequestedAt time.Time
	failedAttempts  int
	lockedUntil     time.Time
}

type VerificationService struct {
	repo         VerificationRepository
	emailService email.Service
	otpTTL       int
	resetTTL     int

	mu       sync.Mutex
	trackers map[string]*emailTracker
}

func NewVerificationService(repo VerificationRepository, emailService email.Service, otpTTL int, resetTTL int) *VerificationService {
	return &VerificationService{
		repo:         repo,
		emailService: emailService,
		otpTTL:       otpTTL,
		resetTTL:     resetTTL,
		trackers:     make(map[string]*emailTracker),
	}
}

func (s *VerificationService) GenerateAndSendOTP(ctx context.Context, userEmail string) error {
	cleanEmail := strings.ToLower(strings.TrimSpace(userEmail))
	if cleanEmail == "" {
		return errors.New("email is required")
	}

	s.mu.Lock()
	tracker, exists := s.trackers[cleanEmail]
	if !exists {
		tracker = &emailTracker{}
		s.trackers[cleanEmail] = tracker
	}

	now := time.Now()
	// Check if account is temporarily locked
	if now.Before(tracker.lockedUntil) {
		remaining := int(tracker.lockedUntil.Sub(now).Seconds()) + 1
		s.mu.Unlock()
		return &LockoutError{RemainingSeconds: remaining}
	}

	// Check 60-second cooldown between requests
	cooldown := 60 * time.Second
	if !tracker.lastRequestedAt.IsZero() && now.Sub(tracker.lastRequestedAt) < cooldown {
		remaining := int((cooldown - now.Sub(tracker.lastRequestedAt)).Seconds()) + 1
		s.mu.Unlock()
		return &CooldownError{RemainingSeconds: remaining}
	}

	tracker.lastRequestedAt = now
	tracker.failedAttempts = 0
	s.mu.Unlock()

	otp, err := s.generateNumericOTP(6)
	if err != nil {
		return fmt.Errorf("failed to generate otp: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash otp: %w", err)
	}

	v := &Verification{
		ID:        uuid.NewString(),
		Email:     cleanEmail,
		Type:      VerificationTypeOTP,
		CodeHash:  string(hash),
		ExpiresAt: now.Add(time.Duration(s.otpTTL) * time.Minute),
		Used:      false,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return err
	}

	return s.emailService.SendOTP(cleanEmail, otp, s.otpTTL)
}

func (s *VerificationService) VerifyOTP(ctx context.Context, userEmail string, otp string) (bool, int, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(userEmail))
	cleanOTP := strings.TrimSpace(otp)

	s.mu.Lock()
	tracker, exists := s.trackers[cleanEmail]
	if !exists {
		tracker = &emailTracker{}
		s.trackers[cleanEmail] = tracker
	}

	now := time.Now()
	if now.Before(tracker.lockedUntil) {
		remaining := int(tracker.lockedUntil.Sub(now).Seconds()) + 1
		s.mu.Unlock()
		return false, 0, &LockoutError{RemainingSeconds: remaining}
	}
	s.mu.Unlock()

	v, err := s.repo.GetLatestValid(ctx, cleanEmail, VerificationTypeOTP)
	if err != nil || v == nil {
		return false, 0, ErrInvalidOrExpiredOTP
	}

	if err := bcrypt.CompareHashAndPassword([]byte(v.CodeHash), []byte(cleanOTP)); err != nil {
		s.mu.Lock()
		tracker.failedAttempts++
		attemptsLeft := 5 - tracker.failedAttempts
		if attemptsLeft <= 0 {
			// Lock out email for 5 minutes and burn the OTP
			tracker.lockedUntil = now.Add(5 * time.Minute)
			tracker.failedAttempts = 0
			s.mu.Unlock()
			_ = s.repo.MarkAsUsed(ctx, v.ID)
			return false, 0, ErrMaxAttemptsExceeded
		}
		s.mu.Unlock()
		return false, attemptsLeft, ErrInvalidOrExpiredOTP
	}

	// Success: Reset tracker and mark OTP as used
	s.mu.Lock()
	tracker.failedAttempts = 0
	tracker.lockedUntil = time.Time{}
	s.mu.Unlock()

	if err := s.repo.MarkAsUsed(ctx, v.ID); err != nil {
		return false, 0, err
	}

	return true, 5, nil
}

func (s *VerificationService) GenerateAndSendPasswordReset(ctx context.Context, userEmail string, resetBaseURL string) error {
	cleanEmail := strings.ToLower(strings.TrimSpace(userEmail))
	token := uuid.NewString()
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash reset token: %w", err)
	}

	v := &Verification{
		ID:        uuid.NewString(),
		Email:     cleanEmail,
		Type:      VerificationTypePasswordReset,
		CodeHash:  string(hash),
		ExpiresAt: time.Now().Add(time.Duration(s.resetTTL) * time.Minute),
		Used:      false,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s?token=%s&email=%s", resetBaseURL, token, cleanEmail)
	return s.emailService.SendPasswordReset(cleanEmail, token, resetURL, s.resetTTL)
}

func (s *VerificationService) VerifyPasswordReset(ctx context.Context, userEmail string, token string) (bool, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(userEmail))
	cleanToken := strings.TrimSpace(token)

	v, err := s.repo.GetLatestValid(ctx, cleanEmail, VerificationTypePasswordReset)
	if err != nil || v == nil {
		return false, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(v.CodeHash), []byte(cleanToken)); err != nil {
		return false, nil
	}

	if err := s.repo.MarkAsUsed(ctx, v.ID); err != nil {
		return false, err
	}

	return true, nil
}

func (s *VerificationService) generateNumericOTP(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}
	return string(result), nil
}
