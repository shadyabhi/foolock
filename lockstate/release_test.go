package lockstate

import (
	"testing"
	"time"

	"github.com/shadyabhi/foolock/lockstate/msg"
)

func TestRelease(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*LockState)
		client  string
		success bool
		message string
	}{
		{"release own lock", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(time.Minute)
		}, "client1", true, msg.LockReleased},
		{"release other's lock", func(ls *LockState) {
			ls.Holder = "client1"
			ls.ExpiresAt = time.Now().Add(time.Minute)
		}, "client2", false, msg.ClientNotHolder},
		{"release empty lock", func(ls *LockState) {}, "client1", false, msg.ClientNotHolder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := New()
			tt.setup(ls)
			result := ls.Release(tt.client)
			if result.Success != tt.success {
				t.Errorf("Success = %v, want %v", result.Success, tt.success)
			}
			if result.Message != tt.message {
				t.Errorf("Message = %q, want %q", result.Message, tt.message)
			}
		})
	}
}
