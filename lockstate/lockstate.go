package lockstate

import (
	"sync"
	"time"
)

const (
	defaultTTL         = 30 * time.Second
	defaultGracePeriod = 5 * time.Second
)

type State struct {
	mu          sync.Mutex
	ttl         time.Duration
	gracePeriod time.Duration

	Job        string
	Holder     string
	AcquiredAt time.Time
	ExpiresAt  time.Time
	GraceUntil time.Time
}

func newState(ttl, gracePeriod time.Duration) *State {
	return &State{
		ttl:         ttl,
		gracePeriod: gracePeriod,
	}
}

// Manager manages locks for multiple jobs
type Manager struct {
	mu          sync.RWMutex
	locks       map[string]*State
	ttl         time.Duration
	gracePeriod time.Duration
}

type Option func(*Manager)

func WithTTL(d time.Duration) Option {
	return func(m *Manager) {
		m.ttl = d
	}
}

func WithGracePeriod(d time.Duration) Option {
	return func(m *Manager) {
		m.gracePeriod = d
	}
}

func New(opts ...Option) *Manager {
	m := &Manager{
		locks:       make(map[string]*State),
		ttl:         defaultTTL,
		gracePeriod: defaultGracePeriod,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// getOrCreateLock returns the lock for a job, creating it if needed
func (m *Manager) getOrCreateLock(job string) *State {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.locks[job]; ok {
		return s
	}

	s := newState(m.ttl, m.gracePeriod)
	s.Job = job
	m.locks[job] = s
	return s
}

// Acquire attempts to acquire a lock for a job
func (m *Manager) Acquire(job, client string, ttl time.Duration) AcquireResult {
	s := m.getOrCreateLock(job)
	return s.Acquire(client, ttl)
}

// Release releases a lock for a job
func (m *Manager) Release(job, client string) ReleaseResult {
	s := m.getOrCreateLock(job)
	return s.Release(client)
}

// Status returns the status of a lock for a job
func (m *Manager) Status(job string) StatusResult {
	s := m.getOrCreateLock(job)
	return s.Status()
}

func (s *State) IsExpired() bool {
	return time.Now().After(s.ExpiresAt) || s.Holder == ""
}

func (s *State) InGracePeriod() bool {
	now := time.Now()
	return now.After(s.ExpiresAt) && now.Before(s.GraceUntil)
}

type StatusResult struct {
	Job        string
	Holder     string
	ExpiresAt  time.Time
	GraceUntil time.Time
	IsExpired  bool
	InGrace    bool
}

func (s *State) Status() StatusResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	return StatusResult{
		Job:        s.Job,
		Holder:     s.Holder,
		ExpiresAt:  s.ExpiresAt,
		GraceUntil: s.GraceUntil,
		IsExpired:  s.IsExpired(),
		InGrace:    s.InGracePeriod(),
	}
}
