package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/google/uuid"

	"co-drive/internal/auth"
	"co-drive/pkg/response"
)

// CSRF implements the double-submit cookie pattern: a request is legitimate
// only if it can both read the csrf_token cookie (same-origin JS can) and echo
// it back in a header. That comparison needs no server-side secret.
type CSRF struct{}

func NewCSRF() *CSRF {
	return &CSRF{}
}

func (c *CSRF) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			cookie, err := r.Cookie("csrf_token")
			if err != nil || cookie.Value == "" {
				c.setTokenCookie(w, r)
			}
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("csrf_token")
		if err != nil || cookie.Value == "" {
			response.Error(w, http.StatusForbidden, "CSRF token cookie missing")
			return
		}

		headerToken := r.Header.Get("X-CSRF-Token")
		if headerToken == "" || subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookie.Value)) != 1 {
			response.Error(w, http.StatusForbidden, "Invalid CSRF token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *CSRF) setTokenCookie(w http.ResponseWriter, r *http.Request) {
	token := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // same-origin JS must read this to echo it back in a header
		Secure:   auth.IsRequestHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}
