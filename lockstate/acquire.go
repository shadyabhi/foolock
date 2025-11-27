package lockstate

import (
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

type AcquireResult struct {
	Success bool
	Job     string
	Holder  string
	Message string

	ExpiresAt  time.Time
	GraceUntil time.Time
}

func (s *State) Acquire(client string, ttl time.Duration) AcquireResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if s.isCurrentHolderRenewing(client) {
		return s.respRenewLock(now, ttl)
	}

	if s.isHeldByAnother(now) {
		return s.respAlreadyLocked()
	}

	if s.isInGracePeriod(now) {
		return s.respActiveGracePeriod()
	}

	return s.acquireLock(client, now, ttl)
}

func (s *State) isCurrentHolderRenewing(client string) bool {
	return s.Holder == client
}

func (s *State) respRenewLock(now time.Time, ttl time.Duration) AcquireResult {
	s.ExpiresAt = now.Add(ttl)
	s.GraceUntil = s.ExpiresAt.Add(s.gracePeriod)
	return AcquireResult{
		Success:   true,
		Job:       s.Job,
		Holder:    s.Holder,
		ExpiresAt: s.ExpiresAt,
		Message:   msg.Renewed,
	}
}

func (s *State) isHeldByAnother(now time.Time) bool {
	return s.Holder != "" && now.Before(s.ExpiresAt)
}

func (s *State) respAlreadyLocked() AcquireResult {
	return AcquireResult{
		Success:   false,
		Job:       s.Job,
		Holder:    s.Holder,
		ExpiresAt: s.ExpiresAt,
		Message:   msg.HeldByAnother,
	}
}

func (s *State) isInGracePeriod(now time.Time) bool {
	return s.Holder != "" && now.After(s.ExpiresAt) && now.Before(s.GraceUntil)
}

func (s *State) respActiveGracePeriod() AcquireResult {
	return AcquireResult{
		Success:    false,
		Job:        s.Job,
		Holder:     s.Holder,
		ExpiresAt:  s.ExpiresAt,
		GraceUntil: s.GraceUntil,
		Message:    msg.GracePeriodActive,
	}
}

func (s *State) acquireLock(client string, now time.Time, ttl time.Duration) AcquireResult {
	previousHolder := s.Holder
	s.Holder = client
	s.AcquiredAt = now
	s.ExpiresAt = now.Add(ttl)
	s.GraceUntil = s.ExpiresAt.Add(s.gracePeriod)

	message := msg.Acquired
	if previousHolder != "" {
		message = msg.Acquired + " from " + previousHolder
	}

	return AcquireResult{
		Success:   true,
		Job:       s.Job,
		Holder:    s.Holder,
		ExpiresAt: s.ExpiresAt,
		Message:   message,
	}
}
