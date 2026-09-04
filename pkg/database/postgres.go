package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// New initializes and validates a connection pool to a PostgreSQL database.
func New(databaseURL string) (*sql.DB, error) {
	return NewPostgres(databaseURL)
}

// NewPostgres initializes and validates a connection pool to a PostgreSQL database
// (e.g. Neon, Supabase, local PostgreSQL, RDS) via database/sql.
func NewPostgres(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgresql database: %w", err)
	}

	// Production-grade connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgresql database: %w", err)
	}

	return db, nil
}
