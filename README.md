# foolock

[![Go](https://github.com/shadyabhi/foolock/actions/workflows/go.yml/badge.svg)](https://github.com/shadyabhi/foolock/actions/workflows/go.yml)
[![Coverage](https://raw.githubusercontent.com/shadyabhi/foolock/badges/.badges/main/coverage.svg)](https://github.com/shadyabhi/foolock/actions/workflows/go.yml)

A simple HTTP-based distributed lock service for coordinating file access across devices sharing an iCloud folder (or similar).

## API

Short summary of API, best to look at `hurl` tests for a complete E2E test.

```bash
# Acquire a lock for a job (job defaults to "default" if not specified)
POST /lock?client=laptop1&job=myjob&ttl=30s

# Renew a lock (same endpoint, same client, same job)
POST /lock?client=laptop1&job=myjob&ttl=30s

# Release a lock
DELETE /lock?client=laptop1&job=myjob

# Check lock status for a job
GET /lock?job=myjob
```

## Example

```bash
# Acquire lock for job "backup"
POST http://localhost:8080/lock?client=laptop1&job=backup&ttl=10s
HTTP 200
# Response: {"holder": "laptop1", "job": "backup", "expires_at": "..."}

# Check status for job "backup"
GET http://localhost:8080/lock?job=backup
HTTP 200
# Response: {"holder": "laptop1", "job": "backup", "expires_at": "...", "is_expired": false}

# Release lock for job "backup"
DELETE http://localhost:8080/lock?client=laptop1&job=backup
HTTP 200
# Response: {"message": "lock released", "job": "backup"}

# Different jobs are independent - laptop2 can lock "sync" while laptop1 holds "backup"
POST http://localhost:8080/lock?client=laptop2&job=sync&ttl=10s
HTTP 200
# Response: {"holder": "laptop2", "job": "sync", "expires_at": "..."}
```

## How it works

- **Client identification**
  - Every client must have a unique identifier (e.g., `laptop1`, `macmini`, or a hardware UUID)
  - This identifier tracks lock ownership and must be consistent across requests

- **Job-based locking**
  - Locks are scoped per job - the combination of `(job, client)` uniquely identifies a lock holder
  - Different jobs are completely independent - one client can hold locks on multiple jobs
  - If `job` is not specified, it defaults to `"default"`

- **Lock lifecycle**
  - Acquire: `POST /lock?client=<id>&job=<name>&ttl=<duration>` - becomes holder if lock is free or expired
  - Renew: same endpoint extends the expiration time (only current holder)
  - Release: `DELETE /lock?client=<id>&job=<name>` - explicitly release when done (only current holder)
  - Locks automatically expire after their TTL

- **Grace period (sticky locks)**
  - After a lock expires, only the previous holder can reclaim it for 5 seconds
  - Prevents lock thrashing when a client temporarily loses connectivity

- **Best practice**
  - Acquire lock → do work → release lock
  - Always release explicitly rather than letting locks expire
