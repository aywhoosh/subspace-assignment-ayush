package mocknet

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/storage"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Credentials struct {
	Username string
	Password string
}

type Options struct {
	Timeout time.Duration
}

func SessionKey(baseURL, username string) string {
	user := strings.TrimSpace(username)
	if user == "" {
		user = "(unknown)"
	}
	return "mocknet|" + strings.TrimSpace(baseURL) + "|" + user
}

func EnsureAuthed(ctx context.Context, br *browser.Client, repos *storage.Repositories, baseURL string, creds Credentials, opts Options) (string, error) {
	if err := ensureLocalBaseURL(baseURL); err != nil {
		return "", err
	}
	if strings.TrimSpace(creds.Username) == "" {
		creds.Username = "demo"
	}
	if strings.TrimSpace(creds.Password) == "" {
		creds.Password = "demo"
	}
	key := SessionKey(baseURL, creds.Username)

	// 1) Try existing session.
	if s, err := repos.Sessions.Get(ctx, key); err == nil {
		params, err := browser.CookiesFromJSON(s.CookiesJSON)
		if err == nil {
			_ = browser.SetCookies(ctx, br.Browser(), params)
			if username, ok := checkAuthed(ctx, br, baseURL, opts.Timeout); ok {
				_ = repos.Sessions.Upsert(ctx, storage.Session{
					Key:         key,
					CookiesJSON: s.CookiesJSON,
					CreatedAt:   s.CreatedAt,
					LastUsedAt:  time.Now().UTC(),
				})
				return username, nil
			}
		}
	}

	// 2) Fresh login.
	username, cookiesJSON, err := loginAndCaptureCookies(ctx, br, baseURL, creds, opts.Timeout)
	if err != nil {
		return "", err
	}

	if err := repos.Sessions.Upsert(ctx, storage.Session{
		Key:         key,
		CookiesJSON: cookiesJSON,
	}); err != nil {
		return "", err
	}
	return username, nil
}

func loginAndCaptureCookies(ctx context.Context, br *browser.Client, baseURL string, creds Credentials, timeout time.Duration) (string, string, error) {
	page, err := br.NewPage(baseURL + "/login")
	if err != nil {
		return "", "", err
	}
	defer func() { _ = page.Close() }()

	// Use a simpler approach: don't set global page timeout, handle timeouts per operation
	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return "", "", fmt.Errorf("automation: wait load: %w", err)
	}

	if err := typeInto(page, "[data-testid='login-username']", creds.Username); err != nil {
		return "", "", err
	}
	if err := typeInto(page, "[data-testid='login-password']", creds.Password); err != nil {
		return "", "", err
	}
	// Click login button and wait for the POST + redirect to complete
	wait := page.MustWaitNavigation()
	if err := click(page, "[data-testid='login-submit']"); err != nil {
		return "", "", err
	}
	wait()

	// After successful login, we should be redirected to /search.
	// Now we need to:
	// 1. Verify we're authenticated by finding the username element
	// 2. Capture the session cookies
	// 3. Save everything to the database

	// Wait for the page to be fully loaded and interactive
	if err := page.WaitLoad(); err != nil {
		return "", "", fmt.Errorf("automation: wait for search page: %w", err)
	}

	// Look for the username element with a reasonable timeout
	el, err := page.Timeout(5 * time.Second).Element("[data-testid='nav-user']")
	if err != nil {
		return "", "", fmt.Errorf("automation: nav-user not found (not authenticated): %w", err)
	}

	username, err := el.Text()
	if err != nil || strings.TrimSpace(username) == "" {
		return "", "", errors.New("automation: failed to read username from nav")
	}

	// Capture cookies from the browser to persist the session
	cookies, err := browser.GetAllCookies(ctx, br.Browser())
	if err != nil {
		return "", "", fmt.Errorf("automation: get cookies: %w", err)
	}

	cookiesJSON, err := browser.CookiesToJSON(cookies)
	if err != nil {
		return "", "", fmt.Errorf("automation: serialize cookies: %w", err)
	}

	return strings.TrimSpace(username), cookiesJSON, nil
}

func checkAuthed(ctx context.Context, br *browser.Client, baseURL string, timeout time.Duration) (string, bool) {
	page, err := br.NewPage(baseURL + "/search")
	if err != nil {
		return "", false
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return "", false
	}

	el, err := page.Timeout(5 * time.Second).Element("[data-testid='nav-user']")
	if err != nil {
		return "", false
	}

	username, err := el.Text()
	if err != nil || strings.TrimSpace(username) == "" {
		return "", false
	}
	return strings.TrimSpace(username), true
}

// LoginWithPage performs login using an existing page (for interactive mode with persistent browser)
func LoginWithPage(ctx context.Context, page *rod.Page, baseURL string, creds Credentials) (string, string, error) {
	if err := ensureLocalBaseURL(baseURL); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(creds.Username) == "" {
		creds.Username = "demo"
	}
	if strings.TrimSpace(creds.Password) == "" {
		creds.Password = "demo"
	}

	// Navigate to login page
	if err := page.Navigate(baseURL + "/login"); err != nil {
		return "", "", fmt.Errorf("navigate to login: %w", err)
	}
	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return "", "", fmt.Errorf("wait load: %w", err)
	}

	// Fill login form
	if err := typeInto(page, "[data-testid='login-username']", creds.Username); err != nil {
		return "", "", err
	}
	if err := typeInto(page, "[data-testid='login-password']", creds.Password); err != nil {
		return "", "", err
	}

	// Click login button and wait for navigation
	wait := page.MustWaitNavigation()
	if err := click(page, "[data-testid='login-submit']"); err != nil {
		return "", "", err
	}
	wait()

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return "", "", fmt.Errorf("wait for search page: %w", err)
	}

	// Verify authentication by finding username element
	el, err := page.Timeout(5 * time.Second).Element("[data-testid='nav-user']")
	if err != nil {
		return "", "", fmt.Errorf("nav-user not found (not authenticated): %w", err)
	}

	username, err := el.Text()
	if err != nil || strings.TrimSpace(username) == "" {
		return "", "", errors.New("failed to read username from nav")
	}

	// Capture cookies to save session
	cookies, err := browser.GetAllCookies(ctx, page.Browser())
	if err != nil {
		return "", "", fmt.Errorf("get cookies: %w", err)
	}

	cookiesJSON, err := browser.CookiesToJSON(cookies)
	if err != nil {
		return "", "", fmt.Errorf("serialize cookies: %w", err)
	}

	return strings.TrimSpace(username), cookiesJSON, nil
}

func ensureLocalBaseURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("automation: invalid base URL: %w", err)
	}
	if u.Scheme != "http" {
		return errors.New("automation: base URL must use http")
	}
	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return errors.New("automation: base URL must be localhost/127.0.0.1")
	}
	return nil
}

func click(page *rod.Page, selector string) error {
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("automation: find %s: %w", selector, err)
	}
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("automation: click %s: %w", selector, err)
	}
	return nil
}

func typeInto(page *rod.Page, selector, value string) error {
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("automation: find %s: %w", selector, err)
	}
	if err := el.Input(value); err != nil {
		return fmt.Errorf("automation: input %s: %w", selector, err)
	}
	return nil
}


