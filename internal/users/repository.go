package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateEmailVerified(ctx context.Context, userID string, verified bool) error
	GetProfile(ctx context.Context, userID string) (*ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*ProfileResponse, error)
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

func (r *PostgresRepository) Create(ctx context.Context, u *User) error {
	if u.FullName == "" {
		u.FullName = "Driver"
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	query := `
		INSERT INTO users (id, email, full_name, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.Email, u.FullName, u.EmailVerified, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Initialize default empty details row for the user
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO user_details (user_id, updated_at)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO NOTHING;
	`, u.ID, now)

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, full_name, email_verified, created_at, updated_at FROM users WHERE id = $1;`
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Email, &u.FullName, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return u, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, full_name, email_verified, created_at, updated_at FROM users WHERE email = $1;`
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.FullName, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return u, nil
}

func (r *PostgresRepository) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	query := `UPDATE users SET email_verified = $1, updated_at = $2 WHERE id = $3;`
	_, err := r.db.ExecContext(ctx, query, verified, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update email verified: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.email_verified,
		       COALESCE(d.phone, ''), COALESCE(d.bio, ''), COALESCE(d.address, ''),
		       COALESCE(d.emergency_contact, ''), COALESCE(d.avatar_url, ''),
		       u.created_at, u.updated_at
		FROM users u
		LEFT JOIN user_details d ON u.id = d.user_id
		WHERE u.id = $1;
	`
	p := &ProfileResponse{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.ID, &p.Email, &p.FullName, &p.EmailVerified,
		&p.Phone, &p.Bio, &p.Address,
		&p.EmergencyContact, &p.AvatarURL,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*ProfileResponse, error) {
	now := time.Now()
	if req.FullName == "" {
		req.FullName = "Driver"
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update full_name on users table
	_, err = tx.ExecContext(ctx, `UPDATE users SET full_name = $1, updated_at = $2 WHERE id = $3;`, req.FullName, now, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user name: %w", err)
	}

	// Upsert into user_details table
	upsertDetails := `
		INSERT INTO user_details (user_id, phone, bio, address, emergency_contact, avatar_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			phone = excluded.phone,
			bio = excluded.bio,
			address = excluded.address,
			emergency_contact = excluded.emergency_contact,
			avatar_url = excluded.avatar_url,
			updated_at = excluded.updated_at;
	`
	_, err = tx.ExecContext(ctx, upsertDetails, userID, req.Phone, req.Bio, req.Address, req.EmergencyContact, req.AvatarURL, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update user details: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetProfile(ctx, userID)
}
