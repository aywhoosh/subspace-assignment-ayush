package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ConnectionsRepo struct {
	db *sql.DB
}

func (r *ConnectionsRepo) MarkAccepted(ctx context.Context, c Connection) error {
	if err := mustNonEmpty("connection.profile_id", c.ProfileID); err != nil {
		return err
	}
	if c.AcceptedAt.IsZero() {
		c.AcceptedAt = nowUTC()
	}

	q := `INSERT INTO connections(profile_id, accepted_at)
		VALUES(?, ?)
		ON CONFLICT(profile_id) DO UPDATE SET accepted_at=excluded.accepted_at`
	if _, err := r.db.ExecContext(ctx, q, c.ProfileID, c.AcceptedAt.Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("storage: mark connection accepted: %w", err)
	}
	return nil
}

func (r *ConnectionsRepo) ListAcceptedSince(ctx context.Context, since time.Time) ([]Connection, error) {
	q := `SELECT profile_id, accepted_at FROM connections WHERE accepted_at >= ? ORDER BY accepted_at ASC`
	rows, err := r.db.QueryContext(ctx, q, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, fmt.Errorf("storage: list connections: %w", err)
	}
	defer rows.Close()

	var out []Connection
	for rows.Next() {
		var c Connection
		var acceptedAt string
		if err := rows.Scan(&c.ProfileID, &acceptedAt); err != nil {
			return nil, fmt.Errorf("storage: scan connection: %w", err)
		}
		tm, err := time.Parse(time.RFC3339Nano, acceptedAt)
		if err != nil {
			return nil, fmt.Errorf("storage: parse connection accepted_at: %w", err)
		}
		c.AcceptedAt = tm
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate connections: %w", err)
	}
	return out, nil
}
