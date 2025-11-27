package lockstate

import (
	"testing"
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

func TestAcquire(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*State)
		client  string
		ttl     time.Duration
		success bool
		message string
	}{
		{"fresh lock", func(s *State) {}, "client1", time.Minute, true, msg.Acquired},
		{"renew own lock", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(time.Minute)
		}, "client1", time.Minute, true, msg.Renewed},
		{"held by another", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(time.Minute)
		}, "client2", time.Minute, false, msg.HeldByAnother},
		{"in grace period", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(-time.Second)
			s.GraceUntil = time.Now().Add(time.Minute)
		}, "client2", time.Minute, false, msg.GracePeriodActive},
		{"expired past grace", func(s *State) {
			s.Holder = "client1"
			s.ExpiresAt = time.Now().Add(-2 * time.Minute)
			s.GraceUntil = time.Now().Add(-time.Minute)
		}, "client2", time.Minute, true, msg.Acquired + " from client1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{ttl: 30 * time.Second, gracePeriod: 5 * time.Second}
			tt.setup(s)
			result := s.Acquire(tt.client, tt.ttl)
			if result.Success != tt.success {
				t.Errorf("Success = %v, want %v", result.Success, tt.success)
			}
			if result.Message != tt.message {
				t.Errorf("Message = %q, want %q", result.Message, tt.message)
			}
		})
	}
}
