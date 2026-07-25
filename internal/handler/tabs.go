package handler

import (
	"net/http"
	"time"

	"co-drive/internal/middleware"
)

func (h *Handlers) VehicleTab(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vehicleID := r.PathValue("id")
	tab := r.PathValue("tab")

	if vehicleID == "" || tab == "" {
		h.renderError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	details, err := h.Queries.GetVehicleByID(r.Context(), vehicleID, user.ID)
	if err != nil {
		h.renderError(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	enriched := enrichVehicleData(details)

	data := map[string]interface{}{
		"Vehicle":   enriched,
		"User":      user,
		"ActiveTab": tab,
		"Now":       time.Now(),
	}

	switch tab {
	case "mileage":
		h.RenderPartial(w, r, "partials/mileage_tab.html", data)
	case "documents":
		h.RenderPartial(w, r, "partials/documents_tab.html", data)
	case "shares":
		h.RenderPartial(w, r, "partials/shares_tab.html", data)
	default:
		h.renderError(w, "Invalid tab", http.StatusBadRequest)
	}
}