package lockstatehttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shadyabhi/foolock/lockstate"
	"github.com/shadyabhi/foolock/lockstate/msg"
	"github.com/stretchr/testify/require"
)

func TestHandleAcquire(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*lockstate.LockState)
		query  string
		status int
		check  func(*testing.T, map[string]any)
	}{
		{
			name:   "missing client",
			setup:  nil,
			query:  "",
			status: http.StatusBadRequest,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] != "client parameter required" {
					t.Error("expected client required error")
				}
			},
		},
		{
			name:   "invalid ttl",
			setup:  nil,
			query:  "?client=c1&ttl=bad",
			status: http.StatusBadRequest,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] != "invalid ttl format" {
					t.Error("expected ttl format error")
				}
			},
		},
		{
			name:   "success",
			setup:  nil,
			query:  "?client=c1",
			status: http.StatusOK,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "c1" {
					t.Errorf("holder = %v, want c1", m["holder"])
				}
				if m["message"] != msg.Acquired {
					t.Errorf("message = %v, want %v", m["message"], msg.Acquired)
				}
			},
		},
		{
			name: "held by another",
			setup: func(ls *lockstate.LockState) {
				ls.Acquire("other", time.Minute)
			},
			query:  "?client=c1",
			status: http.StatusConflict,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != false {
					t.Error("expected success to be false")
				}
				if m["holder"] != "other" {
					t.Errorf("holder = %v, want other", m["holder"])
				}
				if m["message"] != msg.HeldByAnother {
					t.Errorf("message = %v, want %v", m["message"], msg.HeldByAnother)
				}
			},
		},
		{
			name: "grace period",
			setup: func(ls *lockstate.LockState) {
				ls.Holder = "other"
				ls.ExpiresAt = time.Now().Add(-time.Second)
				ls.GraceUntil = time.Now().Add(time.Minute)
			},
			query:  "?client=c1",
			status: http.StatusConflict,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] != msg.GracePeriodActive {
					t.Error("expected grace period error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := lockstate.New()
			if tt.setup != nil {
				tt.setup(ls)
			}
			h := NewHandler(ls)
			req := httptest.NewRequest(http.MethodPost, "/lock"+tt.query, nil)
			w := httptest.NewRecorder()
			h.HandleLock(w, req)

			if w.Code != tt.status {
				t.Errorf("status = %d, want %d", w.Code, tt.status)
			}
			var resp map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.check(t, resp)
		})
	}
}

func TestHandleRelease(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*lockstate.LockState)
		query  string
		status int
		check  func(*testing.T, map[string]any)
	}{
		{
			name:   "missing client",
			setup:  nil,
			query:  "",
			status: http.StatusBadRequest,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] != "client parameter required" {
					t.Error("expected client required error")
				}
			},
		},
		{
			name: "success",
			setup: func(ls *lockstate.LockState) {
				ls.Acquire("c1", time.Minute)
			},
			query:  "?client=c1",
			status: http.StatusOK,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["message"] != msg.LockReleased {
					t.Errorf("expected message %v, got %v", msg.LockReleased, m["message"])
				}
			},
		},
		{
			name: "not holder",
			setup: func(ls *lockstate.LockState) {
				ls.Acquire("other", time.Minute)
			},
			query:  "?client=c1",
			status: http.StatusForbidden,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] == nil {
					t.Error("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := lockstate.New()
			if tt.setup != nil {
				tt.setup(ls)
			}
			h := NewHandler(ls)
			req := httptest.NewRequest(http.MethodDelete, "/lock"+tt.query, nil)
			w := httptest.NewRecorder()
			h.HandleLock(w, req)

			if w.Code != tt.status {
				t.Errorf("status = %d, want %d", w.Code, tt.status)
			}
			var resp map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.check(t, resp)
		})
	}
}

func TestHandleStatus(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*lockstate.LockState)
		check func(*testing.T, map[string]any)
	}{
		{
			name:  "no lock",
			setup: nil,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "" {
					t.Errorf("holder = %v, want empty", m["holder"])
				}
				if m["message"] != msg.NoLockHeld {
					t.Errorf("message = %v, want %v", m["message"], msg.NoLockHeld)
				}
			},
		},
		{
			name: "lock held",
			setup: func(ls *lockstate.LockState) {
				ls.Acquire("c1", time.Minute)
			},
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "c1" {
					t.Errorf("holder = %v, want c1", m["holder"])
				}
				if m["message"] != msg.LockHeld {
					t.Errorf("message = %v, want %v", m["message"], msg.LockHeld)
				}
				if m["expires_at"] == nil {
					t.Error("expected expires_at")
				}
			},
		},
		{
			name: "in grace",
			setup: func(ls *lockstate.LockState) {
				ls.Holder = "c1"
				ls.ExpiresAt = time.Now().Add(-time.Second)
				ls.GraceUntil = time.Now().Add(time.Minute)
			},
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["message"] != msg.LockHeld {
					t.Errorf("message = %v, want %v", m["message"], msg.LockHeld)
				}
				if m["grace_until"] == nil {
					t.Error("expected grace_until")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := lockstate.New()
			if tt.setup != nil {
				tt.setup(ls)
			}
			h := NewHandler(ls)
			req := httptest.NewRequest(http.MethodGet, "/lock", nil)
			w := httptest.NewRecorder()
			h.HandleLock(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
			var resp map[string]any

			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			tt.check(t, resp)
		})
	}
}
