# CLAUDE.md - AI Assistant Guide for foolock

## Project Overview

**foolock** is a simple HTTP-based distributed lock service written in Go, designed for coordinating file access across multiple devices sharing an iCloud-synced folder (or similar). It implements lease-based locking with a grace period mechanism to prevent lock thrashing.

## Quick Reference

```bash
# Build
make build          # or: go build ./...

# Run tests
make test           # Unit tests with race detection and coverage

# Run E2E tests (starts server, runs hurl tests, stops server)
make hurl

# Start server manually
go run . &          # Runs on :8080
```

## Architecture

```
foolock/
├── main.go                    # Entry point - HTTP server setup
├── lockstate/                 # Core lock state management
│   ├── lockstate.go          # LockState struct, Status(), options
│   ├── acquire.go            # Lock acquisition logic
│   ├── release.go            # Lock release logic
│   └── msg/msg.go            # Message constants
├── lockstatehttp/            # HTTP layer
│   └── handlers.go           # HTTP handlers for /lock endpoint
├── hurl/                     # E2E tests using hurl
└── bash.d/                   # Bash client library (macOS)
```

### Key Components

1. **lockstate.LockState** (`lockstate/lockstate.go:13-22`)
   - Thread-safe state management with `sync.Mutex`
   - Holds: `Holder`, `AcquiredAt`, `ExpiresAt`, `GraceUntil`
   - Configurable via functional options: `WithTTL()`, `WithGracePeriod()`

2. **lockstatehttp.Handler** (`lockstatehttp/handlers.go:26-32`)
   - Single `/lock` endpoint handling GET/POST/DELETE
   - Returns JSON responses with appropriate HTTP status codes

## API Endpoints

All operations use the `/lock` endpoint:

| Method | Purpose | Query Params | Success | Error |
|--------|---------|--------------|---------|-------|
| `POST` | Acquire/renew lock | `client` (required), `ttl` (optional, default: 30s) | 200 | 409 Conflict |
| `DELETE` | Release lock | `client` (required) | 200 | 403 Forbidden |
| `GET` | Check status | None | 200 | - |

## Constants

Defined in `lockstate/lockstate.go:8-11`:
- Default TTL: 30 seconds
- Grace period: 5 seconds
- Server address: `:8080` (in `main.go:11`)

## Lock Behavior

1. **Acquire**: Client gets lock if free or expired past grace period
2. **Renew**: Current holder can extend TTL with same POST request
3. **Grace Period**: After expiration, only previous holder can reclaim for 5s
4. **Release**: Explicit release clears lock immediately

## Testing

### Unit Tests
```bash
go test -v -race -count 1 -cover ./...
```

Tests use:
- Standard `testing` package with table-driven tests
- `github.com/stretchr/testify/require` for HTTP handler tests
- `httptest.NewRecorder` for handler testing

### E2E Tests
```bash
make hurl  # Full E2E test cycle
```

Uses [hurl](https://hurl.dev/) for HTTP testing. Tests are in `hurl/*.hurl` and run sequentially (`--jobs 1`).

## Code Conventions

### Error Handling
- HTTP handlers return JSON error responses: `{"error": "message"}`
- All JSON encoding errors are logged but don't fail the response
- Lock operation results use typed structs: `AcquireResult`, `ReleaseResult`, `StatusResult`

### Message Constants
All user-facing messages are defined in `lockstate/msg/msg.go`:
- `Acquired`, `Renewed`, `HeldByAnother`, `GracePeriodActive`
- `LockReleased`, `ClientNotHolder`, `LockHeld`, `NoLockHeld`

### HTTP Status Codes
- 200: Success (acquire, renew, release, status)
- 400: Bad Request (missing client, invalid TTL)
- 403: Forbidden (release by non-holder)
- 409: Conflict (lock held by another, grace period active)
- 405: Method Not Allowed

### Test Patterns
- Table-driven tests with descriptive names
- Setup functions to configure initial state
- Check functions for response validation

## CI/CD

### GitHub Actions (`.github/workflows/`)

1. **go.yml** - Runs on push/PR to main:
   - Build: `go build -v ./...`
   - Test: `go test -v ./... -coverprofile=./cover.out`
   - Coverage badge generation

2. **release.yml** - Runs on tags:
   - GoReleaser for multi-platform builds
   - Docker image push to Docker Hub

3. **dev-release.yml** - Runs on push to main:
   - Snapshot release for development builds

### Release Configuration

GoReleaser (`.goreleaser.yaml`) builds:
- Linux, Windows, Darwin (amd64, arm64)
- Universal binaries for macOS
- Docker images (amd64, arm64)
- Homebrew formula (shadyabhi/homebrew-tap)

## Development Workflow

1. Make changes in relevant package
2. Run `make test` to verify unit tests pass
3. Run `make hurl` for E2E validation (if API changes)
4. Commit with conventional message format (feat:, fix:, docs:, test:, chore:)

## Dependencies

Minimal dependencies (from `go.mod`):
- `github.com/stretchr/testify` - Testing assertions (test only)
- No runtime dependencies beyond Go stdlib

## Common Tasks

### Adding a New Lock Operation
1. Add method to `LockState` in `lockstate/` with corresponding result type
2. Add HTTP handler method in `lockstatehttp/handlers.go`
3. Add message constant in `lockstate/msg/msg.go`
4. Write unit tests for both lockstate and handler
5. Add hurl E2E test

### Modifying Lock Behavior
1. Update logic in `lockstate/acquire.go` or `lockstate/release.go`
2. Update tests in corresponding `*_test.go` files
3. Verify E2E tests still pass with `make hurl`

### Testing Grace Period Changes
Grace period behavior is tested in:
- `hurl/11_grace_period_blocks_others.hurl`
- `hurl/12_grace_period_expires.hurl`

Use `WithGracePeriod()` option for unit testing with custom values.
