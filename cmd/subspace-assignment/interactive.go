package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	automationMocknet "github.com/aywhoosh/subspace-assignment-ayush/internal/automation/mocknet"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/config"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/storage"
	"github.com/go-rod/rod"
)

func runInteractiveMode(ctx context.Context, cfg config.Config) error {
	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║            SUBSPACE AUTOMATION CONTROL PANEL                   ║")
	fmt.Println("║              Browser will stay open between runs               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Open database once
	db, repos, err := openRepos(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create browser once - NON-HEADLESS so user can see it
	browserCfg := browser.Config{
		Headless:      false, // Force visible browser
		SlowMo:        cfg.Browser.SlowMo,
		Leakless:      cfg.Browser.Leakless,
		BinPath:       cfg.Browser.BinPath,
		AllowDownload: cfg.Browser.AllowDownload,
	}

	br, cleanup, err := browser.New(ctx, browserCfg)
	if err != nil {
		return fmt.Errorf("failed to create browser: %w", err)
	}
	defer func() { _ = cleanup() }()

	// Create one persistent page that stays open
	page, err := br.NewPage("about:blank")
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer func() { _ = page.Close() }()

	fmt.Println("✓ Browser launched successfully")
	fmt.Println("✓ Database connected")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	authenticated := false

	for {
		printMenu(authenticated)
		fmt.Print("\n> ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			if err := runLogin(ctx, cfg, br, repos, page); err != nil {
				fmt.Printf("❌ Login failed: %v\n\n", err)
			} else {
				authenticated = true
				fmt.Println("✓ Login successful")
				fmt.Println()
			}

		case "2":
			if !authenticated {
				fmt.Println("⚠️  Please login first (option 1)")
				fmt.Println()
				continue
			}
			if err := runSearch(ctx, cfg, br, repos, page); err != nil {
				fmt.Printf("❌ Search failed: %v\n\n", err)
			}

		case "3":
			if !authenticated {
				fmt.Println("⚠️  Please login first (option 1)")
				fmt.Println()
				continue
			}
			if err := runConnect(ctx, cfg, br, repos, page, scanner); err != nil {
				fmt.Printf("❌ Connect failed: %v\n\n", err)
			}

		case "4":
			if !authenticated {
				fmt.Println("⚠️  Please login first (option 1)")
				fmt.Println()
				continue
			}
			if err := runMessage(ctx, cfg, br, repos, page, scanner); err != nil {
				fmt.Printf("❌ Message failed: %v\n\n", err)
			}

		case "5":
			if err := runCheckSession(ctx, cfg, br, repos, page); err != nil {
				fmt.Printf("❌ Session check failed: %v\n\n", err)
				authenticated = false
			} else {
				authenticated = true
			}

		case "6":
			if !authenticated {
				fmt.Println("⚠️  Please login first (option 1)")
				fmt.Println()
				continue
			}
			if err := runViewInbox(ctx, cfg, br, repos, page); err != nil {
				fmt.Printf("❌ Inbox failed: %v\n\n", err)
			}

		case "0", "q", "quit", "exit":
			fmt.Println("\n👋 Goodbye! Browser will close now.")
			return nil

		default:
			fmt.Println("❌ Invalid choice. Please try again.")
			fmt.Println()
		}
	}

	return nil
}

func printMenu(authenticated bool) {
	status := "🔴 Not authenticated"
	if authenticated {
		status = "🟢 Authenticated"
	}

	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Printf("  Status: %s\n", status)
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  1. 🔐 Login to MockNet")
	fmt.Println("  2. 🔍 Search for profiles")
	fmt.Println("  3. 🤝 Send connection request")
	fmt.Println("  4. 💬 Send message")
	fmt.Println("  5. ✓  Check session status")
	fmt.Println("  6. 📥 View inbox")
	fmt.Println("  0. 🚪 Exit")
	fmt.Println("──────────────────────────────────────────────────────────────")
}

func runLogin(ctx context.Context, cfg config.Config, br *browser.Client, repos *storage.Repositories, page *rod.Page) error {
	fmt.Println("\n🔐 Logging in to MockNet...")

	creds := automationMocknet.Credentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	// Use LoginWithPage to work with persistent page
	username, cookiesJSON, err := automationMocknet.LoginWithPage(ctx, page, cfg.Mocknet.BaseURL, creds)
	if err != nil {
		return err
	}

	// Save session to database
	key := automationMocknet.SessionKey(cfg.Mocknet.BaseURL, creds.Username)
	if err := repos.Sessions.Upsert(ctx, storage.Session{
		Key:         key,
		CookiesJSON: cookiesJSON,
	}); err != nil {
		return err
	}

	fmt.Printf("✓ Authenticated as: %s\n", username)
	fmt.Printf("✓ Session saved\n\n")
	return nil
}

func runCheckSession(ctx context.Context, cfg config.Config, br *browser.Client, repos *storage.Repositories, page *rod.Page) error {
	fmt.Println("\n✓ Checking session...")

	creds := automationMocknet.Credentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}
	opts := automationMocknet.Options{Timeout: cfg.Run.Timeout}

	username, err := automationMocknet.EnsureAuthed(ctx, br, repos, cfg.Mocknet.BaseURL, creds, opts)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Session valid. Authenticated as: %s\n\n", username)
	return nil
}

