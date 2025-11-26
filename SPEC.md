# Distributed Lock Service (Homelab)

A simple HTTP-based lock service in Go for coordinating access across multiple laptops sharing an iCloud-synced folder.

## Design Goals

- Single server, in-memory state (no persistence needed)
- Less than 5 clients
- Lock stickiness: avoid bouncing between clients
- Handle client crashes gracefully via TTL expiration

---

## Approach: Lease-based Lock with Grace Period

When a lock expires, the previous holder gets a short exclusive window to reclaim it before it becomes available to others.

---

## API Endpoints

### `POST /lock`

Acquire or renew a lock.

**Query Parameters:**
| Param    | Required | Description                     | Example |
|----------|----------|---------------------------------|---------|
| `client` | Yes      | Unique client identifier        | laptop1 |
| `ttl`    | No       | Lock duration (default: 30s)    | 30s     |

**Responses:**
| Code | Meaning                                      |
|------|----------------------------------------------|
| 200  | Lock acquired/renewed                        |
| 409  | Lock held by another client (or grace period)|

**Response Body:**
```json
{"holder": "laptop1", "expires_at": "2024-01-15T10:30:00Z"}
```

---

### `DELETE /lock`

Explicitly release the lock.

**Query Parameters:**
| Param    | Required | Description              |
|----------|----------|--------------------------|
| `client` | Yes      | Must match current holder|

**Responses:**
| Code | Meaning                          |
|------|----------------------------------|
| 200  | Lock released                    |
| 403  | Client doesn't hold the lock     |

---

### `GET /lock`

Check current lock status.

**Response Body:**
```json
{
  "holder": "laptop1",
  "expires_at": "2024-01-15T10:30:00Z",
  "is_expired": false,
  "grace_until": "2024-01-15T10:30:05Z"
}
```

Or if no lock:
```json
{
  "holder": "",
  "is_expired": true
}
```

---

## Server State

```go
type LockState struct {
    mu          sync.Mutex
    Holder      string
    ExpiresAt   time.Time
    GraceUntil  time.Time
}
```

---

## Lock Acquisition Logic

```
On POST /lock?client=X&ttl=T:

1. If current holder == X:
   → Renew: set ExpiresAt = now + T, GraceUntil = now + T + grace
   → Return 200

2. If lock is held by someone else AND not expired (now < ExpiresAt):
   → Return 409 Conflict

3. If lock is expired (now >= ExpiresAt) but within grace (now < GraceUntil):
   - If X == previous holder → grant lock, return 200
   - Else → return 409 Conflict with message "grace period active"

4. If lock is expired AND past grace period (now >= GraceUntil):
   → Grant lock to X, return 200
```

---

## Constants

```go
const (
    DefaultTTL    = 30 * time.Second
    GracePeriod   = 5 * time.Second
    ServerAddr    = ":8080"
)
```

---

## Client Usage Pattern

Clients should renew well before expiry to maintain stickiness:

```
ttl = 30s
renew_interval = 10s  (renew at 1/3 of TTL)
```

Example client loop:
```
1. POST /lock?client=myid&ttl=30s
2. If 200 → hold lock, do work
3. Every 10s → POST /lock?client=myid&ttl=30s (renew)
4. On shutdown → DELETE /lock?client=myid
```

---

## Timeline Example

```
t=0s    laptop1 acquires lock (ttl=30s, grace=5s)
t=10s   laptop1 renews → lock extended to t=40s
t=20s   laptop1 renews → lock extended to t=50s
t=25s   laptop1 closes lid (stops renewing)
t=50s   lock expires, grace period starts (until t=55s)
t=52s   laptop2 tries POST /lock → 409 (grace period)
t=56s   laptop2 tries POST /lock → 200 (grace expired, lock granted)
```

---

## Implementation Notes

- Single `main.go` file, no external dependencies
- Use `sync.Mutex` for thread safety
- Parse TTL with `time.ParseDuration`
- Return JSON responses with appropriate status codes
- Log lock acquisitions/releases for debugging

---

## Optional Enhancements (Not Required)

- [ ] Configurable grace period via flag
- [ ] Prometheus metrics endpoint
- [ ] Simple web UI showing lock status
