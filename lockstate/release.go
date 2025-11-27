package lockstate

import (
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

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
			Message: msg.ClientNotHolder,
		}
	}

	ls.Holder = ""
	ls.ExpiresAt = time.Time{}
	ls.GraceUntil = time.Time{}

	return ReleaseResult{
		Success: true,
		Message: msg.LockReleased,
	}
}
