package vehicles

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateVehicle(ctx context.Context, v *Vehicle) error
	GetVehicleByID(ctx context.Context, id string, userID string) (*Vehicle, error)
	ListVehiclesByUserID(ctx context.Context, userID string) ([]Vehicle, error)
	UpdateVehicle(ctx context.Context, v *Vehicle) error
	DeleteVehicle(ctx context.Context, id string, userID string) error

	AddMileageEntry(ctx context.Context, entry *MileageEntry) error
	UpdateMileageEntry(ctx context.Context, entry *MileageEntry) error
	DeleteMileageEntry(ctx context.Context, id string, vehicleID string) error
	GetMileageEntries(ctx context.Context, vehicleID string) ([]MileageEntry, error)

	UpsertDocument(ctx context.Context, doc *VehicleDocument) error
	GetDocument(ctx context.Context, vehicleID string, docType DocumentType) (*VehicleDocument, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateVehicle(ctx context.Context, v *Vehicle) error {
	v.ID = uuid.NewString()
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()

	query := `
		INSERT INTO vehicles (id, user_id, name, make, model, year, license_plate, chassis_number, engine_number, registration_date, next_service_km, next_service_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);
	`
	_, err := r.db.ExecContext(ctx, query, v.ID, v.UserID, v.Name, v.Make, v.Model, v.Year, v.LicensePlate, v.ChassisNumber, v.EngineNumber, v.RegistrationDate, v.NextServiceKM, v.NextServiceDate, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create vehicle: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetVehicleByID(ctx context.Context, id string, userID string) (*Vehicle, error) {
	query := `SELECT id, user_id, name, make, model, year, license_plate, chassis_number, engine_number, registration_date, next_service_km, next_service_date, created_at, updated_at FROM vehicles WHERE id = $1 AND user_id = $2;`
	row := r.db.QueryRowContext(ctx, query, id, userID)

	var v Vehicle
	if err := row.Scan(&v.ID, &v.UserID, &v.Name, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.ChassisNumber, &v.EngineNumber, &v.RegistrationDate, &v.NextServiceKM, &v.NextServiceDate, &v.CreatedAt, &v.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	v.PUC, _ = r.GetDocument(ctx, v.ID, DocumentPUC)
	v.Insurance, _ = r.GetDocument(ctx, v.ID, DocumentInsurance)

	v.Mileage, _ = r.GetMileageEntries(ctx, v.ID)

	return &v, nil
}

func (r *PostgresRepository) ListVehiclesByUserID(ctx context.Context, userID string) ([]Vehicle, error) {
	query := `SELECT id, user_id, name, make, model, year, license_plate, chassis_number, engine_number, registration_date, next_service_km, next_service_date, created_at, updated_at FROM vehicles WHERE user_id = $1 ORDER BY created_at DESC;`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicles: %w", err)
	}

	list := []Vehicle{}
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.ChassisNumber, &v.EngineNumber, &v.RegistrationDate, &v.NextServiceKM, &v.NextServiceDate, &v.CreatedAt, &v.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range list {
		list[i].PUC, _ = r.GetDocument(ctx, list[i].ID, DocumentPUC)
		list[i].Insurance, _ = r.GetDocument(ctx, list[i].ID, DocumentInsurance)
		list[i].Mileage, _ = r.GetMileageEntries(ctx, list[i].ID)
	}

	return list, nil
}

func (r *PostgresRepository) UpdateVehicle(ctx context.Context, v *Vehicle) error {
	v.UpdatedAt = time.Now()
	query := `
		UPDATE vehicles
		SET name = $1, make = $2, model = $3, year = $4, license_plate = $5, chassis_number = $6, engine_number = $7, registration_date = $8, next_service_km = $9, next_service_date = $10, updated_at = $11
		WHERE id = $12 AND user_id = $13;
	`
	res, err := r.db.ExecContext(ctx, query, v.Name, v.Make, v.Model, v.Year, v.LicensePlate, v.ChassisNumber, v.EngineNumber, v.RegistrationDate, v.NextServiceKM, v.NextServiceDate, v.UpdatedAt, v.ID, v.UserID)
	if err != nil {
		return fmt.Errorf("failed to update vehicle: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("vehicle not found or access denied")
	}
	return nil
}

func (r *PostgresRepository) DeleteVehicle(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM vehicles WHERE id = $1 AND user_id = $2;`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *PostgresRepository) AddMileageEntry(ctx context.Context, entry *MileageEntry) error {
	entry.ID = uuid.NewString()
	entry.CreatedAt = time.Now()

	query := `
		INSERT INTO mileage_entries (id, vehicle_id, mileage, date, notes, fuel_litres, fuel_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err := r.db.ExecContext(ctx, query, entry.ID, entry.VehicleID, entry.Mileage, entry.Date, entry.Notes, entry.FuelLitres, entry.FuelAmount, entry.CreatedAt)
	return err
}

func (r *PostgresRepository) GetMileageEntries(ctx context.Context, vehicleID string) ([]MileageEntry, error) {
	query := `SELECT id, vehicle_id, mileage, date, notes, fuel_litres, fuel_amount, created_at FROM mileage_entries WHERE vehicle_id = $1 ORDER BY date ASC;`
	rows, err := r.db.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MileageEntry
	for rows.Next() {
		var e MileageEntry
		if err := rows.Scan(&e.ID, &e.VehicleID, &e.Mileage, &e.Date, &e.Notes, &e.FuelLitres, &e.FuelAmount, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func (r *PostgresRepository) UpdateMileageEntry(ctx context.Context, entry *MileageEntry) error {
	query := `
		UPDATE mileage_entries
		SET mileage = $1, date = $2, notes = $3, fuel_litres = $4, fuel_amount = $5
		WHERE id = $6 AND vehicle_id = $7;
	`
	res, err := r.db.ExecContext(ctx, query, entry.Mileage, entry.Date, entry.Notes, entry.FuelLitres, entry.FuelAmount, entry.ID, entry.VehicleID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresRepository) DeleteMileageEntry(ctx context.Context, id string, vehicleID string) error {
	query := `DELETE FROM mileage_entries WHERE id = $1 AND vehicle_id = $2;`
	res, err := r.db.ExecContext(ctx, query, id, vehicleID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresRepository) UpsertDocument(ctx context.Context, doc *VehicleDocument) error {
	doc.UpdatedAt = time.Now()
	if doc.ID == "" {
		doc.ID = uuid.NewString()
	}

	query := `
		INSERT INTO vehicle_documents (id, vehicle_id, type, reference, issuer, expiry_date, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (vehicle_id, type) DO UPDATE SET
			reference = excluded.reference,
			issuer = excluded.issuer,
			expiry_date = excluded.expiry_date,
			updated_at = excluded.updated_at;
	`
	_, err := r.db.ExecContext(ctx, query, doc.ID, doc.VehicleID, doc.Type, doc.Reference, doc.Issuer, doc.ExpiryDate, doc.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetDocument(ctx context.Context, vehicleID string, docType DocumentType) (*VehicleDocument, error) {
	query := `SELECT id, vehicle_id, type, reference, issuer, expiry_date, updated_at FROM vehicle_documents WHERE vehicle_id = $1 AND type = $2;`
	var d VehicleDocument
	if err := r.db.QueryRowContext(ctx, query, vehicleID, docType).Scan(&d.ID, &d.VehicleID, &d.Type, &d.Reference, &d.Issuer, &d.ExpiryDate, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}
