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

## Implemented Features

✅ Configuration system (YAML + env vars)
✅ Structured logging with slog
✅ SQLite storage with migrations
✅ Mock server with stable test selectors
✅ Rod browser automation client
✅ Cookie-based session persistence
✅ Authentication workflows (login, session check)
✅ Search & profile viewing workflows
✅ Connection request workflows
✅ Messaging workflows
✅ Human-like behavior patterns (delays, typing speed)
✅ Unit test coverage
✅ Comprehensive documentation

## API Reference

### Search Workflows

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

### Connection Workflows

```go
func SendConnectionRequest(ctx, br, baseURL, profileID, note) error
func GetPendingRequests(ctx, br, baseURL) ([]string, error)
```

### Messaging Workflows

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

### Human-like Behaviors

```go
type HumanBehavior struct {
    RandomDelayMin     time.Duration
    RandomDelayMax     time.Duration
    TypingSpeedMin     time.Duration
    TypingSpeedMax     time.Duration
}

func DefaultHumanBehavior() HumanBehavior  // Realistic delays (500-2000ms)
func FastHumanBehavior() HumanBehavior      // Faster but realistic (200-800ms)
func (h HumanBehavior) RandomDelay(ctx)
func (h HumanBehavior) HumanType(ctx, el, text) error
func (h HumanBehavior) HumanClick(ctx, page, el) error
func (h HumanBehavior) ScrollRandom(ctx, page) error
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Test specific package
go test ./internal/automation/mocknet -v
```

Test coverage includes:
- ✅ Automation workflows (search, connect, message)
- ✅ Human-like behavior patterns
- ✅ Browser cookie handling
- ✅ Configuration loading
- ✅ Database migrations

## Troubleshooting

### Browser Detection Issues

```bash
# Validate browser setup
go run ./cmd/subspace-assignment automate doctor

# Set explicit browser path
export SUBSPACE_BROWSER__BIN_PATH="/path/to/browser"
```

### Timeout Errors

All element lookups have 5-10 second timeouts. If operations fail:
- Verify mock server is running: `http://localhost:8080`
- Check network connectivity
- Increase timeout in code if needed

### Session Issues

```bash
# Clear saved sessions
rm data/subspace.db

# Login fresh
go run ./cmd/subspace-assignment automate login
```

## Performance

- **Login**: ~2-3 seconds (with navigation waits)
- **Search**: ~1-2 seconds (depends on results)
- **Session Check**: ~1 second (cookie reuse)
- **Message Send**: ~1 second

With `FastHumanBehavior()`:
- Delays: 200-800ms (vs 500-2000ms default)
- Typing: 20-80ms/char (vs 50-150ms default)

## Database Schema

### Sessions Table
```sql
CREATE TABLE sessions (
    key      TEXT PRIMARY KEY,  -- Format: "mocknet|baseURL|username"
    cookies  TEXT NOT NULL,     -- JSON array of cookies
    created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Security Notes

- ⚠️ Browser runs in **non-leakless mode** for development
- 🔒 Cookies stored in local SQLite (plaintext)
- 🚫 No cloud sync or external transmissions
- ⚠️ Mock server has no authentication (localhost only)

**Production Considerations:**
- Enable leakless mode for cleanup
- Encrypt cookie storage
- Use secure credential management
- Implement rate limiting

## Support

- Check [commands.txt](./commands.txt) for quick reference
- Review test files for usage examples
- Rod documentation: https://go-rod.github.io/
