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
					Key:        key,
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
		Key:        key,
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

	if timeout > 0 {
		page = page.Timeout(timeout)
	}

	if err := page.WaitLoad(); err != nil {
		return "", "", fmt.Errorf("automation: wait load: %w", err)
	}

	if err := typeInto(page, "[data-testid='login-username']", creds.Username); err != nil {
		return "", "", err
	}
	if err := typeInto(page, "[data-testid='login-password']", creds.Password); err != nil {
		return "", "", err
	}
	if err := click(page, "[data-testid='login-submit']"); err != nil {
		return "", "", err
	}
	_ = page.WaitLoad()

	if isCheckpoint(page) {
		return "", "", browser.ErrCheckpoint
	}

	// If we landed somewhere other than /search, go there to validate.
	_ = page.Navigate(baseURL + "/search")
	_ = page.WaitLoad()

	u, ok := readText(page, "[data-testid='nav-user']")
	if !ok || strings.TrimSpace(u) == "" {
		return "", "", errors.New("automation: login failed (not authenticated)")
	}

	cookies, err := browser.GetAllCookies(ctx, br.Browser())
	if err != nil {
		return "", "", err
	}
	cookiesJSON, err := browser.CookiesToJSON(cookies)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(u), cookiesJSON, nil
}

func checkAuthed(ctx context.Context, br *browser.Client, baseURL string, timeout time.Duration) (string, bool) {
	page, err := br.NewPage(baseURL + "/search")
	if err != nil {
		return "", false
	}
	defer func() { _ = page.Close() }()
	if timeout > 0 {
		page = page.Timeout(timeout)
	}
	_ = page.WaitLoad()
	if isCheckpoint(page) {
		return "", false
	}
	username, ok := readText(page, "[data-testid='nav-user']")
	if !ok {
		return "", false
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return "", false
	}
	return username, true
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

func isCheckpoint(page *rod.Page) bool {
	_, err := page.Element("[data-testid='checkpoint-card']")
	return err == nil
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

func readText(page *rod.Page, selector string) (string, bool) {
	el, err := page.Element(selector)
	if err != nil {
		return "", false
	}
	t, err := el.Text()
	if err != nil {
		return "", false
	}
	return t, true
}
