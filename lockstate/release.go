package lockstate

import (
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

type ReleaseResult struct {
	Success    bool
	Message    string
	HeldFor    time.Duration
}

func (ls *LockState) Release(client string) ReleaseResult {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.Holder != client {
		return ReleaseResult{
			Success: false,
			Message: msg.ClientNotHolder,
		}
	}

	heldFor := time.Since(ls.AcquiredAt)

	ls.Holder = ""
	ls.AcquiredAt = time.Time{}
	ls.ExpiresAt = time.Time{}
	ls.GraceUntil = time.Time{}

	return ReleaseResult{
		Success: true,
		Message: msg.LockReleased,
		HeldFor: heldFor,
	}
}
