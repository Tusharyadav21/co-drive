package middleware

import "net/http"

// SecurityHeaders sets the baseline response headers every production HTTP
// service should send regardless of what else is on the page. A full
// Content-Security-Policy is deliberately not included here: this app loads
// fonts from fonts.googleapis.com/fonts.gstatic.com, and a CSP wrong in
// either direction is worse than none (silently breaks the page, or silently
// protects nothing) — it needs its own reviewed pass, not a default guess.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
