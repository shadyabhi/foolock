# foolock

A simple HTTP-based distributed lock service for coordinating file access across devices sharing an iCloud folder (or similar).

## API

Short summary of API, best to look at `hurl` tests for a complete E2E test.

```bash
# Acquire a lock
POST /lock?client=laptop1&ttl=30s

# Renew a lock (same endpoint, same client)
POST /lock?client=laptop1&ttl=30s

# Release a lock
DELETE /lock?client=laptop1

# Check lock status
GET /lock
```

## Example

```bash
# Acquire lock
POST http://localhost:8080/lock?client=laptop1&ttl=10s
HTTP 200
# Response: {"holder": "laptop1", "expires_at": "..."}

# Check status
GET http://localhost:8080/lock
HTTP 200
# Response: {"holder": "laptop1", "expires_at": "...", "is_expired": false}

# Release lock
DELETE http://localhost:8080/lock?client=laptop1
HTTP 200
# Response: {"message": "lock released"}
```

## How it works

- **Client identification**
  - Every client must have a unique identifier (e.g., `laptop1`, `macmini`, or a hardware UUID)
  - This identifier tracks lock ownership and must be consistent across requests

- **Lock lifecycle**
  - Acquire: `POST /lock?client=<id>&ttl=<duration>` - becomes holder if lock is free or expired
  - Renew: same endpoint extends the expiration time (only current holder)
  - Release: `DELETE /lock?client=<id>` - explicitly release when done (only current holder)
  - Locks automatically expire after their TTL

- **Grace period (sticky locks)**
  - After a lock expires, only the previous holder can reclaim it for 5 seconds
  - Prevents lock thrashing when a client temporarily loses connectivity

- **Best practice**
  - Acquire lock → do work → release lock
  - Always release explicitly rather than letting locks expire
