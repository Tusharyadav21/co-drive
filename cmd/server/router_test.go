package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"co-drive/internal/auth"
	"co-drive/internal/middleware"
	"co-drive/internal/users"
	"co-drive/internal/vehicles"
	"co-drive/pkg/database"
)

// newTestRouter wires the real handlers over an in-memory database. The point
// is to assert the routing table — which paths exist, and which demand a
// session — not the handlers' own behaviour.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	db := database.SetupTestDB(t)

	userRepo := users.NewRepository(db)
	sessionService := auth.NewSessionService(auth.NewSessionRepository(db))

	staticDir := t.TempDir()
	for _, f := range []string{"index.html", "auth.html", "dashboard.html", "profile.html", "logs.html", "favicon.ico", "apple-touch-icon.png", "site.webmanifest"} {
		if err := os.WriteFile(filepath.Join(staticDir, f), []byte("<html>"+f+"</html>"), 0o600); err != nil {
			t.Fatalf("failed to write fixture %s: %v", f, err)
		}
	}

	return NewRouter(RouterDeps{
		AuthHandler: auth.NewHandler(userRepo, sessionService),
		VerificationHandler: auth.NewVerificationHandler(
			auth.NewVerificationService(auth.NewVerificationRepository(db), nil, 10, 15),
			userRepo, sessionService,
		),
		UserHandler:    users.NewHandler(userRepo),
		VehicleHandler: vehicles.NewHandler(vehicles.NewRepository(db)),

		AuthMiddleware: middleware.NewAuth(sessionService),
		CSRFMiddleware: middleware.NewCSRF(),
		RateLimiter:    middleware.NewRateLimiter(1000, time.Minute),

		StaticDir: staticDir,
	})
}

// newRequest builds a request that already satisfies the CSRF middleware, which
// sits in front of auth and rejects any unsafe method without a matching
// cookie/header pair. Without this, every POST would 403 before reaching the
// route we are trying to assert.
func newRequest(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "test-token"})
		req.Header.Set("X-CSRF-Token", "test-token")
	}
	return req
}

// TestRoutesRequireSession pins the public/protected split. Every protected
// route must answer 401 without a session cookie — never 404, which would mean
// the route silently does not exist.
func TestRoutesRequireSession(t *testing.T) {
	r := newTestRouter(t)

	protected := []struct{ method, path string }{
		{http.MethodGet, "/api/users/me"},
		{http.MethodPut, "/api/users/me"},
		{http.MethodPost, "/api/auth/logout"},
		{http.MethodGet, "/api/vehicles"},
		{http.MethodGet, "/api/vehicles/"},
		{http.MethodPost, "/api/vehicles"},
		{http.MethodPost, "/api/vehicles/"},
		{http.MethodGet, "/api/vehicles/abc"},
		{http.MethodDelete, "/api/vehicles/abc"},
		{http.MethodPost, "/api/vehicles/abc/mileage"},
		{http.MethodPost, "/api/vehicles/abc/puc"},
		{http.MethodPost, "/api/vehicles/abc/insurance"},
	}

	for _, tt := range protected {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, newRequest(tt.method, tt.path))

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("got %d, want 401 (404 would mean the route is missing)", rec.Code)
			}
		})
	}
}

// TestPublicRoutes covers the paths that must work with no session at all,
// including both trailing-slash forms of each static page.
func TestPublicRoutes(t *testing.T) {
	r := newTestRouter(t)

	public := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/", http.StatusOK},
		{http.MethodGet, "/auth", http.StatusOK},
		{http.MethodGet, "/auth/", http.StatusOK},
		{http.MethodGet, "/dashboard", http.StatusOK},
		{http.MethodGet, "/dashboard/", http.StatusOK},
		{http.MethodGet, "/app", http.StatusOK},
		{http.MethodGet, "/app/", http.StatusOK},
		{http.MethodGet, "/analytics", http.StatusOK},
		{http.MethodGet, "/analytics/", http.StatusOK},
		{http.MethodGet, "/profile", http.StatusOK},
		{http.MethodGet, "/profile/", http.StatusOK},
		{http.MethodGet, "/account", http.StatusOK},
		{http.MethodGet, "/account/", http.StatusOK},
		{http.MethodGet, "/sharing", http.StatusOK},
		{http.MethodGet, "/sharing/", http.StatusOK},
		{http.MethodGet, "/logs", http.StatusOK},
		{http.MethodGet, "/logs/", http.StatusOK},
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/favicon.ico", http.StatusOK},
		{http.MethodGet, "/apple-touch-icon.png", http.StatusOK},
		{http.MethodGet, "/site.webmanifest", http.StatusOK},
		// Public auth endpoints exist; a bad body is a 400, not a 404 or 401.
		{http.MethodPost, "/api/auth/login", http.StatusBadRequest},
		{http.MethodPost, "/api/auth/register", http.StatusBadRequest},
		{http.MethodPost, "/api/auth/otp/request", http.StatusBadRequest},
		{http.MethodPost, "/api/auth/otp/verify", http.StatusBadRequest},
	}

	for _, tt := range public {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, newRequest(tt.method, tt.path))

			if rec.Code != tt.want {
				t.Errorf("got %d, want %d: %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

// TestStaticPagesAreNotCached pins the fix from commit 5f02511.
func TestStaticPagesAreNotCached(t *testing.T) {
	r := newTestRouter(t)

	for _, path := range []string{"/", "/auth", "/dashboard", "/app", "/analytics", "/profile"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, newRequest(http.MethodGet, path))

		if got := rec.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
			t.Errorf("%s Cache-Control = %q, want no-store", path, got)
		}
	}
}

// TestSecurityHeadersPresent pins the baseline headers every response must
// carry, on both a static page and an API route.
func TestSecurityHeadersPresent(t *testing.T) {
	r := newTestRouter(t)

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for _, path := range []string{"/", "/health"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, newRequest(http.MethodGet, path))

		for header, expected := range want {
			if got := rec.Header().Get(header); got != expected {
				t.Errorf("%s %s = %q, want %q", path, header, got, expected)
			}
		}
	}
}
