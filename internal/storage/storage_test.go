package storage

import (
	"context"
	"testing"
	"time"
)

func TestMigrateAndRepositories_BasicFlow(t *testing.T) {
	ctx := context.Background()

	db, err := Open(ctx, OpenOptions{Path: "file:memdb1?mode=memory&cache=shared"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := Migrate(ctx, db.SQL()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	repos := NewRepositories(db.SQL())

	p := Profile{ProfileID: "p1", URL: "http://localhost:8080/profile/p1", FirstName: "A", Company: "C"}
	if err := repos.Profiles.Upsert(ctx, p); err != nil {
		t.Fatalf("Profiles.Upsert: %v", err)
	}
	got, err := repos.Profiles.GetByID(ctx, "p1")
	if err != nil {
		t.Fatalf("Profiles.GetByID: %v", err)
	}
	if got.URL != p.URL {
		t.Fatalf("expected url %q got %q", p.URL, got.URL)
	}

	// Ledger: record a connection request once.
	req := ConnectionRequest{ProfileID: "p1", Note: "hello", Status: "sent"}
	if err := repos.ConnectionRequests.RecordSent(ctx, req); err != nil {
		t.Fatalf("ConnectionRequests.RecordSent: %v", err)
	}
	exists, err := repos.ConnectionRequests.Exists(ctx, "p1")
	if err != nil {
		t.Fatalf("ConnectionRequests.Exists: %v", err)
	}
	if !exists {
		t.Fatalf("expected request to exist")
	}

	// Mark accepted + record message.
	if err := repos.Connections.MarkAccepted(ctx, Connection{ProfileID: "p1"}); err != nil {
		t.Fatalf("Connections.MarkAccepted: %v", err)
	}
	inserted, err := repos.Messages.RecordSent(ctx, Message{ThreadID: "t1", ProfileID: "p1", TemplateID: "followup-1", Body: "hi"})
	if err != nil {
		t.Fatalf("Messages.RecordSent: %v", err)
	}
	if !inserted {
		t.Fatalf("expected message insert")
	}
	inserted2, err := repos.Messages.RecordSent(ctx, Message{ThreadID: "t1", ProfileID: "p1", TemplateID: "followup-1", Body: "hi"})
	if err != nil {
		t.Fatalf("Messages.RecordSent(dup): %v", err)
	}
	if inserted2 {
		t.Fatalf("expected duplicate message to be ignored")
	}

	// Counts.
	since := time.Now().UTC().Add(-1 * time.Hour)
	c1, err := repos.ConnectionRequests.CountSentSince(ctx, since)
	if err != nil {
		t.Fatalf("ConnectionRequests.CountSentSince: %v", err)
	}
	if c1 != 1 {
		t.Fatalf("expected 1 request, got %d", c1)
	}
	m1, err := repos.Messages.CountSentSince(ctx, since)
	if err != nil {
		t.Fatalf("Messages.CountSentSince: %v", err)
	}
	if m1 != 1 {
		t.Fatalf("expected 1 message, got %d", m1)
	}
}
