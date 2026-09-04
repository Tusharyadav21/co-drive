package auth

import (
	"net/http"

	"co-drive/internal/session"
	"co-drive/internal/users"
	"co-drive/pkg/response"
)

type Handler struct {
	userRepo       users.Repository
	sessionService *SessionService
}

func NewHandler(userRepo users.Repository, sessionService *SessionService) *Handler {
	return &Handler{
		userRepo:       userRepo,
		sessionService: sessionService,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusBadRequest, "Password registration is deprecated; please use Email OTP at /api/auth/otp/request")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusBadRequest, "Password login is deprecated; please use Email OTP at /api/auth/otp/request")
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if sess, ok := session.FromContext(r.Context()); ok {
		_ = h.sessionService.RevokeSession(r.Context(), sess.ID)
	}
	ClearSessionCookie(w)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