func runSearch(ctx context.Context, cfg config.Config, br *browser.Client, repos *storage.Repositories, page *rod.Page) error {
	fmt.Println("\n🔍 Searching for profiles...")
	fmt.Print("  Enter job title (or press Enter for 'Engineer'): ")

	scanner := bufio.NewScanner(os.Stdin)
	var title string
	if scanner.Scan() {
		title = strings.TrimSpace(scanner.Text())
	}
	if title == "" {
		title = "Engineer"
	}

	// Navigate to search page on the persistent page
	if err := page.Navigate(cfg.Mocknet.BaseURL + "/search"); err != nil {
		return fmt.Errorf("navigate to search: %w", err)
	}
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("wait load: %w", err)
	}

	searchOpts := automationMocknet.SearchOptions{
		Title: title,
	}

	results, err := automationMocknet.SearchOnPage(ctx, page, searchOpts)
	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Found %d results:\n", len(results))
	for i, r := range results {
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, r.Name, r.ProfileID, r.Title)
	}
	fmt.Println()
	return nil
}

func runConnect(ctx context.Context, cfg config.Config, br *browser.Client, repos *storage.Repositories, page *rod.Page, scanner *bufio.Scanner) error {
	fmt.Println("\n🤝 Sending connection request...")
	fmt.Print("  Enter profile ID: ")

	if !scanner.Scan() {
		return fmt.Errorf("no input")
	}
	profileID := strings.TrimSpace(scanner.Text())

	if profileID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}

	fmt.Print("  Enter note (optional): ")
	var note string
	if scanner.Scan() {
		note = strings.TrimSpace(scanner.Text())
	}
	if note == "" {
		note = "I'd like to connect with you!"
	}

	err := automationMocknet.SendConnectionRequestWithPage(ctx, page, cfg.Mocknet.BaseURL, profileID, note)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Connection request sent to profile %s\n\n", profileID)
	return nil
}

