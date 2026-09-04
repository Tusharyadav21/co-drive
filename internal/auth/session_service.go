package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultSessionDuration is the rolling session lifespan (60 days).
	DefaultSessionDuration = 60 * 24 * time.Hour
	// SessionRefreshThreshold is the inactivity threshold (30 days / 50% of lifespan).
	// When remaining lifetime drops below this, active requests seamlessly renew the session.
	SessionRefreshThreshold = 30 * 24 * time.Hour
)

type SessionService struct {
	repo SessionRepository
}

func NewSessionService(repo SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) CreateSession(ctx context.Context, userID string, userAgent string, ipAddress string) (*Session, error) {
	session := &Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(DefaultSessionDuration),
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.repo.GetByID(ctx, sessionID)
}

func (s *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	return s.repo.Delete(ctx, sessionID)
}

// NeedsRefresh reports whether the session has reached or exceeded 50% of its
// lifetime (less than SessionRefreshThreshold remaining) and should be extended.
func (s *SessionService) NeedsRefresh(session *Session) bool {
	if session == nil {
		return false
	}
	return time.Until(session.ExpiresAt) < SessionRefreshThreshold
}

// ExtendSession extends the session in the repository to now + DefaultSessionDuration.
func (s *SessionService) ExtendSession(ctx context.Context, sessionID string) (*Session, error) {
	newExpiry := time.Now().Add(DefaultSessionDuration)
	if err := s.repo.Extend(ctx, sessionID, newExpiry); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, sessionID)
}

func SetSessionCookie(w http.ResponseWriter, r *http.Request, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   IsRequestHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(DefaultSessionDuration),
		MaxAge:   int(DefaultSessionDuration.Seconds()),
	})
}

// IsRequestHTTPS reports whether the request reached us over TLS, directly or
// terminated at a reverse proxy that sets X-Forwarded-Proto. Shared by every
// cookie-setting and absolute-URL-building call site in this codebase so
// there is exactly one place that knows how to detect the request scheme.
func IsRequestHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
