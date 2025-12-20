# Subspace Assignment — Browser Automation (MockNet)

[![CI](https://github.com/aywhoosh/subspace-assignment-ayush/actions/workflows/ci.yml/badge.svg)](https://github.com/aywhoosh/subspace-assignment-ayush/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](go.mod)

Educational proof-of-concept demonstrating **robust browser automation** and **clean architecture** in Go using **Rod**, against a **local mock social network** shipped with this repository.

## What this is / isn’t

- ✅ A local-only automation demo with resilient patterns (timeouts, state persistence, logging)
- ✅ A safe, controlled environment: `mocknet` runs at `http://localhost:8080`
- ❌ Not for automating real third‑party websites/services
- ❌ Not a bot‑evasion project (no fingerprint masking, webdriver flag tampering, CAPTCHA/2FA bypass)

## Screenshots (placeholders)

Add images into a `./screenshots/` folder and update the links below.

- MockNet login: `screenshots/mocknet-login.png`
- Interactive CLI menu: `screenshots/interactive-menu.png`
- Search results: `screenshots/search-results.png`
- Messaging workflow: `screenshots/send-message.png`

Example:

```markdown
![Interactive CLI](screenshots/interactive-menu.png)
```

## Quick start (Windows)

### 1) Start MockNet (required)

PowerShell:

```powershell
Start-Job -ScriptBlock { go run ./cmd/subspace-assignment mocknet up }
Start-Sleep 2
```

Open:

- `http://localhost:8080/login`

### 2) Run the interactive demo (recommended)

```powershell
go run ./cmd/subspace-assignment interactive
```

The menu supports:

- Login
- Search (shows profile IDs)
- Send connection request
- Send message
- Check session status
- View inbox

### 3) Run individual automations

```powershell
go run ./cmd/subspace-assignment automate doctor
go run ./cmd/subspace-assignment automate login
go run ./cmd/subspace-assignment automate check
```

### 4) Stop server and clean up

```powershell
Get-Job | Stop-Job
Get-Job | Remove-Job

# If a browser was left open:
Stop-Process -Name msedge -Force -ErrorAction SilentlyContinue
Stop-Process -Name chrome -Force -ErrorAction SilentlyContinue
```

More command snippets: see `commands.txt`.

## Configuration

- Copy `config.example.yaml` → `config.yaml`
- Storage:
    - YAML: `storage.sqlite_path`
    - Env: `SUBSPACE_STORAGE_SQLITE_PATH`

### Browser selection (Windows)

This project prefers a system-installed browser (Edge/Chrome). If auto-detection fails:

```powershell
$env:SUBSPACE_BROWSER__BIN_PATH="C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"
```

If your machine blocks Rod’s helper, keep `SUBSPACE_BROWSER__LEAKLESS=false` (default).

## Features implemented

- Configuration (YAML + env overrides)
- Structured logging (`slog`)
- SQLite storage + migrations
- MockNet server with stable selectors
- Rod browser client + cookie persistence
- Workflows: auth, session check, search/profile view, connect, messaging
- Human-like behavior: delays + typing speed
- Unit tests (`go test ./...`)

## Architecture

```mermaid
graph TD
    CLI[cmd/subspace-assignment] --> APP[internal/app]
    APP --> AUTO[internal/automation]
    AUTO --> BROWSER[internal/browser (Rod)]
    AUTO --> STORE[internal/storage (SQLite)]
    APP --> MOCK[mocknet]
    STORE --> DB[(SQLite DB)]
```

Key directories:

- `cmd/` — CLI entrypoints (interactive + automation commands)
- `internal/automation/` — workflows (auth/search/connect/message/human)
- `internal/browser/` — Rod wrapper + cookie utilities
- `internal/storage/` — repositories + migrations
- `mocknet/` — local web app (templates + handlers)

## API surface (developer reference)

### Search

```go
type SearchOptions struct {
	Title    string
	Company  string
	Location string
	Keywords string
	PerPage  int
}

func Search(ctx, br, baseURL, opts) ([]SearchResult, error)
func ViewProfile(ctx, br, baseURL, profileID) (string, error)
```

### Connections

```go
func SendConnectionRequest(ctx, br, baseURL, profileID, note) error
func GetPendingRequests(ctx, br, baseURL) ([]string, error)
```

### Messaging

```go
type Message struct {
	From      string
	Content   string
	Timestamp string
}

func SendMessage(ctx, br, baseURL, recipientID, messageText) error
func GetConversation(ctx, br, baseURL, recipientID) ([]Message, error)
func GetInbox(ctx, br, baseURL) ([]string, error)
```

## Testing

```powershell
go test ./...
```

## Troubleshooting

### Browser setup

```powershell
go run ./cmd/subspace-assignment automate doctor
```

### Clear local state (fresh run)

PowerShell:

```powershell
Remove-Item -Path .\data\subspace.db -Force -ErrorAction SilentlyContinue
```

## Safety / ethics

- Automation targets **only** the included local MockNet app.
- Checkpoint pages are handled via **detection + safe pause** (no bypass logic).

## Links

- Quick command reference: `commands.txt`
- Rod docs: https://go-rod.github.io/
