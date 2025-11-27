#!/bin/bash
#
# foolock.sh - Bash library for interacting with foolock server
#
# Usage:
#   source foolock.sh
#   foolock_acquire [job] [ttl]    # Acquire lock with optional job and TTL
#   foolock_release [job]          # Release the lock for a job
#   foolock_status [job]           # Check lock status for a job
#
# Environment variables:
#   FOOLOCK_SERVER  - Lock server URL (default: http://localhost:8080)
#   FOOLOCK_JOB     - Default job name (default: "default")
#

# Default server URL
FOOLOCK_SERVER="${FOOLOCK_SERVER:-http://localhost:8080}"

# Default job name
FOOLOCK_JOB="${FOOLOCK_JOB:-default}"

# Get the client ID from computer name and hardware UUID
_foolock_get_client_id() {
    local computer_name
    local hardware_uuid
    computer_name=$(system_profiler SPSoftwareDataType | grep "Computer Name" | cut -d: -f2 | sed 's/[^[:alnum:]+._-]//g')
    hardware_uuid=$(system_profiler SPHardwareDataType | awk '/UUID/ { print $3 }')
    echo "${computer_name}-${hardware_uuid}"
}

# Acquire the lock
# Arguments:
#   $1 - Job name (optional, default: $FOOLOCK_JOB or "default")
#   $2 - TTL (optional, default: 30s, e.g., "10s", "5m", "1h")
# Returns:
#   0 on success, 1 on failure
# Outputs:
#   JSON response from server
foolock_acquire() {
    local job="${1:-$FOOLOCK_JOB}"
    local ttl="${2:-30s}"
    local client_id
    client_id=$(_foolock_get_client_id)

    if [[ -z "$client_id" ]]; then
        echo '{"error": "failed to get client ID"}' >&2
        return 1
    fi

    local response
    local http_code

    response=$(curl -s -w "\n%{http_code}" -X POST \
        "${FOOLOCK_SERVER}/lock?client=${client_id}&job=${job}&ttl=${ttl}")

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
# Arguments:
#   $1 - Job name (optional, default: $FOOLOCK_JOB or "default")
# Returns:
#   0 on success, 1 on failure
# Outputs:
#   JSON response from server
foolock_release() {
    local job="${1:-$FOOLOCK_JOB}"
    local client_id
    client_id=$(_foolock_get_client_id)

    if [[ -z "$client_id" ]]; then
        echo '{"error": "failed to get client ID"}' >&2
        return 1
    fi

    local response
    local http_code

    response=$(curl -s -w "\n%{http_code}" -X DELETE \
        "${FOOLOCK_SERVER}/lock?client=${client_id}&job=${job}")

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
# Arguments:
#   $1 - Job name (optional, default: $FOOLOCK_JOB or "default")
# Returns:
#   0 always (status check doesn't fail)
# Outputs:
#   JSON response from server
foolock_status() {
    local job="${1:-$FOOLOCK_JOB}"
    curl -s -X GET "${FOOLOCK_SERVER}/lock?job=${job}"
}

# Get current client ID (useful for debugging)
foolock_client_id() {
    _foolock_get_client_id
}
