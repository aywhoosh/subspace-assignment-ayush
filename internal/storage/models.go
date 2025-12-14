package storage

import "time"

type Session struct {
	Key        string
	CookiesJSON string
	CreatedAt  time.Time
	LastUsedAt time.Time
}

type Profile struct {
	ProfileID string
	URL       string
	FirstName string
	LastName  string
	Company   string
	Title     string
	Location  string
	Keywords  string
	UpdatedAt time.Time
}

type ConnectionRequest struct {
	ProfileID string
	SentAt    time.Time
	Note      string
	Status    string
}

type Connection struct {
	ProfileID  string
	AcceptedAt time.Time
}

type Message struct {
	ThreadID   string
	ProfileID  string
	TemplateID string
	Body       string
	SentAt     time.Time
}

type Run struct {
	RunID        string
	StartedAt    time.Time
	EndedAt      *time.Time
	CountersJSON string
	Outcome      string
}
