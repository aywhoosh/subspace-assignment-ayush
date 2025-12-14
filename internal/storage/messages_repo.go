package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type MessagesRepo struct {
	db *sql.DB
}

func (r *MessagesRepo) RecordSent(ctx context.Context, m Message) (bool, error) {
	if err := mustNonEmpty("message.thread_id", m.ThreadID); err != nil {
		return false, err
	}
	if err := mustNonEmpty("message.profile_id", m.ProfileID); err != nil {
		return false, err
	}
	if err := mustNonEmpty("message.template_id", m.TemplateID); err != nil {
		return false, err
	}
	if err := mustNonEmpty("message.body", m.Body); err != nil {
		return false, err
	}
	if m.SentAt.IsZero() {
		m.SentAt = nowUTC()
	}

	q := `INSERT OR IGNORE INTO messages(thread_id, profile_id, template_id, body, sent_at)
		VALUES(?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q,
		m.ThreadID,
		m.ProfileID,
		m.TemplateID,
		m.Body,
		m.SentAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("storage: record message: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (r *MessagesRepo) CountSentSince(ctx context.Context, since time.Time) (int, error) {
	q := `SELECT COUNT(1) FROM messages WHERE sent_at >= ?`
	var n int
	if err := r.db.QueryRowContext(ctx, q, since.UTC().Format(time.RFC3339Nano)).Scan(&n); err != nil {
		return 0, fmt.Errorf("storage: count messages since: %w", err)
	}
	return n, nil
}
