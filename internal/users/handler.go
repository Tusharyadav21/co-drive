package users

import (
	"net/http"

	"co-drive/internal/session"
	"co-drive/pkg/response"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// GetProfile retrieves the consolidated profile (identity + extended details) of the authenticated user.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	profile, err := h.repo.GetProfile(r.Context(), sess.UserID)
	if err != nil || profile == nil {
		response.Error(w, http.StatusNotFound, "User profile not found")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

// UpdateProfile modifies the user's display name and extended profile fields.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Require(w, r)
	if !ok {
		return
	}

	var req UpdateProfileRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid profile payload")
		return
	}

	profile, err := h.repo.UpdateProfile(r.Context(), sess.UserID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}
