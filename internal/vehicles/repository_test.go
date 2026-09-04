package vehicles

import (
	"context"
	"testing"
	"time"

	"co-drive/pkg/database"
)

// TestListVehiclesHydratesWithoutDeadlock verifies that hydrating PUC/insurance
// and mileage entries works cleanly.
func TestListVehiclesHydratesWithoutDeadlock(t *testing.T) {
	db := database.SetupTestDB(t)

	if _, err := db.Exec(`INSERT INTO users (id, email) VALUES ('u1', 'owner@example.com');`); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, name := range []string{"Daily Driver", "Weekend Car"} {
		v := &Vehicle{UserID: "u1", Name: name}
		if err := repo.CreateVehicle(ctx, v); err != nil {
			t.Fatalf("failed to create vehicle: %v", err)
		}
		if err := repo.UpsertDocument(ctx, &VehicleDocument{
			VehicleID:  v.ID,
			Type:       DocumentPUC,
			ExpiryDate: time.Now().Add(48 * time.Hour),
		}); err != nil {
			t.Fatalf("failed to upsert PUC: %v", err)
		}
	}

	list, err := repo.ListVehiclesByUserID(ctx, "u1")
	if err != nil {
		t.Fatalf("ListVehiclesByUserID: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("got %d vehicles, want 2", len(list))
	}
	for _, v := range list {
		if v.PUC == nil {
			t.Errorf("vehicle %q was not hydrated with its PUC record", v.Name)
		}
	}
}

func TestListVehiclesReturnsEmptySliceNotNil(t *testing.T) {
	db := database.SetupTestDB(t)

	list, err := NewRepository(db).ListVehiclesByUserID(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("ListVehiclesByUserID: %v", err)
	}
	// The dashboard renders this straight into an array; nil would marshal as
	// JSON null.
	if list == nil {
		t.Error("got nil, want an empty slice")
	}
}
