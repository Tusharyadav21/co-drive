package handler

import (
	"log/slog"
	"net/http"
	"net/url"

	"co-drive/internal/auth"
	"co-drive/internal/middleware"
)

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	returnTo := auth.SafeReturnTo(r.URL.Query().Get("return_to"))

	data := map[string]interface{}{
		"Title":      "Sign in to Co-Drive",
		"ReturnTo":   returnTo,
		"DevEnabled": h.Auth.Config.DevLoginEnabled,
		"CSRFToken":  middleware.GetCSRFToken(r),
	}
	h.Render(w, r, "login", data)
}

func (h *Handlers) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	returnTo := auth.SafeReturnTo(r.URL.Query().Get("return_to"))

	// Generate state nonce and store it + returnTo in a cookie
	state, stateCookie, err := h.Auth.GenerateOAuthState(returnTo)
	if err != nil {
		slog.Error("Failed to generate OAuth state", "error", err)
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("internal_error"), http.StatusFound)
		return
	}
	http.SetCookie(w, stateCookie)

	authURL := h.Auth.GetOAuth2Config().AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handlers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	stateParam := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	errParam := r.URL.Query().Get("error")

	if errParam != "" {
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape(errParam), http.StatusFound)
		return
	}

	if stateParam == "" || code == "" {
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("invalid_request"), http.StatusFound)
		return
	}

	// Validate OAuth state against cookie (prevents login CSRF)
	returnTo, err := h.Auth.ValidateOAuthState(r, stateParam)
	if err != nil {
		slog.Error("OAuth state validation failed", "error", err)
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("invalid_state"), http.StatusFound)
		return
	}

	// Clear the state cookie immediately
	http.SetCookie(w, h.Auth.ClearOAuthStateCookie())

	if returnTo == "" {
		returnTo = "/"
	}

	token, err := h.Auth.GetOAuth2Config().Exchange(r.Context(), code)
	if err != nil {
		slog.Error("OAuth exchange failed", "error", err)
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("oauth_exchange_failed"), http.StatusFound)
		return
	}

	user, err := h.Auth.GetUserFromGoogle(r.Context(), token)
	if err != nil {
		slog.Error("Failed to get user from Google", "error", err)
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("user_info_failed"), http.StatusFound)
		return
	}

	sessionToken, _, err := h.Auth.GenerateSession(r.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("session_failed"), http.StatusFound)
		return
	}

	// Use consolidated SessionCookie helper
	http.SetCookie(w, h.Auth.SessionCookie(sessionToken))
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := h.Auth.GetSessionFromRequest(r)
	if token != "" {
		h.Auth.DestroySession(r.Context(), token)
	}
	http.SetCookie(w, h.Auth.ClearCookie())
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

func (h *Handlers) DevLogin(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.Config.DevLoginEnabled {
		http.Error(w, "Dev login disabled", http.StatusForbidden)
		return
	}

	token, err := h.Auth.DevLogin(r.Context(), "dev@codrive.app", "CoDrive Developer")
	if err != nil {
		slog.Error("Dev login failed", "error", err)
		http.Redirect(w, r, "/auth/login?error="+url.QueryEscape("dev_login_failed"), http.StatusFound)
		return
	}

	// Use consolidated SessionCookie helper
	http.SetCookie(w, h.Auth.SessionCookie(token))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}