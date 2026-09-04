package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type VerificationRepository interface {
	Create(ctx context.Context, v *Verification) error
	GetLatestValid(ctx context.Context, email string, vType VerificationType) (*Verification, error)
	MarkAsUsed(ctx context.Context, id string) error
}

type PostgresVerificationRepository struct {
	db *sql.DB
}

func NewVerificationRepository(db *sql.DB) *PostgresVerificationRepository {
	return &PostgresVerificationRepository{db: db}
}

func NewPostgresVerificationRepository(db *sql.DB) *PostgresVerificationRepository {
	return &PostgresVerificationRepository{db: db}
}

func (r *PostgresVerificationRepository) Create(ctx context.Context, v *Verification) error {
	query := `
		INSERT INTO verifications (id, email, type, code_hash, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	v.CreatedAt = time.Now()

	// Sweep spent rows on write
	_, _ = r.db.ExecContext(ctx, `DELETE FROM verifications WHERE expires_at < $1;`, v.CreatedAt)

	_, err := r.db.ExecContext(ctx, query, v.ID, v.Email, v.Type, v.CodeHash, v.ExpiresAt, v.Used, v.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert verification token: %w", err)
	}
	return nil
}

func (r *PostgresVerificationRepository) GetLatestValid(ctx context.Context, email string, vType VerificationType) (*Verification, error) {
	query := `
		SELECT id, email, type, code_hash, expires_at, used, created_at
		FROM verifications
		WHERE email = $1 AND type = $2 AND used = FALSE AND expires_at > $3
		ORDER BY created_at DESC LIMIT 1;
	`
	v := &Verification{}
	err := r.db.QueryRowContext(ctx, query, email, vType, time.Now()).Scan(
		&v.ID, &v.Email, &v.Type, &v.CodeHash, &v.ExpiresAt, &v.Used, &v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query verification: %w", err)
	}
	return v, nil
}

func (r *PostgresVerificationRepository) MarkAsUsed(ctx context.Context, id string) error {
	query := `UPDATE verifications SET used = TRUE WHERE id = $1;`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark verification used: %w", err)
	}
	return nil
}
