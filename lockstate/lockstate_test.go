package lockstate

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.locks == nil {
		t.Fatal("locks map should be initialized")
	}
}

func TestNewWithOptions(t *testing.T) {
	customTTL := 1 * time.Minute
	customGrace := 10 * time.Second
	m := New(WithTTL(customTTL), WithGracePeriod(customGrace))
	if m.ttl != customTTL {
		t.Errorf("ttl = %v, want %v", m.ttl, customTTL)
	}
	if m.gracePeriod != customGrace {
		t.Errorf("gracePeriod = %v, want %v", m.gracePeriod, customGrace)
	}
}

func TestManagerAcquireRelease(t *testing.T) {
	m := New()

	// Acquire lock for job1
	result := m.Acquire("job1", "client1", time.Minute)
	if !result.Success {
		t.Errorf("expected success, got failure")
	}
	if result.Holder != "client1" {
		t.Errorf("holder = %q, want client1", result.Holder)
	}

	// Different job should be independent
	result2 := m.Acquire("job2", "client2", time.Minute)
	if !result2.Success {
		t.Errorf("expected success for job2, got failure")
	}

	// Same job, different client should fail
	result3 := m.Acquire("job1", "client2", time.Minute)
	if result3.Success {
		t.Errorf("expected failure for job1/client2, got success")
	}

	// Release job1
	releaseResult := m.Release("job1", "client1")
	if !releaseResult.Success {
		t.Errorf("expected release success, got failure")
	}

	// Now client2 can acquire job1
	result4 := m.Acquire("job1", "client2", time.Minute)
	if !result4.Success {
		t.Errorf("expected success after release, got failure")
	}

	// Cleanup
	m.Release("job1", "client2")
	m.Release("job2", "client2")
}

func TestManagerStatus(t *testing.T) {
	m := New()

	// Status for non-existent job should return empty state
	status := m.Status("newjob")
	if status.Holder != "" {
		t.Errorf("expected empty holder for new job, got %q", status.Holder)
	}
	if status.Job != "newjob" {
		t.Errorf("expected job 'newjob', got %q", status.Job)
	}
	if !status.IsExpired {
		t.Errorf("expected IsExpired=true for empty lock")
	}

	// Acquire and check status
	m.Acquire("testjob", "client1", time.Minute)
	status = m.Status("testjob")
	if status.Holder != "client1" {
		t.Errorf("expected holder client1, got %q", status.Holder)
	}
	if status.Job != "testjob" {
		t.Errorf("expected job 'testjob', got %q", status.Job)
	}

	m.Release("testjob", "client1")
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*State)
		expected bool
	}{
		{"empty holder", func(s *State) {}, true},
		{"future expiry", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(time.Minute)
		}, false},
		{"past expiry", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(-time.Minute)
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{}
			tt.setup(s)
			if got := s.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInGracePeriod(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*State)
		expected bool
	}{
		{"not expired", func(s *State) {
			s.ExpiresAt = time.Now().Add(time.Minute)
			s.GraceUntil = time.Now().Add(2 * time.Minute)
		}, false},
		{"in grace", func(s *State) {
			s.ExpiresAt = time.Now().Add(-time.Second)
			s.GraceUntil = time.Now().Add(time.Minute)
		}, true},
		{"past grace", func(s *State) {
			s.ExpiresAt = time.Now().Add(-2 * time.Minute)
			s.GraceUntil = time.Now().Add(-time.Minute)
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{}
			tt.setup(s)
			if got := s.InGracePeriod(); got != tt.expected {
				t.Errorf("InGracePeriod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*State)
		holder    string
		isExpired bool
		inGrace   bool
	}{
		{"empty", func(s *State) {}, "", true, false},
		{"active lock", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(time.Minute)
			s.GraceUntil = time.Now().Add(2 * time.Minute)
		}, "client1", false, false},
		{"in grace", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(-time.Second)
			s.GraceUntil = time.Now().Add(time.Minute)
		}, "client1", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{}
			tt.setup(s)
			result := s.Status()
			if result.Holder != tt.holder {
				t.Errorf("Holder = %q, want %q", result.Holder, tt.holder)
			}
			if result.IsExpired != tt.isExpired {
				t.Errorf("IsExpired = %v, want %v", result.IsExpired, tt.isExpired)
			}
			if result.InGrace != tt.inGrace {
				t.Errorf("InGrace = %v, want %v", result.InGrace, tt.inGrace)
			}
		})
	}
}
