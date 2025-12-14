package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type RunsRepo struct {
	db *sql.DB
}

func (r *RunsRepo) Start(ctx context.Context, run Run) error {
	if err := mustNonEmpty("run.run_id", run.RunID); err != nil {
		return err
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = nowUTC()
	}
	if run.CountersJSON == "" {
		run.CountersJSON = "{}"
	}
	if run.Outcome == "" {
		run.Outcome = "running"
	}

	q := `INSERT INTO runs(run_id, started_at, ended_at, counters_json, outcome)
		VALUES(?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		run.RunID,
		run.StartedAt.Format(time.RFC3339Nano),
		nullTime(run.EndedAt),
		run.CountersJSON,
		run.Outcome,
	)
	if err != nil {
		return fmt.Errorf("storage: start run: %w", err)
	}
	return nil
}

func (r *RunsRepo) Finish(ctx context.Context, runID string, outcome string, countersJSON string) error {
	if err := mustNonEmpty("run.run_id", runID); err != nil {
		return err
	}
	if outcome == "" {
		outcome = "unknown"
	}
	if countersJSON == "" {
		countersJSON = "{}"
	}

	q := `UPDATE runs SET ended_at = ?, outcome = ?, counters_json = ? WHERE run_id = ?`
	if _, err := r.db.ExecContext(ctx, q, nowUTC().Format(time.RFC3339Nano), outcome, countersJSON, runID); err != nil {
		return fmt.Errorf("storage: finish run: %w", err)
	}
	return nil
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}
