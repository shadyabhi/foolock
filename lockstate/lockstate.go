package lockstate

import (
	"sync"
	"time"
)

const (
	defaultTTL         = 30 * time.Second
	defaultGracePeriod = 5 * time.Second
)

type LockState struct {
	mu          sync.Mutex
	ttl         time.Duration
	gracePeriod time.Duration

	Holder     string
	AcquiredAt time.Time
	ExpiresAt  time.Time
	GraceUntil time.Time
}

type Option func(*LockState)

func WithTTL(d time.Duration) Option {
	return func(ls *LockState) {
		ls.ttl = d
	}
}

func WithGracePeriod(d time.Duration) Option {
	return func(ls *LockState) {
		ls.gracePeriod = d
	}
}

func New(opts ...Option) *LockState {
	ls := &LockState{
		ttl:         defaultTTL,
		gracePeriod: defaultGracePeriod,
	}
	for _, opt := range opts {
		opt(ls)
	}
	return ls
}

func (ls *LockState) IsExpired() bool {
	return time.Now().After(ls.ExpiresAt) || ls.Holder == ""
}

func (ls *LockState) InGracePeriod() bool {
	now := time.Now()
	return now.After(ls.ExpiresAt) && now.Before(ls.GraceUntil)
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
