package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"co-drive/internal/users"
	"co-drive/pkg/response"
)

type VerificationHandler struct {
	verificationService *VerificationService
	userRepo            users.Repository
	sessionService      *SessionService
}

func NewVerificationHandler(
	verificationService *VerificationService,
	userRepo users.Repository,
	sessionService *SessionService,
) *VerificationHandler {
	return &VerificationHandler{
		verificationService: verificationService,
		userRepo:            userRepo,
		sessionService:      sessionService,
	}
}

type RequestOTPRequest struct {
	Email string `json:"email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (h *VerificationHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestOTPRequest
	if err := response.DecodeJSON(r, &req); err != nil || req.Email == "" {
		response.Error(w, http.StatusBadRequest, "Valid email address is required")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if err := h.verificationService.GenerateAndSendOTP(r.Context(), req.Email); err != nil {
		var cdErr *CooldownError
		var lockErr *LockoutError
		if errors.As(err, &cdErr) {
			w.Header().Set("Retry-After", strconv.Itoa(cdErr.RemainingSeconds))
			response.JSON(w, http.StatusTooManyRequests, map[string]any{
				"error":        cdErr.Error(),
				"retry_after": cdErr.RemainingSeconds,
			})
			return
		}
		if errors.As(err, &lockErr) {
			w.Header().Set("Retry-After", strconv.Itoa(lockErr.RemainingSeconds))
			response.JSON(w, http.StatusTooManyRequests, map[string]any{
				"error":        lockErr.Error(),
				"retry_after": lockErr.RemainingSeconds,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to send OTP code")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message":          "Verification OTP sent to email",
		"cooldown_seconds": 60,
	})
}

func (h *VerificationHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := response.DecodeJSON(r, &req); err != nil || req.Email == "" || req.OTP == "" {
		response.Error(w, http.StatusBadRequest, "Email and OTP are required")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	valid, attemptsLeft, err := h.verificationService.VerifyOTP(r.Context(), req.Email, req.OTP)
	if err != nil || !valid {
		var lockErr *LockoutError
		if errors.As(err, &lockErr) {
			w.Header().Set("Retry-After", strconv.Itoa(lockErr.RemainingSeconds))
			response.JSON(w, http.StatusTooManyRequests, map[string]any{
				"error":        lockErr.Error(),
				"retry_after": lockErr.RemainingSeconds,
			})
			return
		}
		if errors.Is(err, ErrMaxAttemptsExceeded) {
			w.Header().Set("Retry-After", "300")
			response.JSON(w, http.StatusTooManyRequests, map[string]any{
				"error":         "Maximum failed attempts exceeded. Code invalidated. Please wait 5 minutes before requesting a new code.",
				"attempts_left": 0,
				"retry_after":   300,
			})
			return
		}
		response.JSON(w, http.StatusUnauthorized, map[string]any{
			"error":         "Invalid or expired OTP code",
			"attempts_left": attemptsLeft,
		})
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	if user == nil {
		// Auto-register user on first passwordless OTP login with friendly default name
		user = &users.User{
			ID:            uuid.NewString(),
			Email:         req.Email,
			FullName:      deriveDefaultName(req.Email),
			EmailVerified: true,
		}
		if err := h.userRepo.Create(r.Context(), user); err != nil {
			response.Error(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
	} else if !user.EmailVerified {
		_ = h.userRepo.UpdateEmailVerified(r.Context(), user.ID, true)
	}

	sess, err := h.sessionService.CreateSession(r.Context(), user.ID, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	SetSessionCookie(w, r, sess.ID)
	response.JSON(w, http.StatusOK, user)
}

// deriveDefaultName creates a clean display name from the local part of an email.
// e.g. "john.doe@example.com" -> "John Doe", "driver@example.com" -> "Driver"
func deriveDefaultName(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 && parts[0] != "" {
		namePart := parts[0]
		namePart = strings.ReplaceAll(namePart, ".", " ")
		namePart = strings.ReplaceAll(namePart, "_", " ")
		namePart = strings.ReplaceAll(namePart, "-", " ")
		words := strings.Fields(namePart)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
			}
		}
		if len(words) > 0 {
			return strings.Join(words, " ")
		}
	}
	return "Driver"
}
