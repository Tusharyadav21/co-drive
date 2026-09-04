package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"co-drive/internal/auth"
	"co-drive/internal/session"
	"co-drive/internal/users"
	"co-drive/pkg/database"
)

func TestAuthMiddlewareAndRollingSession(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	userRepo := users.NewRepository(db)
	sessionRepo := auth.NewSessionRepository(db)
	sessionService := auth.NewSessionService(sessionRepo)
	authMiddleware := NewAuth(sessionService)

	ctx := context.Background()

	// Seed test users
	_ = userRepo.Create(ctx, &users.User{ID: "user-fresh", Email: "fresh@example.com", FullName: "Fresh User"})
	_ = userRepo.Create(ctx, &users.User{ID: "user-rolling", Email: "rolling@example.com", FullName: "Rolling User"})

	// Handler that returns 200 OK and echos user_id from session context
	testHandler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, ok := session.FromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.UserID))
	}))

	// Case 1: Missing session cookie -> 401
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	testHandler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing cookie, got %d", rec1.Code)
	}

	// Case 2: Invalid/non-existent session cookie -> 401
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req2.AddCookie(&http.Cookie{Name: "session_id", Value: "non-existent-session-id"})
	testHandler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad session id, got %d", rec2.Code)
	}

	// Case 3: Fresh session (no refresh needed)
	freshSess, err := sessionService.CreateSession(ctx, "user-fresh", "agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req3.AddCookie(&http.Cookie{Name: "session_id", Value: freshSess.ID})
	testHandler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200 for fresh session, got %d", rec3.Code)
	}
	if rec3.Body.String() != "user-fresh" {
		t.Errorf("expected body 'user-fresh', got %q", rec3.Body.String())
	}
	// No refreshed session cookie should be emitted because session has > 30 days left
	cookies3 := rec3.Result().Cookies()
	var renewedCookie3 *http.Cookie
	for _, c := range cookies3 {
		if c.Name == "session_id" {
			renewedCookie3 = c
			break
		}
	}
	if renewedCookie3 != nil {
		t.Errorf("fresh session should not re-issue session_id cookie")
	}

	// Case 4: Rolling session renewal (< 30 days left)
	// Create a session and artificially age it so it has only 10 days remaining
	staleSess, err := sessionService.CreateSession(ctx, "user-rolling", "agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	oldExpiry := time.Now().Add(10 * 24 * time.Hour)
	if err := sessionRepo.Extend(ctx, staleSess.ID, oldExpiry); err != nil {
		t.Fatalf("failed to artificially age session: %v", err)
	}

	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req4.AddCookie(&http.Cookie{Name: "session_id", Value: staleSess.ID})
	testHandler.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusOK {
		t.Errorf("expected 200 for rolling session, got %d", rec4.Code)
	}
	if rec4.Body.String() != "user-rolling" {
		t.Errorf("expected body 'user-rolling', got %q", rec4.Body.String())
	}

	// Verify that a refreshed cookie WAS emitted
	cookies4 := rec4.Result().Cookies()
	var renewedCookie4 *http.Cookie
	for _, c := range cookies4 {
		if c.Name == "session_id" {
			renewedCookie4 = c
			break
		}
	}
	if renewedCookie4 == nil {
		t.Fatal("expected rolling session to re-issue refreshed session_id cookie")
	}
	if renewedCookie4.MaxAge != int(auth.DefaultSessionDuration.Seconds()) {
		t.Errorf("expected refreshed cookie MaxAge %d, got %d", int(auth.DefaultSessionDuration.Seconds()), renewedCookie4.MaxAge)
	}

	// Verify that the database record was renewed to ~60 days
	updatedInDB, err := sessionRepo.GetByID(ctx, staleSess.ID)
	if err != nil || updatedInDB == nil {
		t.Fatalf("failed to query updated session from DB: %v", err)
	}
	if diff := updatedInDB.ExpiresAt.Sub(time.Now().Add(auth.DefaultSessionDuration)); diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expected DB session expiration to be ~60 days, got %v (diff %v)", updatedInDB.ExpiresAt, diff)
	}
}
