package handler

import (
	"net/http"
	"strings"

	"co-drive/internal/middleware"
)

func (h *Handlers) ShareVehicle(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	if vehicleID == "" {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	isOwner, err := 	h.Queries.IsOwner(r.Context(), vehicleID, user.ID)
	if err != nil || !isOwner {
		h.renderError(w, "Only the vehicle owner can share", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		h.renderError(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Find user by email
	targetUser, err := 	h.Queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		h.renderError(w, "User not found", http.StatusNotFound)
		return
	}

	if targetUser.ID == user.ID {
		h.renderError(w, "Cannot share with yourself", http.StatusBadRequest)
		return
	}

	_, err = 	h.Queries.ShareVehicle(r.Context(), vehicleID, targetUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "already has access") {
			h.renderError(w, "User already has access to this vehicle", http.StatusBadRequest)
			return
		}
		h.renderError(w, "Failed to share vehicle", http.StatusInternalServerError)
		return
	}

	// Return updated shares tab
	details, _ := 	h.Queries.GetVehicleByID(r.Context(), vehicleID, user.ID)
	h.RenderPartial(w, r, "partials/shares_tab.html", map[string]interface{}{
		"Vehicle": details,
		"User":    middleware.GetUser(r.Context()),
		"SharingSuccess": "Shared successfully with " + email,
	})
}

func (h *Handlers) UnshareVehicle(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	vehicleUserID := r.PathValue("vu_id")
	if vehicleID == "" || vehicleUserID == "" {
		h.renderError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	isOwner, err := 	h.Queries.IsOwner(r.Context(), vehicleID, user.ID)
	if err != nil || !isOwner {
		h.renderError(w, "Only the vehicle owner can remove shares", http.StatusForbidden)
		return
	}

	err = 	h.Queries.RemoveVehicleShare(r.Context(), vehicleUserID, user.ID)
	if err != nil {
		h.renderError(w, "Failed to remove share", http.StatusInternalServerError)
		return
	}

	details, _ := 	h.Queries.GetVehicleByID(r.Context(), vehicleID, user.ID)
	h.RenderPartial(w, r, "partials/shares_tab.html", map[string]interface{}{
		"Vehicle": details,
		"User":    middleware.GetUser(r.Context()),
		"SharingSuccess": "Access removed successfully",
	})
}