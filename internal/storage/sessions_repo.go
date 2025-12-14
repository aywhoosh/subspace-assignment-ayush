package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SessionsRepo struct {
	db *sql.DB
}

func (r *SessionsRepo) Upsert(ctx context.Context, s Session) error {
	if err := mustNonEmpty("session.key", s.Key); err != nil {
		return err
	}
	if err := mustNonEmpty("session.cookies_json", s.CookiesJSON); err != nil {
		return err
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = nowUTC()
	}
	if s.LastUsedAt.IsZero() {
		s.LastUsedAt = s.CreatedAt
	}

	q := `INSERT INTO sessions(key, cookies_json, created_at, last_used_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET cookies_json=excluded.cookies_json, last_used_at=excluded.last_used_at`
	if _, err := r.db.ExecContext(ctx, q,
		s.Key,
		s.CookiesJSON,
		s.CreatedAt.Format(time.RFC3339Nano),
		s.LastUsedAt.Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("storage: upsert session: %w", err)
	}
	return nil
}

func (r *SessionsRepo) Get(ctx context.Context, key string) (Session, error) {
	if err := mustNonEmpty("session.key", key); err != nil {
		return Session{}, err
	}

	q := `SELECT key, cookies_json, created_at, last_used_at FROM sessions WHERE key = ?`
	var s Session
	var createdAt, lastUsedAt string
	if err := r.db.QueryRowContext(ctx, q, key).Scan(&s.Key, &s.CookiesJSON, &createdAt, &lastUsedAt); err != nil {
		return Session{}, fmt.Errorf("storage: get session: %w", err)
	}
	var err error
	s.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Session{}, fmt.Errorf("storage: parse session created_at: %w", err)
	}
	s.LastUsedAt, err = time.Parse(time.RFC3339Nano, lastUsedAt)
	if err != nil {
		return Session{}, fmt.Errorf("storage: parse session last_used_at: %w", err)
	}
	return s, nil
}

func (r *SessionsRepo) Delete(ctx context.Context, key string) error {
	if err := mustNonEmpty("session.key", key); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE key = ?`, key); err != nil {
		return fmt.Errorf("storage: delete session: %w", err)
	}
	return nil
}
