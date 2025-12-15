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

Open:
- `http://localhost:8080/login`

Run Rod automation against the local mock app (persists cookies to SQLite):
- `go run ./cmd/subspace-assignment automate login`
- `go run ./cmd/subspace-assignment automate check`

Notes:
- Stop the mock server with `Ctrl+C`.
- The optional logo is served from your local `./logos` folder at `/brand/logo.png` (assets are intentionally ignored by git).

Windows note:
- If you don’t have `make`, run the equivalent commands directly:
  - `go test ./...`
  - `golangci-lint run`
  - `go run ./cmd/subspace-assignment --help`

Rod on Windows note (Defender / AV):
- This project prefers using a system-installed browser (Edge/Chrome) instead of downloading a bundled Chromium.
- If automation fails to launch a browser, set `SUBSPACE_BROWSER__BIN_PATH` to your browser executable.
- If your machine blocks Rod’s leakless helper, keep `SUBSPACE_BROWSER__LEAKLESS=false` (default).
- If you explicitly want Rod to download Chromium, set `SUBSPACE_BROWSER__ALLOW_DOWNLOAD=true`.

## Roadmap
We will add (in small commits): config, logging, SQLite storage, mock server + pages with stable selectors, Rod session wrapper, human-like interactions, workflows, templates, integration tests, and docs.
