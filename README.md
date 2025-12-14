# Subspace Assignment (Educational Browser Automation PoC)

This repository will become an educational proof-of-concept for **browser automation realism** and **clean architecture** using:
- Go
- Rod (browser automation)
- A locally hosted **Mock Social Network** web app included in this repo
- SQLite persistence

## What this is
- A learning-oriented automation project showing robust patterns (timeouts, retries, state persistence, logging)
- A local-only demo that automates a mock site you control

## What this is NOT
- Not a tool for automating real third-party websites/services
- Not a bot-evasion project: **no fingerprint masking, no webdriver-flag tampering, no captcha/2FA bypass**

## Safety / Ethical use
This project will **only** run against the included local mock app on `http://localhost:<port>`.
Any mock “checkpoint” pages (captcha/2FA) will be handled by **detection + safe pause for manual intervention**.

## Quickstart (scaffold)
Prereqs:
- Go (from `go.mod`)
- `golangci-lint` (for `make lint`)

Config inputs (added in commit 2):
- YAML: `config.example.yaml` (copy to `config.yaml` later)
- Env overrides: see `.env.example` (not committed as `.env`)

Storage (added in commit 4):
- SQLite database path is configured via `storage.sqlite_path` or `SUBSPACE_STORAGE_SQLITE_PATH`

Commands:
- `make test`
- `make lint`
- `make run`

Run the local mock app:
- `go run ./cmd/subspace-assignment mocknet up`

Windows note:
- If you don’t have `make`, run the equivalent commands directly:
  - `go test ./...`
  - `golangci-lint run`
  - `go run ./cmd/subspace-assignment --help`

## Roadmap
We will add (in small commits): config, logging, SQLite storage, mock server + pages with stable selectors, Rod session wrapper, human-like interactions, workflows, templates, integration tests, and docs.
