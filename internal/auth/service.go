package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"co-drive/internal/db"
)

type Config struct {
	SessionSecret   string
	CookieDomain    string
	GoogleClientID  string
	GoogleSecret    string
	GoogleRedirect  string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
	DevLoginEnabled bool
}

// Context keys for storing user and session in request context
type contextKey string

const (
	UserContextKey    contextKey = "user"
	SessionContextKey contextKey = "session"
)

type Service struct {
	Queries    *db.Queries
	Config     Config
	oauth2Conf *oauth2.Config
}

func NewService(pool *pgxpool.Pool, cfg Config) *Service {
	return &Service{
		Queries: db.NewQueries(pool),
		Config:  cfg,
		oauth2Conf: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleSecret,
			RedirectURL:  cfg.GoogleRedirect,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (s *Service) StartExpiryChecker(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	s.checkExpiry(ctx)
	s.cleanupExpiredSessions(ctx)

	for {
		select {
		case <-ticker.C:
			s.checkExpiry(ctx)
			s.cleanupExpiredSessions(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) checkExpiry(ctx context.Context) {
	docs, err := s.Queries.GetExpiringDocuments(ctx, 7, 15)
	if err != nil {
		slog.Error("Failed to get expiring documents", "error", err)
		return
	}

	if len(docs) == 0 {
		return
	}

	slog.Info("Found expiring documents", "count", len(docs))
}

func (s *Service) cleanupExpiredSessions(ctx context.Context) {
	if err := s.Queries.DeleteExpiredSessions(ctx); err != nil {
		slog.Error("Failed to cleanup expired sessions", "error", err)
	} else {
		slog.Info("Cleaned up expired sessions")
	}
}

// GenerateSession creates a new session for a user, invalidating all previous sessions.
func (s *Service) GenerateSession(ctx context.Context, userID string) (string, *db.Session, error) {
	// Session rotation: invalidate all existing sessions for this user
	if err := s.Queries.DeleteUserSessions(ctx, userID); err != nil {
		slog.Error("Failed to delete old sessions during rotation", "userID", userID, "error", err)
		// Continue — creating a new session is more important than failing here
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	rawToken := hex.EncodeToString(token)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	session, err := s.Queries.CreateSession(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return "", nil, err
	}

	return rawToken, session, nil
}

// ValidateSession validates a raw session token and returns the session and user.
func (s *Service) ValidateSession(ctx context.Context, rawToken string) (*db.Session, *db.User, error) {
	if rawToken == "" {
		return nil, nil, fmt.Errorf("empty token")
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	session, err := s.Queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session")
	}

	// Defense-in-depth: explicit expiry check in application code
	if time.Now().After(session.ExpiresAt) {
		return nil, nil, fmt.Errorf("session expired")
	}

	user, err := s.Queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	return session, user, nil
}

func (s *Service) DestroySession(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	return s.Queries.DeleteSession(ctx, tokenHash)
}

func (s *Service) GetOAuth2Config() *oauth2.Config {
	return s.oauth2Conf
}

func (s *Service) GetUserFromGoogle(ctx context.Context, token *oauth2.Token) (*db.User, error) {
	client := s.oauth2Conf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var googleUser struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	if googleUser.Email == "" {
		return nil, fmt.Errorf("Google did not return email")
	}

	// Reject unverified email addresses
	if !googleUser.VerifiedEmail {
		return nil, fmt.Errorf("email address not verified by Google")
	}

	user, err := s.Queries.FindOrCreateUser(ctx, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		return nil, err
	}

	if err := s.Queries.FindOrCreateOAuthAccount(ctx, "google", googleUser.Sub, user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// SessionCookie creates a session cookie with the given token value.
// This is the single source of truth for session cookie configuration.
func (s *Service) SessionCookie(token string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.Config.CookieDomain != "localhost" && s.Config.CookieDomain != "",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	}
	if s.Config.CookieDomain != "" && s.Config.CookieDomain != "localhost" {
		cookie.Domain = s.Config.CookieDomain
	}
	return cookie
}

func (s *Service) ClearCookie() *http.Cookie {
	cookie := s.SessionCookie("")
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(0, 0)
	return cookie
}

func (s *Service) GetSessionFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Service) DevLogin(ctx context.Context, email, name string) (string, error) {
	if !s.Config.DevLoginEnabled {
		return "", fmt.Errorf("dev login disabled")
	}

	user, err := s.Queries.FindOrCreateUser(ctx, email, name, "")
	if err != nil {
		return "", err
	}

	token, _, err := s.GenerateSession(ctx, user.ID)
	return token, err
}

// --- OAuth State Management ---

const (
	oauthStateCookieName = "__oauth_state"
	oauthStateTTL        = 10 * time.Minute
)

// GenerateOAuthState creates a random state string and a cookie to store it.
// The returnTo URL is stored alongside the state nonce in the cookie value.
func (s *Service) GenerateOAuthState(returnTo string) (state string, cookie *http.Cookie, err error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("failed to generate OAuth state: %w", err)
	}
	nonce := hex.EncodeToString(b)

	// Cookie value stores both nonce and returnTo separated by ":"
	cookieValue := nonce + ":" + returnTo

	stateCookie := &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.Config.CookieDomain != "localhost" && s.Config.CookieDomain != "",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(oauthStateTTL.Seconds()),
	}

	return nonce, stateCookie, nil
}

// ValidateOAuthState verifies the state query param against the stored cookie.
// Returns the returnTo URL if valid.
func (s *Service) ValidateOAuthState(r *http.Request, stateParam string) (returnTo string, err error) {
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return "", fmt.Errorf("missing OAuth state cookie")
	}

	parts := strings.SplitN(cookie.Value, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed OAuth state cookie")
	}

	storedNonce := parts[0]
	storedReturnTo := parts[1]

	if storedNonce != stateParam {
		return "", fmt.Errorf("OAuth state mismatch")
	}

	return SafeReturnTo(storedReturnTo), nil
}

// ClearOAuthStateCookie returns a cookie that deletes the OAuth state cookie.
func (s *Service) ClearOAuthStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}