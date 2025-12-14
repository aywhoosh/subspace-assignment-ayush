package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ConnectionRequestsRepo struct {
	db *sql.DB
}

func (r *ConnectionRequestsRepo) RecordSent(ctx context.Context, req ConnectionRequest) error {
	if err := mustNonEmpty("connection_request.profile_id", req.ProfileID); err != nil {
		return err
	}
	if err := mustNonEmpty("connection_request.note", req.Note); err != nil {
		return err
	}
	if err := mustNonEmpty("connection_request.status", req.Status); err != nil {
		return err
	}
	if req.SentAt.IsZero() {
		req.SentAt = nowUTC()
	}

	q := `INSERT INTO connection_requests(profile_id, sent_at, note, status)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(profile_id) DO UPDATE SET sent_at=excluded.sent_at, note=excluded.note, status=excluded.status`
	if _, err := r.db.ExecContext(ctx, q,
		req.ProfileID,
		req.SentAt.Format(time.RFC3339Nano),
		req.Note,
		req.Status,
	); err != nil {
		return fmt.Errorf("storage: record connection_request: %w", err)
	}
	return nil
}

func (r *ConnectionRequestsRepo) Exists(ctx context.Context, profileID string) (bool, error) {
	if err := mustNonEmpty("connection_request.profile_id", profileID); err != nil {
		return false, err
	}

	q := `SELECT 1 FROM connection_requests WHERE profile_id = ? LIMIT 1`
	var one int
	err := r.db.QueryRowContext(ctx, q, profileID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("storage: exists connection_request: %w", err)
	}
	return true, nil
}

func (r *ConnectionRequestsRepo) CountSentSince(ctx context.Context, since time.Time) (int, error) {
	q := `SELECT COUNT(1) FROM connection_requests WHERE sent_at >= ?`
	var n int
	if err := r.db.QueryRowContext(ctx, q, since.UTC().Format(time.RFC3339Nano)).Scan(&n); err != nil {
		return 0, fmt.Errorf("storage: count connection_requests since: %w", err)
	}
	return n, nil
}
