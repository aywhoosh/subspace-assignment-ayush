package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migration struct {
	Version int
	SQL     string
}

func Migrate(ctx context.Context, db *sql.DB) error {
	migs, err := loadMigrations()
	if err != nil {
		return err
	}
	if len(migs) == 0 {
		return nil
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("storage: enable foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);`); err != nil {
		return fmt.Errorf("storage: ensure schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return err
	}

	for _, m := range migs {
		if applied[m.Version] {
			continue
		}
		if err := applyOne(ctx, db, m); err != nil {
			return err
		}
	}
	return nil
}

func applyOne(ctx context.Context, db *sql.DB, m Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("storage: apply migration %d: %w", m.Version, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, m.Version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("storage: record migration %d: %w", m.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: commit migration %d: %w", m.Version, err)
	}
	return nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("storage: read schema_migrations: %w", err)
	}
	defer rows.Close()

	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("storage: scan schema_migrations: %w", err)
		}
		out[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate schema_migrations: %w", err)
	}
	return out, nil
}

func loadMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("storage: read migrations dir: %w", err)
	}

	var migs []Migration
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		vStr, _, ok := strings.Cut(name, "_")
		if !ok {
			return nil, fmt.Errorf("storage: invalid migration filename %q", name)
		}
		v, err := strconv.Atoi(vStr)
		if err != nil {
			return nil, fmt.Errorf("storage: invalid migration version %q", name)
		}
		b, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return nil, fmt.Errorf("storage: read migration %q: %w", name, err)
		}
		sql := extractUpSQL(string(b))
		migs = append(migs, Migration{Version: v, SQL: sql})
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].Version < migs[j].Version })
	return migs, nil
}

func extractUpSQL(s string) string {
	// Very small convention parser: only run statements between "-- +subspace Up" and "-- +subspace Down".
	// If markers are missing, run entire file.
	upMarker := "-- +subspace Up"
	downMarker := "-- +subspace Down"
	upIdx := strings.Index(s, upMarker)
	if upIdx < 0 {
		return s
	}
	downIdx := strings.Index(s, downMarker)
	if downIdx < 0 || downIdx <= upIdx {
		return s[upIdx+len(upMarker):]
	}
	return s[upIdx+len(upMarker) : downIdx]
}
