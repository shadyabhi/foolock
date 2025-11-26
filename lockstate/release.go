package lockstate

import "time"

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
