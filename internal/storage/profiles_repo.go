package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ProfilesRepo struct {
	db *sql.DB
}

func (r *ProfilesRepo) Upsert(ctx context.Context, p Profile) error {
	if err := mustNonEmpty("profile.profile_id", p.ProfileID); err != nil {
		return err
	}
	if err := mustNonEmpty("profile.url", p.URL); err != nil {
		return err
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = nowUTC()
	}

	q := `INSERT INTO profiles(profile_id, url, first_name, last_name, company, title, location, keywords, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id) DO UPDATE SET
			url=excluded.url,
			first_name=excluded.first_name,
			last_name=excluded.last_name,
			company=excluded.company,
			title=excluded.title,
			location=excluded.location,
			keywords=excluded.keywords,
			updated_at=excluded.updated_at`

	if _, err := r.db.ExecContext(ctx, q,
		p.ProfileID,
		p.URL,
		nullIfEmpty(p.FirstName),
		nullIfEmpty(p.LastName),
		nullIfEmpty(p.Company),
		nullIfEmpty(p.Title),
		nullIfEmpty(p.Location),
		nullIfEmpty(p.Keywords),
		p.UpdatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("storage: upsert profile: %w", err)
	}
	return nil
}

func (r *ProfilesRepo) GetByID(ctx context.Context, profileID string) (Profile, error) {
	if err := mustNonEmpty("profile.profile_id", profileID); err != nil {
		return Profile{}, err
	}

	q := `SELECT profile_id, url, COALESCE(first_name,''), COALESCE(last_name,''), COALESCE(company,''), COALESCE(title,''), COALESCE(location,''), COALESCE(keywords,''), updated_at
		FROM profiles WHERE profile_id = ?`
	var p Profile
	var updatedAt string
	if err := r.db.QueryRowContext(ctx, q, profileID).Scan(
		&p.ProfileID,
		&p.URL,
		&p.FirstName,
		&p.LastName,
		&p.Company,
		&p.Title,
		&p.Location,
		&p.Keywords,
		&updatedAt,
	); err != nil {
		return Profile{}, fmt.Errorf("storage: get profile: %w", err)
	}
	var err error
	p.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("storage: parse profile updated_at: %w", err)
	}
	return p, nil
}

func (r *ProfilesRepo) ExistsByURL(ctx context.Context, url string) (bool, error) {
	if err := mustNonEmpty("profile.url", url); err != nil {
		return false, err
	}

	q := `SELECT 1 FROM profiles WHERE url = ? LIMIT 1`
	var one int
	err := r.db.QueryRowContext(ctx, q, url).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("storage: exists profile by url: %w", err)
	}
	return true, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