func runMessage(ctx context.Context, cfg config.Config, br *browser.Client, repos *storage.Repositories, page *rod.Page, scanner *bufio.Scanner) error {
	fmt.Println("\n💬 Sending message...")
	fmt.Print("  Enter user ID: ")

	if !scanner.Scan() {
		return fmt.Errorf("no input")
	}
	userID := strings.TrimSpace(scanner.Text())

	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	fmt.Print("  Enter message: ")
	if !scanner.Scan() {
		return fmt.Errorf("no input")
	}
	message := strings.TrimSpace(scanner.Text())

	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	// First, ensure connection is accepted (navigate to connections and accept)
	fmt.Println("  Checking connection status...")
	if err := page.Navigate(cfg.Mocknet.BaseURL + "/connections"); err == nil {
		_ = page.Timeout(5 * time.Second).WaitLoad()
		time.Sleep(500 * time.Millisecond)
		
		// Try to find and accept pending connection for this user
		pendingItem, err := page.Timeout(2 * time.Second).Element(fmt.Sprintf("[data-testid='pending-connection'][data-profile-id='%s']", userID))
		if err == nil {
			// Found pending connection, click accept button
			acceptBtn, err := pendingItem.Element("button[data-testid='pending-accept']")
			if err == nil {
				fmt.Println("  Accepting pending connection...")
				// Wait for navigation after form submission
				wait := page.MustWaitNavigation()
				acceptBtn.MustClick()
				wait()
				_ = page.Timeout(10 * time.Second).WaitLoad()
				time.Sleep(4 * time.Second) // Wait longer for connection to be fully processed in backend
				
				// Navigate back to connections to verify
				_ = page.Navigate(cfg.Mocknet.BaseURL + "/connections")
				_ = page.Timeout(5 * time.Second).WaitLoad()
				time.Sleep(500 * time.Millisecond)
				
				// Verify connection was accepted by checking for accepted connection
				_, err = page.Timeout(5 * time.Second).Element(fmt.Sprintf("[data-testid='accepted-connection'][data-profile-id='%s']", userID))
				if err != nil {
					return fmt.Errorf("connection acceptance may have failed - connection not found in accepted list after waiting. Please manually accept the connection at %s/connections and try again", cfg.Mocknet.BaseURL)
				}
				fmt.Println("  ✓ Connection accepted and verified")
			}
		} else {
			// No pending connection, check if already connected
			_, err = page.Timeout(2 * time.Second).Element(fmt.Sprintf("[data-testid='accepted-connection'][data-profile-id='%s']", userID))
			if err != nil {
				return fmt.Errorf("no connection found with user %s. Please send a connection request first (option 3)", userID)
			}
			fmt.Println("  ✓ Already connected")
		}
	}

	// Navigate to messages tab first to ensure thread list is loaded
	fmt.Println("  Opening messages...")
	if err := page.Navigate(cfg.Mocknet.BaseURL + "/messages"); err != nil {
		return fmt.Errorf("navigate to messages: %w", err)
	}
	if err := page.Timeout(5 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("wait for messages page: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Now navigate to the specific thread
	threadID := "thread-" + userID
	fmt.Println("  Opening conversation thread...")
	if err := page.Navigate(cfg.Mocknet.BaseURL + "/messages?thread=" + threadID); err != nil {
		return fmt.Errorf("navigate to thread: %w", err)
	}
	if err := page.Timeout(5 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("wait for thread: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Type message into input
	fmt.Println("  Typing message...")
	inputEl, err := page.Timeout(5 * time.Second).Element("[data-testid='message-input']")
	if err != nil {
		return fmt.Errorf("message input not found: %w", err)
	}
	if err := inputEl.Input(message); err != nil {
		return fmt.Errorf("input text: %w", err)
	}

	// Send message using JavaScript fetch to avoid form navigation
	fmt.Println("  Sending...")
	_, err = page.Eval(`async (profileID, messageBody) => {
		const response = await fetch('/messages/send', {
			method: 'POST',
			headers: {'Content-Type': 'application/x-www-form-urlencoded'},
			body: new URLSearchParams({
				profile_id: profileID,
				body: messageBody
			})
		});
		return response.ok;
	}`, userID, message)
	
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	
	time.Sleep(1 * time.Second)
	
	// Reload page to show the message
	_ = page.Reload()
	_ = page.Timeout(5 * time.Second).WaitLoad()

	fmt.Printf("✓ Message sent to user %s\n\n", userID)
	return nil
}

func runViewInbox(ctx context.Context, cfg config.Config, br *browser.Client, repos *storage.Repositories, page *rod.Page) error {
	fmt.Println("\n📥 Viewing inbox...")

	// Navigate to inbox
	if err := page.Navigate(cfg.Mocknet.BaseURL + "/messages"); err != nil {
		return err
	}
	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return err
	}

	// Find conversations list
	container, err := page.Timeout(5 * time.Second).Element("[data-testid='conversations-list']")
	if err != nil {
		fmt.Println("  No conversations found")
		fmt.Println()
		return nil
	}

	items, err := container.Elements("[data-testid='conversation-item']")
	if err != nil {
		fmt.Println("  No conversations found")
		fmt.Println()
		return nil
	}

	var conversations []string
	for _, item := range items {
		nameEl, err := item.Element("[data-testid='conversation-name']")
		if err != nil {
			continue
		}
		name, err := nameEl.Text()
		if err != nil {
			continue
		}
		conversations = append(conversations, name)
	}
	if err != nil {
		return err
	}

	if len(conversations) == 0 {
		fmt.Println("  No conversations found")
		fmt.Println()
		return nil
	}

	fmt.Printf("✓ Found %d conversation(s):\n", len(conversations))
	for i, name := range conversations {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
	fmt.Println()
	return nil
}
