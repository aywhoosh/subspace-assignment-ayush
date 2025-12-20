package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type Repositories struct {
	Sessions           *SessionsRepo
	Profiles           *ProfilesRepo
	ConnectionRequests *ConnectionRequestsRepo
	Connections        *ConnectionsRepo
	Messages           *MessagesRepo
	Runs               *RunsRepo
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Sessions:           &SessionsRepo{db: db},
		Profiles:           &ProfilesRepo{db: db},
		ConnectionRequests: &ConnectionRequestsRepo{db: db},
		Connections:        &ConnectionsRepo{db: db},
		Messages:           &MessagesRepo{db: db},
		Runs:               &RunsRepo{db: db},
	}
}

// Helpers used by repositories.

func nowUTC() time.Time { return time.Now().UTC() }

func mustNonEmpty(field, value string) error {
	if value == "" {
		return fmt.Errorf("storage: %s is required", field)
	}
	return nil
}
