package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCSRFCookieSecureFlagMatchesScheme pins the Secure attribute to the
// request's actual scheme: plain HTTP in local dev must still work (a
// hardcoded Secure=true would silently break every non-TLS deployment,
// since browsers refuse to send Secure cookies back over HTTP), while a
// request behind a TLS-terminating proxy must get Secure=true.
func TestCSRFCookieSecureFlagMatchesScheme(t *testing.T) {
	csrf := NewCSRF()
	handler := csrf.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	cases := []struct {
		name           string
		forwardedProto string
		wantSecure     bool
	}{
		{"plain http", "", false},
		{"behind a TLS-terminating proxy", "https", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", tc.forwardedProto)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			cookies := rec.Result().Cookies()
			if len(cookies) != 1 || cookies[0].Name != "csrf_token" {
				t.Fatalf("expected exactly one csrf_token cookie, got %v", cookies)
			}
			if cookies[0].Secure != tc.wantSecure {
				t.Errorf("Secure = %v, want %v", cookies[0].Secure, tc.wantSecure)
			}
		})
	}
}
