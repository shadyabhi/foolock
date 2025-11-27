# CLAUDE.md - AI Assistant Guide for foolock

## Quick Reference

```bash
make build    # Build: go build ./...
make test     # Unit tests with race detection and coverage
make hurl     # E2E tests (starts server, runs hurl, stops server)
go run . &    # Start server on :8080
```

## Project Structure

```
foolock/
├── main.go              # Entry point
├── lockstate/           # Core lock logic (acquire.go, release.go, lockstate.go)
│   └── msg/msg.go       # Message constants
├── lockstatehttp/       # HTTP handlers
├── hurl/                # E2E tests
└── bash.d/              # Bash client library (macOS)
```

## Guidelines

### Code Style
- Keep dependencies minimal - only stdlib for runtime, testify for tests
- All user-facing messages go in `lockstate/msg/msg.go`
- Lock operations return typed result structs (`AcquireResult`, `ReleaseResult`, `StatusResult`)
- JSON encoding errors are logged but don't fail responses

### HTTP Status Codes
- 200: Success
- 400: Bad Request (missing client, invalid TTL)
- 403: Forbidden (release by non-holder)
- 409: Conflict (lock held, grace period active)
- 405: Method Not Allowed

### Testing Requirements
- Write table-driven tests with descriptive names
- Use setup functions to configure initial state
- Use `httptest.NewRecorder` for handler tests
- Run `make test` before committing
- Run `make hurl` if API behavior changes

### Adding New Features
1. Add logic to `lockstate/` package with result type
2. Add HTTP handler in `lockstatehttp/handlers.go`
3. Add message constants in `lockstate/msg/msg.go`
4. Write unit tests for both packages
5. Add hurl E2E test if endpoint behavior changes

### Commit Messages
Use conventional format: `feat:`, `fix:`, `docs:`, `test:`, `chore:`

## API Summary

Single `/lock` endpoint:
- `POST /lock?client=X&ttl=30s` - Acquire/renew lock
- `DELETE /lock?client=X` - Release lock
- `GET /lock` - Check status
