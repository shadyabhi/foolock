This are functions you can import to quickly use this lock service.

## Usage

```bash
source foolock.sh

# Acquire lock for a job (default job name is "default")
foolock_acquire                    # Uses default job, 30s TTL
foolock_acquire myjob              # Uses "myjob", 30s TTL
foolock_acquire myjob 60s          # Uses "myjob", 60s TTL

# Release lock for a job
foolock_release                    # Releases default job
foolock_release myjob              # Releases "myjob"

# Check status for a job
foolock_status                     # Status of default job
foolock_status myjob               # Status of "myjob"

# Get client ID (useful for debugging)
foolock_client_id
```

## Environment Variables

- `FOOLOCK_SERVER` - Lock server URL (default: http://localhost:8080)
- `FOOLOCK_JOB` - Default job name (default: "default")
