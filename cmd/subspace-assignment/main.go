package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	automationMocknet "github.com/aywhoosh/subspace-assignment-ayush/internal/automation/mocknet"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/app"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/browser"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/config"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/logging"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/storage"
	"github.com/aywhoosh/subspace-assignment-ayush/mocknet"
)

func main() {
	ctx := context.Background()

	// Config is loaded so future commits can rely on it.
	// For now we keep UX minimal: use defaults + env overrides.
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	log := logging.New(logging.Config{Level: cfg.Logging.Level, JSON: cfg.Logging.JSON})
	runID := app.NewRunID()
	ctx = app.WithRunID(ctx, runID)
	ctx = app.WithLogger(ctx, log.With().Str("run_id", runID).Logger())

	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Println("subspace-assignment (scaffold)\n\nThis repository will implement an educational Rod-based automation PoC that ONLY runs against the included local Mock Social Network.\n\nNext: run `make lint` and `make test`, then follow README Quickstart.")
		return
	}

	// Minimal command surface for now.
	if len(args) >= 2 && args[0] == "mocknet" && args[1] == "up" {
		l := app.Logger(ctx, log)
		(&l).Info().Int("port", cfg.Mocknet.Port).Msg("starting mocknet")

		srv, err := mocknet.New(mocknet.Config{
			Port:              cfg.Mocknet.Port,
			CheckpointEnabled: false,
			SeedCredentials: mocknet.Credentials{
				Username: firstNonEmpty(cfg.Auth.Username, os.Getenv("SUBSPACE_AUTH_USERNAME"), "demo"),
				Password: firstNonEmpty(cfg.Auth.Password, os.Getenv("SUBSPACE_AUTH_PASSWORD"), "demo"),
			},
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Println("MockNet running at:", srv.BaseURL())
		fmt.Println("Login at:", srv.BaseURL()+"/login")

		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
		_ = srv.Run(ctx)
		return
	}

	if len(args) >= 2 && args[0] == "automate" && args[1] == "login" {
		l := app.Logger(ctx, log)
		(&l).Info().Str("base_url", cfg.Mocknet.BaseURL).Msg("starting automation login")

		db, repos, err := openRepos(ctx, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer func() { _ = db.Close() }()

		br, cleanup, err := browser.New(ctx, browser.Config{Headless: cfg.Browser.Headless, SlowMo: cfg.Browser.SlowMo})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer func() { _ = cleanup() }()

		username, err := automationMocknet.EnsureAuthed(ctx, br, repos, cfg.Mocknet.BaseURL, automationMocknet.Credentials{
			Username: cfg.Auth.Username,
			Password: cfg.Auth.Password,
		}, automationMocknet.Options{Timeout: cfg.Run.Timeout})
		if err != nil {
			if errors.Is(err, browser.ErrCheckpoint) {
				fmt.Fprintln(os.Stderr, "Checkpoint detected. Open the browser window, complete the checkpoint manually, then re-run `automate login`.")
				os.Exit(3)
			}
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Println("Authenticated as:", username)
		fmt.Println("Session saved key:", automationMocknet.SessionKey(cfg.Mocknet.BaseURL, firstNonEmpty(cfg.Auth.Username, "demo")))
		return
	}

	if len(args) >= 2 && args[0] == "automate" && args[1] == "check" {
		db, repos, err := openRepos(ctx, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer func() { _ = db.Close() }()

		br, cleanup, err := browser.New(ctx, browser.Config{Headless: cfg.Browser.Headless, SlowMo: cfg.Browser.SlowMo})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer func() { _ = cleanup() }()

		username, err := automationMocknet.EnsureAuthed(ctx, br, repos, cfg.Mocknet.BaseURL, automationMocknet.Credentials{
			Username: cfg.Auth.Username,
			Password: cfg.Auth.Password,
		}, automationMocknet.Options{Timeout: cfg.Run.Timeout})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println("Session is valid. Authenticated as:", username)
		return
	}

	l := app.Logger(ctx, log)
	(&l).Info().Msg("subspace-assignment starting")
	fmt.Println("subspace-assignment scaffold. Try: `go run ./cmd/subspace-assignment --help`")
}

func openRepos(ctx context.Context, cfg config.Config) (*storage.DB, *storage.Repositories, error) {
	db, err := storage.Open(ctx, storage.OpenOptions{Path: cfg.Storage.SQLitePath})
	if err != nil {
		return nil, nil, err
	}
	if err := storage.Migrate(ctx, db.SQL()); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return db, storage.NewRepositories(db.SQL()), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v2 := strings.TrimSpace(v); v2 != "" {
			return v2
		}
	}
	return ""
}
