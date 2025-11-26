package lockstate

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ls := New()
	if ls == nil {
		t.Fatal("New() returned nil")
	}
	if ls.Holder != "" {
		t.Errorf("expected empty Holder, got %q", ls.Holder)
	}
}

func TestNewWithGracePeriod(t *testing.T) {
	customGrace := 10 * time.Second
	ls := New(WithGracePeriod(customGrace))
	if ls.gracePeriod != customGrace {
		t.Errorf("gracePeriod = %v, want %v", ls.gracePeriod, customGrace)
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*LockState)
		expected bool
	}{
		{"empty holder", func(ls *LockState) {}, true},
		{"future expiry", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(time.Minute)
		}, false},
		{"past expiry", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(-time.Minute)
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := New()
			tt.setup(ls)
			if got := ls.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInGracePeriod(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*LockState)
		expected bool
	}{
		{"not expired", func(ls *LockState) {
			ls.ExpiresAt = time.Now().Add(time.Minute)
			ls.GraceUntil = time.Now().Add(2 * time.Minute)
		}, false},
		{"in grace", func(ls *LockState) {
			ls.ExpiresAt = time.Now().Add(-time.Second)
			ls.GraceUntil = time.Now().Add(time.Minute)
		}, true},
		{"past grace", func(ls *LockState) {
			ls.ExpiresAt = time.Now().Add(-2 * time.Minute)
			ls.GraceUntil = time.Now().Add(-time.Minute)
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := New()
			tt.setup(ls)
			if got := ls.InGracePeriod(); got != tt.expected {
				t.Errorf("InGracePeriod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*LockState)
		holder    string
		isExpired bool
		inGrace   bool
	}{
		{"empty", func(ls *LockState) {}, "", true, false},
		{"active lock", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(time.Minute)
			ls.GraceUntil = time.Now().Add(2 * time.Minute)
		}, "client1", false, false},
		{"in grace", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(-time.Second)
			ls.GraceUntil = time.Now().Add(time.Minute)
		}, "client1", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := New()
			tt.setup(ls)
			result := ls.Status()
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
