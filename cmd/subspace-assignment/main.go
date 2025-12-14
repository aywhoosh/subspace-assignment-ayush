package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/app"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/config"
	"github.com/aywhoosh/subspace-assignment-ayush/internal/logging"
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

	l := app.Logger(ctx, log)
	(&l).Info().Msg("subspace-assignment starting")
	fmt.Println("subspace-assignment scaffold. Try: `go run ./cmd/subspace-assignment --help`")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v2 := strings.TrimSpace(v); v2 != "" {
			return v2
		}
	}
	return ""
}
