package middleware

import (
	"net/http"

	"co-drive/internal/auth"
	"co-drive/internal/session"
	"co-drive/pkg/response"
)

type Auth struct {
	sessionService *auth.SessionService
}

func NewAuth(sessionService *auth.SessionService) *Auth {
	return &Auth{sessionService: sessionService}
}

func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			response.Error(w, http.StatusUnauthorized, "Unauthorized: missing session cookie")
			return
		}

		s, err := a.sessionService.GetSession(r.Context(), cookie.Value)
		if err != nil || s == nil {
			response.Error(w, http.StatusUnauthorized, "Unauthorized: invalid or expired session")
			return
		}

		// Rolling session: when an active user interacts and their session has less than
		// SessionRefreshThreshold remaining (e.g., 30 days out of 60 days), extend it
		// in the database and re-issue the session_id cookie so active users stay logged in.
		if a.sessionService.NeedsRefresh(s) {
			if extended, err := a.sessionService.ExtendSession(r.Context(), s.ID); err == nil && extended != nil {
				s = extended
				auth.SetSessionCookie(w, r, s.ID)
			}
		}

		next.ServeHTTP(w, r.WithContext(session.WithSession(r.Context(), s)))
	})
}
