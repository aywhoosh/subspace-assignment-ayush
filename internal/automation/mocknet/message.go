package mocknet

import (
	"context"
	"fmt"
	"time"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
	"github.com/go-rod/rod"
)

// Message represents a single message in a conversation
type Message struct {
	From      string
	Content   string
	Timestamp string
}

// SendMessage sends a message to a user
func SendMessage(ctx context.Context, br *browser.Client, baseURL, recipientID, messageText string) error {
	threadID := "t-" + recipientID
	page, err := br.NewPage(baseURL + "/messages?thread=" + threadID)
	if err != nil {
		return fmt.Errorf("send message: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	return SendMessageWithPage(ctx, page, baseURL, recipientID, messageText)
}

// SendMessageWithPage sends a message using an existing page
func SendMessageWithPage(ctx context.Context, page *rod.Page, baseURL, recipientID, messageText string) error {
	// Navigate to messages with thread query parameter
	threadID := "t-" + recipientID
	if err := page.Navigate(baseURL + "/messages?thread=" + threadID); err != nil {
		return fmt.Errorf("send message: navigate: %w", err)
	}

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("send message: wait load: %w", err)
	}

	// Type message into input
	inputEl, err := page.Timeout(5 * time.Second).Element("[data-testid='message-input']")
	if err != nil {
		return fmt.Errorf("send message: input not found: %w", err)
	}

	if err := inputEl.Input(messageText); err != nil {
		return fmt.Errorf("send message: input text: %w", err)
	}

	// Click send button
	if err := click(page, "[data-testid='message-send']"); err != nil {
		return fmt.Errorf("send message: click send: %w", err)
	}

	// Wait briefly for message to send
	time.Sleep(500 * time.Millisecond)

	return nil
}

// GetConversation retrieves messages from a conversation
func GetConversation(ctx context.Context, br *browser.Client, baseURL, recipientID string) ([]Message, error) {
	threadID := "t-" + recipientID
	page, err := br.NewPage(baseURL + "/messages?thread=" + threadID)
	if err != nil {
		return nil, fmt.Errorf("get conversation: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return nil, fmt.Errorf("get conversation: wait load: %w", err)
	}

	// Find message container
	container, err := page.Timeout(5 * time.Second).Element("[data-testid='messages-list']")
	if err != nil {
		return []Message{}, nil // No messages yet
	}

	// Get all message items
	items, err := container.Elements("[data-testid='message-item']")
	if err != nil {
		return []Message{}, nil
	}

	var messages []Message
	for _, item := range items {
		fromEl, err := item.Element("[data-testid='message-from']")
		if err != nil {
			continue
		}
		from, _ := fromEl.Text()

		contentEl, err := item.Element("[data-testid='message-content']")
		if err != nil {
			continue
		}
		content, _ := contentEl.Text()

		timestampEl, err := item.Element("[data-testid='message-timestamp']")
		if err != nil {
			continue
		}
		timestamp, _ := timestampEl.Text()

		messages = append(messages, Message{
			From:      from,
			Content:   content,
			Timestamp: timestamp,
		})
	}

	return messages, nil
}

// GetInbox returns list of conversation threads
func GetInbox(ctx context.Context, br *browser.Client, baseURL string) ([]string, error) {
	page, err := br.NewPage(baseURL + "/messages")
	if err != nil {
		return nil, fmt.Errorf("get inbox: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return nil, fmt.Errorf("get inbox: wait load: %w", err)
	}

	// Find conversations list
	container, err := page.Timeout(5 * time.Second).Element("[data-testid='conversations-list']")
	if err != nil {
		return []string{}, nil // No conversations
	}

	items, err := container.Elements("[data-testid='conversation-item']")
	if err != nil {
		return []string{}, nil
	}

	var names []string
	for _, item := range items {
		nameEl, err := item.Element("[data-testid='conversation-name']")
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
