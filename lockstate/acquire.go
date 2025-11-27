package lockstate

import (
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

type AcquireResult struct {
	Success bool
	Holder  string
	Message string

	ExpiresAt  time.Time
	GraceUntil time.Time
}

func (ls *LockState) Acquire(client string, ttl time.Duration) AcquireResult {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	now := time.Now()

	if ls.isCurrentHolderRenewing(client) {
		return ls.respRenewLock(now, ttl)
	}

	if ls.isHeldByAnother(now) {
		return ls.respAlreadyLocked()
	}

	if ls.isInGracePeriod(now) {
		return ls.respActiveGracePeriod()
	}

	return ls.acquireLock(client, now, ttl)
}

func (ls *LockState) isCurrentHolderRenewing(client string) bool {
	return ls.Holder == client
}

func (ls *LockState) respRenewLock(now time.Time, ttl time.Duration) AcquireResult {
	ls.ExpiresAt = now.Add(ttl)
	ls.GraceUntil = ls.ExpiresAt.Add(ls.gracePeriod)
	return AcquireResult{
		Success:   true,
		Holder:    ls.Holder,
		ExpiresAt: ls.ExpiresAt,
		Message:   msg.Renewed,
	}
}

func (ls *LockState) isHeldByAnother(now time.Time) bool {
	return ls.Holder != "" && now.Before(ls.ExpiresAt)
}

func (ls *LockState) respAlreadyLocked() AcquireResult {
	return AcquireResult{
		Success:   false,
		Holder:    ls.Holder,
		ExpiresAt: ls.ExpiresAt,
		Message:   msg.HeldByAnother,
	}
}

func (ls *LockState) isInGracePeriod(now time.Time) bool {
	return ls.Holder != "" && now.After(ls.ExpiresAt) && now.Before(ls.GraceUntil)
}

func (ls *LockState) respActiveGracePeriod() AcquireResult {
	return AcquireResult{
		Success:    false,
		Holder:     ls.Holder,
		ExpiresAt:  ls.ExpiresAt,
		GraceUntil: ls.GraceUntil,
		Message:    msg.GracePeriodActive,
	}
}

func (ls *LockState) acquireLock(client string, now time.Time, ttl time.Duration) AcquireResult {
	previousHolder := ls.Holder
	ls.Holder = client
	ls.ExpiresAt = now.Add(ttl)
	ls.GraceUntil = ls.ExpiresAt.Add(ls.gracePeriod)

	message := msg.Acquired
	if previousHolder != "" {
		message = msg.Acquired + " from " + previousHolder
	}

	return AcquireResult{
		Success:   true,
		Holder:    ls.Holder,
		ExpiresAt: ls.ExpiresAt,
		Message:   message,
	}
}
