package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"co-drive/internal/session"
	"co-drive/internal/users"
	"co-drive/pkg/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	return database.SetupTestDB(t)
}

type mockEmailService struct {
	lastOTP   string
	lastToken string
}

func (m *mockEmailService) SendEmail(to string, subject string, htmlBody string) error {
	return nil
}

func (m *mockEmailService) SendOTP(to, otp string, ttlMinutes int) error {
	m.lastOTP = otp
	return nil
}

func (m *mockEmailService) SendPasswordReset(to, token, resetURL string, ttlMinutes int) error {
	m.lastToken = token
	return nil
}

func TestEmailOTPVerificationFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := users.NewRepository(db)
	sessionRepo := NewSessionRepository(db)
	verificationRepo := NewVerificationRepository(db)
	mockEmail := &mockEmailService{}

	verificationService := NewVerificationService(verificationRepo, mockEmail, 10, 15)
	sessionService := NewSessionService(sessionRepo)

	handler := NewVerificationHandler(verificationService, userRepo, sessionService)
	authHandler := NewHandler(userRepo, sessionService)

	// 1. Request OTP
	reqOTPBody, _ := json.Marshal(RequestOTPRequest{Email: "tushar.yadav@example.com"})
	reqOTP := httptest.NewRequest(http.MethodPost, "/api/auth/otp/request", bytes.NewBuffer(reqOTPBody))
	recOTP := httptest.NewRecorder()

	handler.RequestOTP(recOTP, reqOTP)
	if recOTP.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK on OTP request, got %d: %s", recOTP.Code, recOTP.Body.String())
	}

	if mockEmail.lastOTP == "" {
		t.Fatal("mock email service did not receive OTP code")
	}

	// 2. Verify OTP with Invalid Code (checks attempts_left)
	verifyInvalidBody, _ := json.Marshal(VerifyOTPRequest{Email: "tushar.yadav@example.com", OTP: "000000"})
	reqVerifyInvalid := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBuffer(verifyInvalidBody))
	recVerifyInvalid := httptest.NewRecorder()

	handler.VerifyOTP(recVerifyInvalid, reqVerifyInvalid)
	if recVerifyInvalid.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for bad OTP, got %d", recVerifyInvalid.Code)
	}

	var invalidResp map[string]any
	_ = json.Unmarshal(recVerifyInvalid.Body.Bytes(), &invalidResp)
	if attempts, ok := invalidResp["attempts_left"].(float64); !ok || attempts != 4 {
		t.Errorf("expected 4 attempts_left, got %v", invalidResp["attempts_left"])
	}

	// 3. Verify OTP with Correct Code (Auto-registers user with default name)
	verifyValidBody, _ := json.Marshal(VerifyOTPRequest{Email: "tushar.yadav@example.com", OTP: mockEmail.lastOTP})
	reqVerifyValid := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBuffer(verifyValidBody))
	recVerifyValid := httptest.NewRecorder()

	handler.VerifyOTP(recVerifyValid, reqVerifyValid)
	if recVerifyValid.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for valid OTP, got %d: %s", recVerifyValid.Code, recVerifyValid.Body.String())
	}

	// Verify user created, default name derived, and marked verified
	user, err := userRepo.GetByEmail(context.Background(), "tushar.yadav@example.com")
	if err != nil || user == nil {
		t.Fatalf("user was not auto-created on OTP verification")
	}
	if user.FullName != "Tushar Yadav" {
		t.Errorf("expected derived default name 'Tushar Yadav', got %q", user.FullName)
	}
	if !user.EmailVerified {
		t.Error("user email_verified was not set to true")
	}

	// Verify session cookie set
	cookies := recVerifyValid.Result().Cookies()
	var sessionID string
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionID = c.Value
		}
	}
	if sessionID == "" {
		t.Fatal("session_id cookie was not set on OTP verification")
	}

	// 4. Logout
	reqLogout := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	ctx := session.WithSession(reqLogout.Context(), &session.Session{ID: sessionID, UserID: user.ID})
	recLogout := httptest.NewRecorder()

	authHandler.Logout(recLogout, reqLogout.WithContext(ctx))
	if recLogout.Code != http.StatusOK {
		t.Errorf("expected status 200 OK on logout, got %d", recLogout.Code)
	}

	// Verify session revoked
	revocationCheck, _ := sessionService.GetSession(context.Background(), sessionID)
	if revocationCheck != nil {
		t.Error("session was not revoked after logout")
	}
}

func TestEmailOTPRateLimitingAndLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := users.NewRepository(db)
	sessionRepo := NewSessionRepository(db)
	verificationRepo := NewVerificationRepository(db)
	mockEmail := &mockEmailService{}

	verificationService := NewVerificationService(verificationRepo, mockEmail, 10, 15)
	sessionService := NewSessionService(sessionRepo)
	handler := NewVerificationHandler(verificationService, userRepo, sessionService)

	email := "cooldown@example.com"

	// First request succeeds
	reqBody, _ := json.Marshal(RequestOTPRequest{Email: email})
	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/otp/request", bytes.NewBuffer(reqBody))
	rec1 := httptest.NewRecorder()
	handler.RequestOTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on first request, got %d", rec1.Code)
	}

	// Immediate second request is rate-limited by 60s cooldown
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/otp/request", bytes.NewBuffer(reqBody))
	rec2 := httptest.NewRecorder()
	handler.RequestOTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests on immediate second request, got %d", rec2.Code)
	}
	if rec2.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header to be set on 429 response")
	}

	// 4 failed attempts reduces attempts_left
	for i := 1; i <= 4; i++ {
		badBody, _ := json.Marshal(VerifyOTPRequest{Email: email, OTP: "999999"})
		reqBad := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBuffer(badBody))
		recBad := httptest.NewRecorder()
		handler.VerifyOTP(recBad, reqBad)
		if recBad.Code != http.StatusUnauthorized {
			t.Errorf("attempt %d: expected 401, got %d", i, recBad.Code)
		}
	}

	// 5th failed attempt should trigger lockout & 429 Too Many Requests
	badBody5, _ := json.Marshal(VerifyOTPRequest{Email: email, OTP: "999999"})
	reqBad5 := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBuffer(badBody5))
	recBad5 := httptest.NewRecorder()
	handler.VerifyOTP(recBad5, reqBad5)
	if recBad5.Code != http.StatusTooManyRequests {
		t.Errorf("5th attempt: expected 429 lockout, got %d", recBad5.Code)
	}
}

func TestDeprecatedPasswordAuth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := users.NewRepository(db)
	sessionRepo := NewSessionRepository(db)
	sessionService := NewSessionService(sessionRepo)
	handler := NewHandler(userRepo, sessionService)

	recReg := httptest.NewRecorder()
	reqReg := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
	handler.Register(recReg, reqReg)
	if recReg.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on password register, got %d", recReg.Code)
	}

	recLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	handler.Login(recLogin, reqLogin)
	if recLogin.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on password login, got %d", recLogin.Code)
	}
}

func TestRollingSessionLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := users.NewRepository(db)
	sessionRepo := NewSessionRepository(db)
	sessionService := NewSessionService(sessionRepo)
	ctx := context.Background()

	_ = userRepo.Create(ctx, &users.User{ID: "test-user-id", Email: "sessiontest@example.com", FullName: "Session Test"})

	// 1. CreateSession defaults to 60 days
	sess, err := sessionService.CreateSession(ctx, "test-user-id", "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	expectedExpiry := time.Now().Add(DefaultSessionDuration)
	if diff := sess.ExpiresAt.Sub(expectedExpiry); diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expected ExpiresAt around %v, got %v (diff %v)", expectedExpiry, sess.ExpiresAt, diff)
	}

	// 2. SetSessionCookie includes MaxAge and 60-day expiry
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SetSessionCookie(rec, req, sess.ID)

	cookies := rec.Result().Cookies()
	var sc *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			sc = c
			break
		}
	}
	if sc == nil {
		t.Fatal("session_id cookie was not set")
	}
	if sc.MaxAge != int(DefaultSessionDuration.Seconds()) {
		t.Errorf("expected MaxAge = %d, got %d", int(DefaultSessionDuration.Seconds()), sc.MaxAge)
	}

	// 3. NeedsRefresh threshold check
	// Fresh session (60 days left) -> does NOT need refresh
	if sessionService.NeedsRefresh(sess) {
		t.Errorf("newly created session should not need refresh")
	}

	// Simulated session with only 10 days left (< 30 days threshold) -> needs refresh
	staleSession := &Session{
		ID:        sess.ID,
		ExpiresAt: time.Now().Add(10 * 24 * time.Hour),
	}
	if !sessionService.NeedsRefresh(staleSession) {
		t.Errorf("session with 10 days left should need refresh")
	}

	// 4. ExtendSession updates the database to 60 days from now
	extended, err := sessionService.ExtendSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("failed to extend session: %v", err)
	}

	persisted, err := sessionRepo.GetByID(ctx, sess.ID)
	if err != nil || persisted == nil {
		t.Fatalf("failed to get persisted session: %v", err)
	}
	if diff := persisted.ExpiresAt.Sub(extended.ExpiresAt); diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("persisted expiry %v does not match extended %v", persisted.ExpiresAt, extended.ExpiresAt)
	}
	if diff := persisted.ExpiresAt.Sub(time.Now().Add(DefaultSessionDuration)); diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("extended expiry not ~60 days in future: %v", persisted.ExpiresAt)
	}

	// 5. ClearSessionCookie sets MaxAge = -1
	recClear := httptest.NewRecorder()
	ClearSessionCookie(recClear)
	clearCookies := recClear.Result().Cookies()
	var clearC *http.Cookie
	for _, c := range clearCookies {
		if c.Name == "session_id" {
			clearC = c
			break
		}
	}
	if clearC == nil {
		t.Fatal("expected clear session cookie to be set")
	}
	if clearC.MaxAge != -1 {
		t.Errorf("expected clear cookie MaxAge = -1, got %d", clearC.MaxAge)
	}
}
