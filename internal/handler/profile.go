package handler

import (
	"net/http"

	"co-drive/internal/db"
	"co-drive/internal/middleware"
)

func (h *Handlers) Profile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	stats, _ := h.Queries.GetUserStats(r.Context(), user.ID)
	subs, _ := h.Queries.GetPushSubscriptionsByUser(r.Context(), user.ID)

	type profileUser struct {
		*db.User
		NotificationsEnabled bool
	}

	pu := &profileUser{User: user, NotificationsEnabled: len(subs) > 0}

	data := map[string]interface{}{
		"User":  pu,
		"Stats": stats,
	}
	h.Render(w, r, "profile", data)
}

func (h *Handlers) SubscribePush(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	endpoint := r.FormValue("endpoint")
	auth := r.FormValue("auth")
	p256dh := r.FormValue("p256dh")

	if endpoint == "" || auth == "" || p256dh == "" {
		http.Error(w, "Missing subscription fields", http.StatusBadRequest)
		return
	}

	if err := h.Queries.UpsertPushSubscription(r.Context(), user.ID, endpoint, auth, p256dh); err != nil {
		http.Error(w, "Failed to save subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}