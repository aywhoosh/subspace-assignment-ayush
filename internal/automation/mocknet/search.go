package mocknet

import (
	"context"
	"fmt"
	"time"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
	"github.com/go-rod/rod"
)

// SearchOptions configures search behavior
type SearchOptions struct {
	Title    string
	Company  string
	Location string
	Keywords string
	PerPage  int
}

// SearchResult represents a single search result
type SearchResult struct {
	ProfileID string
	Name      string
	Title     string
	Company   string
	Location  string
}

// Search performs a people search and returns results
func Search(ctx context.Context, br *browser.Client, baseURL string, opts SearchOptions) ([]SearchResult, error) {
	page, err := br.NewPage(baseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("search: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return nil, fmt.Errorf("search: wait load: %w", err)
	}

	// Fill search form if criteria provided
	if opts.Title != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-title-input']", opts.Title); err != nil {
			return nil, err
		}
	}
	if opts.Company != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-company-input']", opts.Company); err != nil {
			return nil, err
		}
	}
	if opts.Location != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-location-input']", opts.Location); err != nil {
			return nil, err
		}
	}
	if opts.Keywords != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-keywords-input']", opts.Keywords); err != nil {
			return nil, err
		}
	}

	// Click search if any criteria was provided
	if opts.Title != "" || opts.Company != "" || opts.Location != "" || opts.Keywords != "" {
		wait := page.MustWaitNavigation()
		if err := click(page, "[data-testid='search-submit']"); err != nil {
			return nil, err
		}
		wait()
		if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
			return nil, fmt.Errorf("search: wait after submit: %w", err)
		}
	}

	// Parse results
	resultsContainer, err := page.Timeout(5 * time.Second).Element("[data-testid='search-results']")
	if err != nil {
		return nil, fmt.Errorf("search: results container not found: %w", err)
	}

	resultElements, err := resultsContainer.Elements("[data-testid='search-result']")
	if err != nil {
		return []SearchResult{}, nil // No results is ok
	}

	var results []SearchResult
	for _, el := range resultElements {
		profileID, _ := el.Attribute("data-profile-id")
		if profileID == nil {
			continue
		}

		nameEl, err := el.Element("[data-testid='search-result-name']")
		if err != nil {
			continue
		}
		name, _ := nameEl.Text()

		metaEl, err := el.Element("[data-testid='search-result-meta']")
		if err != nil {
			continue
		}
		meta, _ := metaEl.Text()

		results = append(results, SearchResult{
			ProfileID: *profileID,
			Name:      name,
			Title:     meta, // simplified - meta contains "Title • Company • Location"
		})
	}

	return results, nil
}

// ViewProfile navigates to a profile page and returns basic info
func ViewProfile(ctx context.Context, br *browser.Client, baseURL, profileID string) (string, error) {
	page, err := br.NewPage(baseURL + "/profile/" + profileID)
	if err != nil {
		return "", fmt.Errorf("view profile: new page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return "", fmt.Errorf("view profile: wait load: %w", err)
	}

	// Get profile name as confirmation
	nameEl, err := page.Timeout(5 * time.Second).Element("[data-testid='profile-name']")
	if err != nil {
		return "", fmt.Errorf("view profile: name not found: %w", err)
	}

	name, err := nameEl.Text()
	if err != nil {
		return "", fmt.Errorf("view profile: read name: %w", err)
	}

	return name, nil
}

func typeIntoIfExists(page *rod.Page, selector, value string) error {
	el, err := page.Timeout(2 * time.Second).Element(selector)
	if err != nil {
		return nil // Element not found is ok
	}
	if err := el.Input(value); err != nil {
		return fmt.Errorf("automation: input %s: %w", selector, err)
	}
	return nil
}

// SearchOnPage performs search on an already-open page (for interactive mode)
func SearchOnPage(ctx context.Context, page *rod.Page, opts SearchOptions) ([]SearchResult, error) {
	// Page should already be on /search, just fill the form
	if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
		return nil, fmt.Errorf("search: wait load: %w", err)
	}

	// Fill search form if criteria provided
	if opts.Title != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-title-input']", opts.Title); err != nil {
			return nil, err
		}
	}
	if opts.Company != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-company-input']", opts.Company); err != nil {
			return nil, err
		}
	}
	if opts.Location != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-location-input']", opts.Location); err != nil {
			return nil, err
		}
	}
	if opts.Keywords != "" {
		if err := typeIntoIfExists(page, "[data-testid='search-keywords-input']", opts.Keywords); err != nil {
			return nil, err
		}
	}

	// Click search if any criteria was provided
	if opts.Title != "" || opts.Company != "" || opts.Location != "" || opts.Keywords != "" {
		wait := page.MustWaitNavigation()
		if err := click(page, "[data-testid='search-submit']"); err != nil {
			return nil, err
		}
		wait()
		if err := page.Timeout(10 * time.Second).WaitLoad(); err != nil {
			return nil, fmt.Errorf("search: wait after submit: %w", err)
		}
	}

	// Parse results
	resultsContainer, err := page.Timeout(5 * time.Second).Element("[data-testid='search-results']")
	if err != nil {
		return nil, fmt.Errorf("search: results container not found: %w", err)
	}

	resultElements, err := resultsContainer.Elements("[data-testid='search-result']")
	if err != nil {
		return []SearchResult{}, nil // No results is ok
	}

	var results []SearchResult
	for _, el := range resultElements {
		profileID, _ := el.Attribute("data-profile-id")
		if profileID == nil {
			continue
		}

		nameEl, err := el.Element("[data-testid='search-result-name']")
		if err != nil {
			continue
		}
		name, _ := nameEl.Text()

		metaEl, err := el.Element("[data-testid='search-result-meta']")
		if err != nil {
			continue
		}
		meta, _ := metaEl.Text()

		results = append(results, SearchResult{
			ProfileID: *profileID,
			Name:      name,
			Title:     meta,
		})
	}

	return results, nil
}
