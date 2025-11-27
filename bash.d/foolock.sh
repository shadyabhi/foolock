#!/bin/bash
#
# foolock.sh - Bash library for interacting with foolock server
#
# Usage:
#   source foolock.sh
#   foolock_acquire [ttl]    # Acquire lock with optional TTL (default: 30s)
#   foolock_release          # Release the lock
#   foolock_status           # Check lock status
#
# Environment variables:
#   FOOLOCK_SERVER  - Lock server URL (default: http://localhost:8080)
#

# Default server URL
FOOLOCK_SERVER="${FOOLOCK_SERVER:-http://localhost:8080}"

# Get the client ID from hardware UUID
_foolock_get_client_id() {
    system_profiler SPHardwareDataType | awk '/UUID/ { print $3 }'
}

# Acquire the lock
# Arguments:
#   $1 - TTL (optional, default: 30s, e.g., "10s", "5m", "1h")
# Returns:
#   0 on success, 1 on failure
# Outputs:
#   JSON response from server
foolock_acquire() {
    local ttl="${1:-30s}"
    local client_id
    client_id=$(_foolock_get_client_id)

    if [[ -z "$client_id" ]]; then
        echo '{"error": "failed to get client ID"}' >&2
        return 1
    fi

    local response
    local http_code

    response=$(curl -s -w "\n%{http_code}" -X POST \
        "${FOOLOCK_SERVER}/lock?client=${client_id}&ttl=${ttl}")

    http_code=$(echo "$response" | tail -n1)
    response=$(echo "$response" | sed '$d')

    echo "$response"

    if [[ "$http_code" == "200" ]]; then
        return 0
    else
        return 1
    fi
}

# Release the lock
# Returns:
#   0 on success, 1 on failure
# Outputs:
#   JSON response from server
foolock_release() {
    local client_id
    client_id=$(_foolock_get_client_id)

    if [[ -z "$client_id" ]]; then
        echo '{"error": "failed to get client ID"}' >&2
        return 1
    fi

    local response
    local http_code

    response=$(curl -s -w "\n%{http_code}" -X DELETE \
        "${FOOLOCK_SERVER}/lock?client=${client_id}")

    http_code=$(echo "$response" | tail -n1)
    response=$(echo "$response" | sed '$d')

    echo "$response"

    if [[ "$http_code" == "200" ]]; then
        return 0
    else
        return 1
    fi
}

# Check lock status
# Returns:
#   0 always (status check doesn't fail)
# Outputs:
#   JSON response from server
foolock_status() {
    curl -s -X GET "${FOOLOCK_SERVER}/lock"
}

# Get current client ID (useful for debugging)
foolock_client_id() {
    _foolock_get_client_id
}
