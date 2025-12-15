package mocknet

import (
	"context"
	"fmt"
	"time"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
	"github.com/go-rod/rod"
)

// SendConnectionRequest sends a connection request to a user by profile ID
func SendConnectionRequest(ctx context.Context, br *browser.Client, baseURL, profileID, note string) error {
	page, err := br.NewPage(baseURL + "/profile/" + profileID)
	if err != nil {
		return fmt.Errorf("connect: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	return SendConnectionRequestWithPage(ctx, page, baseURL, profileID, note)
}

// SendConnectionRequestWithPage sends a connection request using an existing page
func SendConnectionRequestWithPage(ctx context.Context, page *rod.Page, baseURL, profileID, note string) error {
	// Navigate to profile
	if err := page.Navigate(baseURL + "/profile/" + profileID); err != nil {
		return fmt.Errorf("connect: navigate: %w", err)
	}

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("connect: wait load: %w", err)
	}

	// Click connect button to open modal
	if err := click(page, "[data-testid='connect-button']"); err != nil {
		return fmt.Errorf("connect: click button: %w", err)
	}

	// Wait for modal to appear
	time.Sleep(500 * time.Millisecond)

	// Fill note in the textarea
	if note != "" {
		noteEl, err := page.Timeout(5 * time.Second).Element("[data-testid='connect-note']")
		if err == nil {
			if err := noteEl.Input(note); err != nil {
				return fmt.Errorf("connect: input note: %w", err)
			}
		}
	}

	// Click send button in the modal
	if err := click(page, "[data-testid='connect-send']"); err != nil {
		return fmt.Errorf("connect: click send: %w", err)
	}

	// Wait for form submission
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
