package lockstate

import (
	"sync"
	"time"
)

const (
	DefaultTTL  = 30 * time.Second
	GracePeriod = 5 * time.Second
)

type LockState struct {
	mu         sync.Mutex
	Holder     string
	ExpiresAt  time.Time
	GraceUntil time.Time
}

func New() *LockState {
	return &LockState{}
}

func (ls *LockState) IsExpired() bool {
	return time.Now().After(ls.ExpiresAt) || ls.Holder == ""
}

func (ls *LockState) InGracePeriod() bool {
	now := time.Now()
	return now.After(ls.ExpiresAt) && now.Before(ls.GraceUntil)
}

type AcquireResult struct {
	Success    bool
	Holder     string
	ExpiresAt  time.Time
	GraceUntil time.Time
	Message    string
}

func (ls *LockState) Acquire(client string, ttl time.Duration) AcquireResult {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	now := time.Now()

	// Case 1: Current holder is renewing
	if ls.Holder == client {
		ls.ExpiresAt = now.Add(ttl)
		ls.GraceUntil = now.Add(ttl).Add(GracePeriod)
		return AcquireResult{
			Success:   true,
			Holder:    ls.Holder,
			ExpiresAt: ls.ExpiresAt,
			Message:   "renewed",
		}
	}

	// Case 2: Lock is held by someone else and not expired
	if ls.Holder != "" && now.Before(ls.ExpiresAt) {
		return AcquireResult{
			Success:   false,
			Holder:    ls.Holder,
			ExpiresAt: ls.ExpiresAt,
			Message:   "held by another client",
		}
	}

	// Case 3: Lock is expired but within grace period
	if ls.Holder != "" && now.After(ls.ExpiresAt) && now.Before(ls.GraceUntil) {
		return AcquireResult{
			Success:    false,
			Holder:     ls.Holder,
			ExpiresAt:  ls.ExpiresAt,
			GraceUntil: ls.GraceUntil,
			Message:    "grace period active",
		}
	}

	// Case 4: Lock is available (expired and past grace period, or never held)
	previousHolder := ls.Holder
	ls.Holder = client
	ls.ExpiresAt = now.Add(ttl)
	ls.GraceUntil = now.Add(ttl).Add(GracePeriod)

	message := "acquired"
	if previousHolder != "" {
		message = "acquired from " + previousHolder
	}

	return AcquireResult{
		Success:   true,
		Holder:    ls.Holder,
		ExpiresAt: ls.ExpiresAt,
		Message:   message,
	}
}

type ReleaseResult struct {
	Success bool
	Message string
}

func (ls *LockState) Release(client string) ReleaseResult {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.Holder != client {
		return ReleaseResult{
			Success: false,
			Message: "client does not hold the lock",
		}
	}

	ls.Holder = ""
	ls.ExpiresAt = time.Time{}
	ls.GraceUntil = time.Time{}

	return ReleaseResult{
		Success: true,
		Message: "lock released",
	}
}

type StatusResult struct {
	Holder     string
	ExpiresAt  time.Time
	GraceUntil time.Time
	IsExpired  bool
	InGrace    bool
}

func (ls *LockState) Status() StatusResult {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	return StatusResult{
		Holder:     ls.Holder,
		ExpiresAt:  ls.ExpiresAt,
		GraceUntil: ls.GraceUntil,
		IsExpired:  ls.IsExpired(),
		InGrace:    ls.InGracePeriod(),
	}
}
