package lockstate

import (
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

type ReleaseResult struct {
	Success bool
	Job     string
	Message string
	HeldFor time.Duration
}

func (s *State) Release(client string) ReleaseResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Holder != client {
		return ReleaseResult{
			Success: false,
			Job:     s.Job,
			Message: msg.ClientNotHolder,
		}
	}

	heldFor := time.Since(s.AcquiredAt)
	job := s.Job

	s.Holder = ""
	s.AcquiredAt = time.Time{}
	s.ExpiresAt = time.Time{}
	s.GraceUntil = time.Time{}

	return ReleaseResult{
		Success: true,
		Job:     job,
		Message: msg.LockReleased,
		HeldFor: heldFor,
	}
}
