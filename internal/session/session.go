// Package session owns the authenticated Session and how it travels on a
// request context. It deliberately has no dependency on internal/auth or
// internal/users so that every handler package can read the session without an
// import cycle (internal/auth already imports internal/users).
package session

import (
	"context"
	"net/http"
	"time"

	"co-drive/pkg/response"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ctxKey is unexported so nothing outside this package can write the session
// slot — WithSession is the only way in, FromContext the only way out.
type ctxKey struct{}

func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

func FromContext(ctx context.Context) (*Session, bool) {
	s, ok := ctx.Value(ctxKey{}).(*Session)
	if !ok || s == nil || s.UserID == "" {
		return nil, false
	}
	return s, true
}

// Require is the single 401 path for protected handlers. It returns ok=false
// and has already written the response when there is no session.
func Require(w http.ResponseWriter, r *http.Request) (*Session, bool) {
	s, ok := FromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return nil, false
	}
	return s, true
}
