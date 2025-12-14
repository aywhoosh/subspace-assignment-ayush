package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
}

type OpenOptions struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Open(ctx context.Context, opts OpenOptions) (*DB, error) {
	if opts.Path == "" {
		return nil, errors.New("storage: sqlite path is required")
	}

	// modernc sqlite DSN: https://pkg.go.dev/modernc.org/sqlite
	db, err := sql.Open("sqlite", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("storage: open sqlite: %w", err)
	}

	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: ping: %w", err)
	}

	return &DB{sql: db}, nil
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) SQL() *sql.DB {
	return d.sql
}
