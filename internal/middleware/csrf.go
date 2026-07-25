package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	csrfCookieName = "__csrf"
	csrfFormField  = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenLen   = 32
)

// CSRFProtect implements double-submit cookie CSRF protection.
// On every request, it ensures a CSRF cookie exists (generating one if needed).
// On POST/PUT/PATCH/DELETE requests, it validates that the submitted token
// (from form field or header) matches the cookie value.
func CSRFProtect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure CSRF cookie exists on every request
		token := getOrSetCSRFCookie(w, r)

		// Safe methods don't need validation
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Validate CSRF token on state-changing methods
		submitted := r.FormValue(csrfFormField)
		if submitted == "" {
			submitted = r.Header.Get(csrfHeaderName)
		}

		if submitted == "" || !constantTimeCompare(token, submitted) {
			http.Error(w, "Forbidden: invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getOrSetCSRFCookie returns the existing CSRF token from the cookie,
// or generates a new one, sets the cookie, and returns it.
func getOrSetCSRFCookie(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(csrfCookieName)
	if err == nil && cookie.Value != "" && len(cookie.Value) == csrfTokenLen*2 {
		return cookie.Value
	}

	token := generateCSRFToken()
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // JS needs to read this for HTMX headers
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0, // Session cookie — expires when browser closes
	})
	return token
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLen)
	if _, err := rand.Read(b); err != nil {
		// This should never happen on a modern OS, but if it does, panic
		// rather than returning a predictable token.
		panic("csrf: failed to read random bytes: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func constantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	// Check common reverse-proxy headers
	proto := strings.ToLower(r.Header.Get("X-Forwarded-Proto"))
	return proto == "https"
}

// GetCSRFToken extracts the current CSRF token from the request cookie.
// Use this in template rendering to inject the token into forms.
func GetCSRFToken(r *http.Request) string {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
