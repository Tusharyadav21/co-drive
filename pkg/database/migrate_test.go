package database

import "testing"

// TestDeletingAUserCascades pins the FK graph: every table that belongs to a
// user must empty out when that user is deleted.
// verifications is deliberately absent — a signup OTP is issued before the user
// row exists, so it is keyed by email and swept by expiry instead.
func TestDeletingAUserCascades(t *testing.T) {
	db := SetupTestDB(t)

	seed := []string{
		`INSERT INTO users (id, email) VALUES ('u1', 'owner@example.com');`,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ('s1', 'u1', '2099-01-01');`,
		`INSERT INTO vehicles (id, user_id, name) VALUES ('v1', 'u1', 'Daily Driver');`,
		`INSERT INTO mileage_entries (id, vehicle_id, mileage, date) VALUES ('m1', 'v1', 1000, '2026-01-01');`,
		`INSERT INTO vehicle_documents (id, vehicle_id, type, expiry_date) VALUES ('p1', 'v1', 'puc', '2099-01-01');`,
		`INSERT INTO vehicle_documents (id, vehicle_id, type, expiry_date) VALUES ('i1', 'v1', 'insurance', '2099-01-01');`,
	}
	for _, q := range seed {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("seed %q: %v", q, err)
		}
	}

	if _, err := db.Exec(`DELETE FROM users WHERE id = 'u1';`); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	for _, table := range []string{
		"sessions", "vehicles", "mileage_entries", "vehicle_documents",
	} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM ` + table + `;`).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("%s kept %d row(s) after its user was deleted", table, n)
		}
	}
}

// TestVehicleDocumentTypeIsConstrained pins the CHECK that stands in for an
// enum: without this a typo'd type value would insert silently and never be
// found by GetDocument's exact match.
func TestVehicleDocumentTypeIsConstrained(t *testing.T) {
	db := SetupTestDB(t)

	if _, err := db.Exec(`INSERT INTO users (id, email) VALUES ('u1', 'owner@example.com');`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO vehicles (id, user_id, name) VALUES ('v1', 'u1', 'Daily Driver');`); err != nil {
		t.Fatalf("seed vehicle: %v", err)
	}

	_, err := db.Exec(`INSERT INTO vehicle_documents (id, vehicle_id, type, expiry_date) VALUES ('d1', 'v1', 'road_tax', '2099-01-01');`)
	if err == nil {
		t.Fatal("expected the CHECK constraint to reject an unknown document type, got no error")
	}
}
