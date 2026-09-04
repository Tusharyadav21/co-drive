package vehicles

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"co-drive/internal/session"
	"co-drive/pkg/database"
	"github.com/go-chi/chi/v5"
)

// TestSubResourceWritesRequireOwnership pins the ownership gate on the three
// writers that take a vehicle id from the URL. Without it, knowing an id was
// enough to write to a stranger's vehicle.
func TestSubResourceWritesRequireOwnership(t *testing.T) {
	db := database.SetupTestDB(t)

	if _, err := db.Exec(`INSERT INTO users (id, email) VALUES ('owner', 'owner@example.com'), ('attacker', 'attacker@example.com');`); err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	repo := NewRepository(db)
	v := &Vehicle{UserID: "owner", Name: "Daily Driver"}
	if err := repo.CreateVehicle(context.Background(), v); err != nil {
		t.Fatalf("failed to create vehicle: %v", err)
	}

	h := NewHandler(repo)
	r := chi.NewRouter()
	r.Post("/api/vehicles/{id}/mileage", h.AddMileage)
	r.Post("/api/vehicles/{id}/puc", h.UpdatePUC)
	r.Post("/api/vehicles/{id}/insurance", h.UpdateInsurance)

	cases := []struct {
		path string
		body string
	}{
		{"mileage", `{"mileage":99999}`},
		{"puc", `{"expiry_date":"2030-01-01"}`},
		{"insurance", `{"expiry_date":"2030-01-01"}`},
	}

	for _, c := range cases {
		for _, tc := range []struct {
			user string
			want int
		}{{"attacker", http.StatusNotFound}, {"owner", http.StatusOK}} {
			req := httptest.NewRequest(http.MethodPost, "/api/vehicles/"+v.ID+"/"+c.path, strings.NewReader(c.body))
			req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s1", UserID: tc.user}))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// AddMileage answers 201 on success, the upserts 200.
			if got := w.Code; got != tc.want && !(tc.user == "owner" && got == http.StatusCreated) {
				t.Errorf("%s as %s: got %d, want %d", c.path, tc.user, got, tc.want)
			}
		}
	}

	// Only the owner's write may have landed.
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM mileage_entries WHERE vehicle_id = $1;`, v.ID).Scan(&n); err != nil {
		t.Fatalf("count mileage_entries: %v", err)
	}
	if n != 1 {
		t.Errorf("got %d mileage entries on the vehicle, want 1 (the owner's)", n)
	}
}

func TestUpdateAndDeleteMileageEntry(t *testing.T) {
	db := database.SetupTestDB(t)

	if _, err := db.Exec(`INSERT INTO users (id, email) VALUES ('owner', 'owner@example.com'), ('attacker', 'attacker@example.com');`); err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	repo := NewRepository(db)
	v := &Vehicle{UserID: "owner", Name: "Daily Driver"}
	if err := repo.CreateVehicle(context.Background(), v); err != nil {
		t.Fatalf("failed to create vehicle: %v", err)
	}

	entry := &MileageEntry{VehicleID: v.ID, Mileage: 1000}
	if err := repo.AddMileageEntry(context.Background(), entry); err != nil {
		t.Fatalf("failed to add mileage: %v", err)
	}

	h := NewHandler(repo)
	r := chi.NewRouter()
	r.Put("/api/vehicles/{id}/mileage/{mileageId}", h.UpdateMileage)
	r.Delete("/api/vehicles/{id}/mileage/{mileageId}", h.DeleteMileage)

	// Attacker tries to update -> 404
	req := httptest.NewRequest(http.MethodPut, "/api/vehicles/"+v.ID+"/mileage/"+entry.ID, strings.NewReader(`{"mileage":2000,"notes":"hacked"}`))
	req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s2", UserID: "attacker"}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("attacker update got %d, want 404", w.Code)
	}

	// Owner updates -> 200
	req = httptest.NewRequest(http.MethodPut, "/api/vehicles/"+v.ID+"/mileage/"+entry.ID, strings.NewReader(`{"mileage":2500,"notes":"highway trip"}`))
	req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s1", UserID: "owner"}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("owner update got %d, want 200: %s", w.Code, w.Body.String())
	}

	// Attacker tries to delete -> 404
	req = httptest.NewRequest(http.MethodDelete, "/api/vehicles/"+v.ID+"/mileage/"+entry.ID, nil)
	req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s2", UserID: "attacker"}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("attacker delete got %d, want 404", w.Code)
	}

	// Owner deletes -> 200
	req = httptest.NewRequest(http.MethodDelete, "/api/vehicles/"+v.ID+"/mileage/"+entry.ID, nil)
	req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s1", UserID: "owner"}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("owner delete got %d, want 200: %s", w.Code, w.Body.String())
	}
}

func TestUpdateVehicleEndpoint(t *testing.T) {
	db := database.SetupTestDB(t)

	if _, err := db.Exec(`INSERT INTO users (id, email) VALUES ('owner', 'owner@example.com'), ('attacker', 'attacker@example.com');`); err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	repo := NewRepository(db)
	v := &Vehicle{UserID: "owner", Name: "Honda City", Make: "Honda", Model: "City", Year: 2022, LicensePlate: "DL01AB1234"}
	if err := repo.CreateVehicle(context.Background(), v); err != nil {
		t.Fatalf("failed to create vehicle: %v", err)
	}

	h := NewHandler(repo)
	r := chi.NewRouter()
	r.Put("/api/vehicles/{id}", h.UpdateVehicle)

	// Attacker tries to update -> 500 / error (access denied)
	req := httptest.NewRequest(http.MethodPut, "/api/vehicles/"+v.ID, strings.NewReader(`{"name":"Hacked Vehicle","make":"Fake"}`))
	req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s2", UserID: "attacker"}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Errorf("attacker vehicle update expected error, got 200 OK")
	}

	// Owner updates -> 200 OK
	req = httptest.NewRequest(http.MethodPut, "/api/vehicles/"+v.ID, strings.NewReader(`{"name":"City Pro","make":"Honda","model":"City ZX","year":2024,"license_plate":"DL01XX9999"}`))
	req = req.WithContext(session.WithSession(req.Context(), &session.Session{ID: "s1", UserID: "owner"}))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("owner vehicle update got %d, want 200: %s", w.Code, w.Body.String())
	}

	updated, err := repo.GetVehicleByID(context.Background(), v.ID, "owner")
	if err != nil || updated == nil {
		t.Fatalf("failed to get updated vehicle: %v", err)
	}
	if updated.Name != "City Pro" || updated.LicensePlate != "DL01XX9999" {
		t.Errorf("got name=%q plate=%q, want 'City Pro', 'DL01XX9999'", updated.Name, updated.LicensePlate)
	}
}
