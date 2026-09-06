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

func TestVehicleRTOAndServicePersistence(t *testing.T) {
	db := database.SetupTestDB(t)

	if _, err := db.Exec(`INSERT INTO users (id, email) VALUES ('u1', 'owner@example.com');`); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	repo := NewRepository(db)
	ctx := context.Background()

	v := &Vehicle{
		UserID:           "u1",
		Name:             "Creta",
		Make:             "Hyundai",
		Model:            "SX(O)",
		Year:             2024,
		LicensePlate:     "MH 02 CD 5678",
		ChassisNumber:    "MALC12345678",
		EngineNumber:     "ENG987654",
		RegistrationDate: "2024-01-15",
		NextServiceKM:    10000,
		NextServiceDate:  "2025-01-15",
	}

	if err := repo.CreateVehicle(ctx, v); err != nil {
		t.Fatalf("CreateVehicle failed: %v", err)
	}

	fetched, err := repo.GetVehicleByID(ctx, v.ID, "u1")
	if err != nil {
		t.Fatalf("GetVehicleByID failed: %v", err)
	}

	if fetched.ChassisNumber != "MALC12345678" || fetched.EngineNumber != "ENG987654" || fetched.RegistrationDate != "2024-01-15" || fetched.NextServiceKM != 10000 || fetched.NextServiceDate != "2025-01-15" {
		t.Errorf("RTO/Service fields mismatch on create: %+v", fetched)
	}

	fetched.NextServiceKM = 15000
	fetched.ChassisNumber = "MALC99999999"
	if err := repo.UpdateVehicle(ctx, fetched); err != nil {
		t.Fatalf("UpdateVehicle failed: %v", err)
	}

	updated, err := repo.GetVehicleByID(ctx, v.ID, "u1")
	if err != nil {
		t.Fatalf("GetVehicleByID after update failed: %v", err)
	}
	if updated.NextServiceKM != 15000 || updated.ChassisNumber != "MALC99999999" {
		t.Errorf("RTO/Service fields mismatch after update: %+v", updated)
	}
}
