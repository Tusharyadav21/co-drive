package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	Extend(ctx context.Context, id string, newExpiry time.Time) error
	Delete(ctx context.Context, id string) error
}

type PostgresSessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func NewPostgresSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Create(ctx context.Context, session *Session) error {
	// Sweep expired sessions on write
	_, _ = r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < $1;`, time.Now())

	query := `
		INSERT INTO sessions (id, user_id, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	session.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, session.ID, session.UserID, session.UserAgent, session.IPAddress, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepository) GetByID(ctx context.Context, id string) (*Session, error) {
	query := `SELECT id, user_id, user_agent, ip_address, expires_at, created_at FROM sessions WHERE id = $1 AND expires_at > $2;`
	s := &Session{}
	err := r.db.QueryRowContext(ctx, query, id, time.Now()).Scan(&s.ID, &s.UserID, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}
	return s, nil
}

func (r *PostgresSessionRepository) Extend(ctx context.Context, id string, newExpiry time.Time) error {
	query := `UPDATE sessions SET expires_at = $1 WHERE id = $2 AND expires_at > $3;`
	res, err := r.db.ExecContext(ctx, query, newExpiry, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("session not found or expired")
	}
	return nil
}

func (r *PostgresSessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1;`, id)
	return err
}
