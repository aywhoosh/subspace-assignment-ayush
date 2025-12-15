package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// If opts.Path is a normal file path (not :memory: / not a URI), ensure the
	// parent directory exists. On Windows, missing directories often surface as
	// SQLITE_CANTOPEN with confusing messages.
	if dir := sqliteParentDir(opts.Path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("storage: create sqlite dir %q: %w", dir, err)
		}
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

func sqliteParentDir(dsn string) string {
	s := strings.TrimSpace(dsn)
	if s == "" {
		return ""
	}
	if s == ":memory:" {
		return ""
	}
	if strings.Contains(s, "mode=memory") {
		return ""
	}
	if strings.HasPrefix(s, "file:") {
		// DSN format: file:path[?params]
		p := strings.TrimPrefix(s, "file:")
		if q := strings.IndexByte(p, '?'); q >= 0 {
			p = p[:q]
		}
		p = strings.TrimPrefix(p, "//")
		p = strings.TrimSpace(p)
		if p == "" {
			return ""
		}
		return filepath.Dir(p)
	}
	// Treat as a plain file path.
	return filepath.Dir(s)
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) SQL() *sql.DB {
	return d.sql
}
