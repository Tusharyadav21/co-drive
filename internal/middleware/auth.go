package middleware

import (
	"context"
	"net/http"
	"net/url"

	"co-drive/internal/auth"
	"co-drive/internal/db"
)

// AuthMiddleware extracts session from cookie and adds user to context
func AuthMiddleware(authSvc *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := authSvc.GetSessionFromRequest(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		session, user, err := authSvc.ValidateSession(r.Context(), token)
		if err != nil {
			// Clear invalid cookie
			http.SetCookie(w, authSvc.ClearCookie())
			next.ServeHTTP(w, r)
			return
		}

		// Add user and session to context
		ctx := context.WithValue(r.Context(), auth.UserContextKey, user)
		ctx = context.WithValue(ctx, auth.SessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth redirects to login if not authenticated
func RequireAuth(authSvc *auth.Service, loginPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			if user == nil {
				// Redirect with return URL
				returnTo := r.URL.Path
				if r.URL.RawQuery != "" {
					returnTo += "?" + r.URL.RawQuery
				}
				http.Redirect(w, r, loginPath+"?return_to="+url.QueryEscape(returnTo), http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUser retrieves user from context
func GetUser(ctx context.Context) *db.User {
	if u, ok := ctx.Value(auth.UserContextKey).(*db.User); ok {
		return u
	}
	return nil
}

// GetSession retrieves session from context
func GetSession(ctx context.Context) *db.Session {
	if s, ok := ctx.Value(auth.SessionContextKey).(*db.Session); ok {
		return s
	}
	return nil
}