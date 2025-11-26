package lockstate

import (
	"testing"
	"time"
)

func TestAcquire(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*LockState)
		client  string
		ttl     time.Duration
		success bool
		message string
	}{
		{"fresh lock", func(ls *LockState) {}, "client1", time.Minute, true, "acquired"},
		{"renew own lock", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(time.Minute)
		}, "client1", time.Minute, true, "renewed"},
		{"held by another", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(time.Minute)
		}, "client2", time.Minute, false, "held by another client"},
		{"in grace period", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(-time.Second)
			ls.GraceUntil = time.Now().Add(time.Minute)
		}, "client2", time.Minute, false, "grace period active"},
		{"expired past grace", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(-2 * time.Minute)
			ls.GraceUntil = time.Now().Add(-time.Minute)
		}, "client2", time.Minute, true, "acquired from client1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := New()
			tt.setup(ls)
			result := ls.Acquire(tt.client, tt.ttl)
			if result.Success != tt.success {
				t.Errorf("Success = %v, want %v", result.Success, tt.success)
			}
			if result.Message != tt.message {
				t.Errorf("Message = %q, want %q", result.Message, tt.message)
			}
		})
	}
}
