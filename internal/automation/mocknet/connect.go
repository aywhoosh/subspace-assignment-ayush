package mocknet

import (
	"context"
	"fmt"
	"time"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
)

// SendConnectionRequest sends a connection request to a user by profile ID
func SendConnectionRequest(ctx context.Context, br *browser.Client, baseURL, profileID, note string) error {
	page, err := br.NewPage(baseURL + "/profile/" + profileID)
	if err != nil {
		return fmt.Errorf("connect: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("connect: wait load: %w", err)
	}

	// Click connect button
	if err := click(page, "[data-testid='profile-connect-btn']"); err != nil {
		return fmt.Errorf("connect: click button: %w", err)
	}

	// Wait for note modal if it appears
	time.Sleep(500 * time.Millisecond)

	// If note provided and modal exists, add note
	if note != "" {
		noteEl, err := page.Timeout(2 * time.Second).Element("[data-testid='connect-note-input']")
		if err == nil {
			if err := noteEl.Input(note); err != nil {
				return fmt.Errorf("connect: input note: %w", err)
			}
		}
	}

	// Click send/submit (either in modal or directly)
	_, err = page.Timeout(5 * time.Second).Element("[data-testid='connect-send-btn']")
	if err != nil {
		// Maybe already sent without modal
		return nil
	}

	if err := click(page, "[data-testid='connect-send-btn']"); err != nil {
		return fmt.Errorf("connect: send: %w", err)
	}

	// Wait briefly for confirmation
	time.Sleep(500 * time.Millisecond)

	return nil
}

// GetPendingRequests returns list of pending connection requests
func GetPendingRequests(ctx context.Context, br *browser.Client, baseURL string) ([]string, error) {
	page, err := br.NewPage(baseURL + "/connections/pending")
	if err != nil {
		return nil, fmt.Errorf("pending requests: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return nil, fmt.Errorf("pending requests: wait load: %w", err)
	}

	// Find all pending request items
	container, err := page.Timeout(5 * time.Second).Element("[data-testid='pending-requests-list']")
	if err != nil {
		return []string{}, nil // No pending requests
	}

	items, err := container.Elements("[data-testid='pending-request-item']")
	if err != nil {
		return []string{}, nil
	}

	var names []string
	for _, item := range items {
		nameEl, err := item.Element("[data-testid='pending-request-name']")
		if err != nil {
			continue
		}
		name, err := nameEl.Text()
		if err != nil {
			continue
		}
		names = append(names, name)
	}

	return names, nil
}
