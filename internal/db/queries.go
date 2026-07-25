package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID            string
	Name          string
	Email         string
	EmailVerified bool
	Image         pgtype.Text
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserStats struct {
	TotalVehicles int
	TotalRefills  int
	TotalSpent    float64
}

type Vehicle struct {
	ID            string
	Name          string
	Make          string
	Model         string
	Year          int
	LicensePlate  string
	OwnerID       string
	Role          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type VehicleWithDetails struct {
	Vehicle
	Users          []VehicleUser
	MileageEntries []MileageEntry
	PUC            *PUC
	Insurance      *Insurance
}

type VehicleUser struct {
	ID        string
	UserID    string
	VehicleID string
	Role      string
	CreatedAt time.Time
	User      *User
}

type MileageEntry struct {
	ID          string
	VehicleID   string
	UserID      string
	Mileage     int
	Date        time.Time
	Notes       pgtype.Text
	FuelLitres  pgtype.Float8
	FuelAmount  pgtype.Float8
	Efficiency  float64
	CreatedAt   time.Time
}

type PUC struct {
	VehicleID         string
	ExpiryDate        time.Time
	CertificateNumber pgtype.Text
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Insurance struct {
	VehicleID    string
	ExpiryDate   time.Time
	PolicyNumber pgtype.Text
	Provider     pgtype.Text
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PushSubscription struct {
	ID        string
	UserID    string
	Endpoint  string
	Auth      string
	P256dh    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Session struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type OAuthAccount struct {
	ID                string
	UserID            string
	Provider          string
	ProviderAccountID string
	AccessToken       pgtype.Text
	RefreshToken      pgtype.Text
	ExpiresAt         pgtype.Timestamptz
}

// User queries
func (q *Queries) FindOrCreateUser(ctx context.Context, email, name, image string) (*User, error) {
	// Try to find existing user
	user, err := q.GetUserByEmail(ctx, email)
	if err == nil {
		return user, nil
	}

	// Create new user
	return q.CreateUser(ctx, email, name, image)
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, name, email, email_verified, image, created_at, updated_at
		FROM users WHERE email = $1
	`
	row := q.db.QueryRow(ctx, query, email)
	return scanUser(row)
}

func (q *Queries) GetUserByID(ctx context.Context, id string) (*User, error) {
	const query = `
		SELECT id, name, email, email_verified, image, created_at, updated_at
		FROM users WHERE id = $1
	`
	row := q.db.QueryRow(ctx, query, id)
	return scanUser(row)
}

func (q *Queries) CreateUser(ctx context.Context, email, name, image string) (*User, error) {
	const query = `
		INSERT INTO users (email, name, image, email_verified)
		VALUES ($1, $2, $3, true)
		RETURNING id, name, email, email_verified, image, created_at, updated_at
	`
	row := q.db.QueryRow(ctx, query, email, name, image)
	return scanUser(row)
}

func (q *Queries) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	const query = `
		SELECT 
			(SELECT COUNT(*) FROM vehicles WHERE owner_id = $1 OR id IN (SELECT vehicle_id FROM vehicle_users WHERE user_id = $1)) AS total_vehicles,
			(SELECT COUNT(*) FROM mileage_entries WHERE user_id = $1 AND (fuel_litres IS NOT NULL OR fuel_amount IS NOT NULL)) AS total_refills,
			(SELECT COALESCE(SUM(fuel_amount), 0) FROM mileage_entries WHERE user_id = $1) AS total_spent
	`
	row := q.db.QueryRow(ctx, query, userID)
	var stats UserStats
	err := row.Scan(&stats.TotalVehicles, &stats.TotalRefills, &stats.TotalSpent)
	return &stats, err
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.Image, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Vehicle queries
func (q *Queries) GetVehiclesByUserID(ctx context.Context, userID string) ([]Vehicle, error) {
	const query = `
		SELECT v.id, v.name, v.make, v.model, v.year, v.license_plate, v.owner_id, COALESCE(vu.role, 'owner') AS role, v.created_at, v.updated_at
		FROM vehicles v
		LEFT JOIN vehicle_users vu ON v.id = vu.vehicle_id AND vu.user_id = $1
		WHERE v.owner_id = $1 OR vu.user_id = $1
		ORDER BY v.created_at DESC
	`
	rows, err := q.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		var role pgtype.Text
		if err := rows.Scan(&v.ID, &v.Name, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.OwnerID, &role, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		if role.Valid {
			v.Role = role.String
		} else {
			v.Role = "owner"
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, rows.Err()
}

func (q *Queries) GetVehicleByID(ctx context.Context, vehicleID, userID string) (*VehicleWithDetails, error) {
	const vehicleQuery = `
		SELECT v.id, v.name, v.make, v.model, v.year, v.license_plate, v.owner_id, vu.role, v.created_at, v.updated_at
		FROM vehicles v
		LEFT JOIN vehicle_users vu ON v.id = vu.vehicle_id AND vu.user_id = $2
		WHERE v.id = $1 AND (v.owner_id = $2 OR vu.user_id = $2)
	`
	row := q.db.QueryRow(ctx, vehicleQuery, vehicleID, userID)

	var v VehicleWithDetails
	var role pgtype.Text
	err := row.Scan(&v.ID, &v.Name, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.OwnerID, &role, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if role.Valid {
		v.Role = role.String
	} else if v.OwnerID == userID {
		v.Role = "owner"
	} else {
		v.Role = "viewer"
	}

	// Fetch related data
	v.Users, _ = q.GetVehicleUsers(ctx, vehicleID)
	v.MileageEntries, _ = q.GetMileageEntries(ctx, vehicleID)
	v.PUC, _ = q.GetPUC(ctx, vehicleID)
	v.Insurance, _ = q.GetInsurance(ctx, vehicleID)

	return &v, nil
}

func (q *Queries) CreateVehicle(ctx context.Context, name, make, model string, year int, licensePlate, ownerID string) (*Vehicle, error) {
	const query = `
		INSERT INTO vehicles (name, make, model, year, license_plate, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, make, model, year, license_plate, owner_id, 'owner' as role, created_at, updated_at
	`
	row := q.db.QueryRow(ctx, query, name, make, model, year, licensePlate, ownerID)
	return scanVehicle(row)
}

func (q *Queries) GetVehicleUsers(ctx context.Context, vehicleID string) ([]VehicleUser, error) {
	const query = `
		SELECT vu.id, vu.user_id, vu.vehicle_id, vu.role, vu.created_at,
		       u.id, u.name, u.email, u.email_verified, u.image, u.created_at, u.updated_at
		FROM vehicle_users vu
		JOIN users u ON vu.user_id = u.id
		WHERE vu.vehicle_id = $1
		ORDER BY vu.created_at
	`
	rows, err := q.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []VehicleUser
	for rows.Next() {
		var vu VehicleUser
		var u User
		if err := rows.Scan(&vu.ID, &vu.UserID, &vu.VehicleID, &vu.Role, &vu.CreatedAt,
			&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.Image, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		vu.User = &u
		users = append(users, vu)
	}
	return users, rows.Err()
}

func (q *Queries) ShareVehicle(ctx context.Context, vehicleID, targetUserID string) (*VehicleUser, error) {
	const query = `
		INSERT INTO vehicle_users (user_id, vehicle_id, role)
		VALUES ($1, $2, 'viewer')
		ON CONFLICT (user_id, vehicle_id) DO NOTHING
		RETURNING id, user_id, vehicle_id, role, created_at
	`
	row := q.db.QueryRow(ctx, query, targetUserID, vehicleID)
	var vu VehicleUser
	err := row.Scan(&vu.ID, &vu.UserID, &vu.VehicleID, &vu.Role, &vu.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user already has access to this vehicle")
	}
	return &vu, err
}

func (q *Queries) RemoveVehicleShare(ctx context.Context, vehicleUserID, requestingUserID string) error {
	const query = `
		DELETE FROM vehicle_users
		WHERE id = $1 AND (vehicle_id IN (
			SELECT id FROM vehicles WHERE owner_id = $2
		) OR user_id = $2)
	`
	_, err := q.db.Exec(ctx, query, vehicleUserID, requestingUserID)
	return err
}

func (q *Queries) HasAccess(ctx context.Context, vehicleID, userID string) (bool, string, error) {
	const query = `
		SELECT v.owner_id, vu.role
		FROM vehicles v
		LEFT JOIN vehicle_users vu ON v.id = vu.vehicle_id AND vu.user_id = $2
		WHERE v.id = $1 AND (v.owner_id = $2 OR vu.user_id = $2)
	`
	row := q.db.QueryRow(ctx, query, vehicleID, userID)
	var ownerID, role pgtype.Text
	err := row.Scan(&ownerID, &role)
	if err == pgx.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, role.String, nil
}

func (q *Queries) IsOwner(ctx context.Context, vehicleID, userID string) (bool, error) {
	const query = `SELECT 1 FROM vehicles WHERE id = $1 AND owner_id = $2`
	var exists int
	err := q.db.QueryRow(ctx, query, vehicleID, userID).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func scanVehicle(row pgx.Row) (*Vehicle, error) {
	var v Vehicle
	err := row.Scan(&v.ID, &v.Name, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.OwnerID, &v.Role, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Mileage queries
func (q *Queries) AddMileageEntry(ctx context.Context, vehicleID, userID string, mileage int, date time.Time, notes string, fuelLitres, fuelAmount *float64) (*MileageEntry, error) {
	const query = `
		INSERT INTO mileage_entries (vehicle_id, user_id, mileage, date, notes, fuel_litres, fuel_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, vehicle_id, user_id, mileage, date, notes, fuel_litres, fuel_amount, created_at
	`
	row := q.db.QueryRow(ctx, query, vehicleID, userID, mileage, date, notes, fuelLitres, fuelAmount)
	return scanMileageEntry(row)
}

func (q *Queries) GetMileageEntries(ctx context.Context, vehicleID string) ([]MileageEntry, error) {
	const query = `
		SELECT id, vehicle_id, user_id, mileage, date, notes, fuel_litres, fuel_amount, created_at
		FROM mileage_entries
		WHERE vehicle_id = $1
		ORDER BY date DESC
	`
	rows, err := q.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MileageEntry
	for rows.Next() {
		e, err := scanMileageEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *e)
	}
	return entries, rows.Err()
}

func scanMileageEntry(scanner interface {
	Scan(dest ...interface{}) error
}) (*MileageEntry, error) {
	var e MileageEntry
	err := scanner.Scan(&e.ID, &e.VehicleID, &e.UserID, &e.Mileage, &e.Date, &e.Notes, &e.FuelLitres, &e.FuelAmount, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Document queries (PUC, Insurance)
func (q *Queries) UpsertPUC(ctx context.Context, vehicleID string, expiryDate time.Time, certificateNumber string) (*PUC, error) {
	const query = `
		INSERT INTO puc (vehicle_id, expiry_date, certificate_number)
		VALUES ($1, $2, $3)
		ON CONFLICT (vehicle_id) DO UPDATE SET
			expiry_date = EXCLUDED.expiry_date,
			certificate_number = EXCLUDED.certificate_number,
			updated_at = now()
		RETURNING vehicle_id, expiry_date, certificate_number, created_at, updated_at
	`
	row := q.db.QueryRow(ctx, query, vehicleID, expiryDate, certificateNumber)
	return scanPUC(row)
}

func (q *Queries) UpsertInsurance(ctx context.Context, vehicleID string, expiryDate time.Time, policyNumber, provider string) (*Insurance, error) {
	const query = `
		INSERT INTO insurance (vehicle_id, expiry_date, policy_number, provider)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (vehicle_id) DO UPDATE SET
			expiry_date = EXCLUDED.expiry_date,
			policy_number = EXCLUDED.policy_number,
			provider = EXCLUDED.provider,
			updated_at = now()
		RETURNING vehicle_id, expiry_date, policy_number, provider, created_at, updated_at
	`
	row := q.db.QueryRow(ctx, query, vehicleID, expiryDate, policyNumber, provider)
	return scanInsurance(row)
}

func (q *Queries) GetPUC(ctx context.Context, vehicleID string) (*PUC, error) {
	const query = `SELECT vehicle_id, expiry_date, certificate_number, created_at, updated_at FROM puc WHERE vehicle_id = $1`
	row := q.db.QueryRow(ctx, query, vehicleID)
	return scanPUC(row)
}

func (q *Queries) GetInsurance(ctx context.Context, vehicleID string) (*Insurance, error) {
	const query = `SELECT vehicle_id, expiry_date, policy_number, provider, created_at, updated_at FROM insurance WHERE vehicle_id = $1`
	row := q.db.QueryRow(ctx, query, vehicleID)
	return scanInsurance(row)
}

func scanPUC(row pgx.Row) (*PUC, error) {
	var p PUC
	err := row.Scan(&p.VehicleID, &p.ExpiryDate, &p.CertificateNumber, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func scanInsurance(row pgx.Row) (*Insurance, error) {
	var i Insurance
	err := row.Scan(&i.VehicleID, &i.ExpiryDate, &i.PolicyNumber, &i.Provider, &i.CreatedAt, &i.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

// Push subscription queries
func (q *Queries) UpsertPushSubscription(ctx context.Context, userID, endpoint, auth, p256dh string) error {
	const query = `
		INSERT INTO push_subscriptions (user_id, endpoint, auth, p256dh)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (endpoint) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			auth = EXCLUDED.auth,
			p256dh = EXCLUDED.p256dh,
			updated_at = now()
	`
	_, err := q.db.Exec(ctx, query, userID, endpoint, auth, p256dh)
	return err
}

func (q *Queries) GetPushSubscriptionsByUser(ctx context.Context, userID string) ([]PushSubscription, error) {
	const query = `SELECT id, user_id, endpoint, auth, p256dh, created_at, updated_at FROM push_subscriptions WHERE user_id = $1`
	rows, err := q.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []PushSubscription
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.Auth, &s.P256dh, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func (q *Queries) GetExpiringDocuments(ctx context.Context, pucDays, insuranceDays int) ([]ExpiringDocument, error) {
	const query = `
		SELECT 
			v.id as vehicle_id, v.name as vehicle_name, v.owner_id,
			u.name as owner_name,
			ps.endpoint, ps.auth, ps.p256dh,
			'puc' as type, p.expiry_date,
			p.expiry_date - CURRENT_DATE as days_left
		FROM vehicles v
		JOIN users u ON v.owner_id = u.id
		LEFT JOIN push_subscriptions ps ON u.id = ps.user_id
		JOIN puc p ON v.id = p.vehicle_id
		WHERE p.expiry_date BETWEEN CURRENT_DATE AND CURRENT_DATE + $1::int

		UNION ALL

		SELECT 
			v.id as vehicle_id, v.name as vehicle_name, v.owner_id,
			u.name as owner_name,
			ps.endpoint, ps.auth, ps.p256dh,
			'insurance' as type, i.expiry_date,
			i.expiry_date - CURRENT_DATE as days_left
		FROM vehicles v
		JOIN users u ON v.owner_id = u.id
		LEFT JOIN push_subscriptions ps ON u.id = ps.user_id
		JOIN insurance i ON v.id = i.vehicle_id
		WHERE i.expiry_date BETWEEN CURRENT_DATE AND CURRENT_DATE + $2::int
	`
	rows, err := q.db.Query(ctx, query, pucDays, insuranceDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []ExpiringDocument
	for rows.Next() {
		var d ExpiringDocument
		if err := rows.Scan(&d.VehicleID, &d.VehicleName, &d.OwnerID, &d.OwnerName,
			&d.Endpoint, &d.Auth, &d.P256dh,
			&d.Type, &d.ExpiryDate, &d.DaysLeft); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

type ExpiringDocument struct {
	VehicleID   string
	VehicleName string
	OwnerID     string
	OwnerName   string
	Endpoint    pgtype.Text
	Auth        pgtype.Text
	P256dh      pgtype.Text
	Type        string
	ExpiryDate  time.Time
	DaysLeft    int
}

// Session queries
func (q *Queries) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*Session, error) {
	const query = `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, created_at
	`
	row := q.db.QueryRow(ctx, query, userID, tokenHash, expiresAt)
	return scanSession(row)
}

func (q *Queries) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	const query = `SELECT id, user_id, token_hash, expires_at, created_at FROM sessions WHERE token_hash = $1 AND expires_at > now()`
	row := q.db.QueryRow(ctx, query, tokenHash)
	return scanSession(row)
}

func (q *Queries) DeleteSession(ctx context.Context, tokenHash string) error {
	const query = `DELETE FROM sessions WHERE token_hash = $1`
	_, err := q.db.Exec(ctx, query, tokenHash)
	return err
}

func (q *Queries) DeleteExpiredSessions(ctx context.Context) error {
	const query = `DELETE FROM sessions WHERE expires_at <= now()`
	_, err := q.db.Exec(ctx, query)
	return err
}

func (q *Queries) DeleteUserSessions(ctx context.Context, userID string) error {
	const query = `DELETE FROM sessions WHERE user_id = $1`
	_, err := q.db.Exec(ctx, query, userID)
	return err
}

func scanSession(row pgx.Row) (*Session, error) {
	var s Session
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.CreatedAt)
	return &s, err
}

// OAuth queries
func (q *Queries) FindOrCreateOAuthAccount(ctx context.Context, provider, providerAccountID string, userID string) error {
	const query = `
		INSERT INTO oauth_accounts (user_id, provider, provider_account_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_account_id) DO UPDATE SET
			user_id = EXCLUDED.user_id
	`
	_, err := q.db.Exec(ctx, query, userID, provider, providerAccountID)
	return err
}

// Queries holds the database pool
type Queries struct {
	db *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{db: pool}
}