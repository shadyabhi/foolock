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
		setup  func(*lockstate.Manager)
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
			name:   "success with default job",
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
				if m["job"] != "default" {
					t.Errorf("job = %v, want default", m["job"])
				}
				if m["message"] != msg.Acquired {
					t.Errorf("message = %v, want %v", m["message"], msg.Acquired)
				}
			},
		},
		{
			name:   "success with custom job",
			setup:  nil,
			query:  "?client=c1&job=myjob",
			status: http.StatusOK,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "c1" {
					t.Errorf("holder = %v, want c1", m["holder"])
				}
				if m["job"] != "myjob" {
					t.Errorf("job = %v, want myjob", m["job"])
				}
			},
		},
		{
			name: "held by another",
			setup: func(m *lockstate.Manager) {
				m.Acquire("default", "other", time.Minute)
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
			name: "different jobs are independent",
			setup: func(m *lockstate.Manager) {
				m.Acquire("job1", "other", time.Minute)
			},
			query:  "?client=c1&job=job2",
			status: http.StatusOK,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true - different jobs should be independent")
				}
				if m["job"] != "job2" {
					t.Errorf("job = %v, want job2", m["job"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := lockstate.New()
			if tt.setup != nil {
				tt.setup(m)
			}
			h := New(m)
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
		setup  func(*lockstate.Manager)
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
			name: "success with default job",
			setup: func(m *lockstate.Manager) {
				m.Acquire("default", "c1", time.Minute)
			},
			query:  "?client=c1",
			status: http.StatusOK,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["job"] != "default" {
					t.Errorf("job = %v, want default", m["job"])
				}
				if m["message"] != msg.LockReleased {
					t.Errorf("expected message %v, got %v", msg.LockReleased, m["message"])
				}
			},
		},
		{
			name: "success with custom job",
			setup: func(m *lockstate.Manager) {
				m.Acquire("myjob", "c1", time.Minute)
			},
			query:  "?client=c1&job=myjob",
			status: http.StatusOK,
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["job"] != "myjob" {
					t.Errorf("job = %v, want myjob", m["job"])
				}
			},
		},
		{
			name: "not holder",
			setup: func(m *lockstate.Manager) {
				m.Acquire("default", "other", time.Minute)
			},
			query:  "?client=c1",
			status: http.StatusForbidden,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] == nil {
					t.Error("expected error")
				}
			},
		},
		{
			name: "wrong job",
			setup: func(m *lockstate.Manager) {
				m.Acquire("job1", "c1", time.Minute)
			},
			query:  "?client=c1&job=job2",
			status: http.StatusForbidden,
			check: func(t *testing.T, m map[string]any) {
				if m["error"] == nil {
					t.Error("expected error - client doesn't hold lock for job2")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := lockstate.New()
			if tt.setup != nil {
				tt.setup(m)
			}
			h := New(m)
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
		setup func(*lockstate.Manager)
		query string
		check func(*testing.T, map[string]any)
	}{
		{
			name:  "no lock with default job",
			setup: nil,
			query: "",
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "" {
					t.Errorf("holder = %v, want empty", m["holder"])
				}
				if m["job"] != "default" {
					t.Errorf("job = %v, want default", m["job"])
				}
				if m["message"] != msg.NoLockHeld {
					t.Errorf("message = %v, want %v", m["message"], msg.NoLockHeld)
				}
			},
		},
		{
			name: "lock held with default job",
			setup: func(m *lockstate.Manager) {
				m.Acquire("default", "c1", time.Minute)
			},
			query: "",
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "c1" {
					t.Errorf("holder = %v, want c1", m["holder"])
				}
				if m["job"] != "default" {
					t.Errorf("job = %v, want default", m["job"])
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
			name: "lock held with custom job",
			setup: func(m *lockstate.Manager) {
				m.Acquire("myjob", "c1", time.Minute)
			},
			query: "?job=myjob",
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "c1" {
					t.Errorf("holder = %v, want c1", m["holder"])
				}
				if m["job"] != "myjob" {
					t.Errorf("job = %v, want myjob", m["job"])
				}
			},
		},
		{
			name: "different jobs are independent",
			setup: func(m *lockstate.Manager) {
				m.Acquire("job1", "c1", time.Minute)
			},
			query: "?job=job2",
			check: func(t *testing.T, m map[string]any) {
				if m["success"] != true {
					t.Error("expected success to be true")
				}
				if m["holder"] != "" {
					t.Errorf("holder = %v, want empty (different job)", m["holder"])
				}
				if m["job"] != "job2" {
					t.Errorf("job = %v, want job2", m["job"])
				}
				if m["message"] != msg.NoLockHeld {
					t.Errorf("message = %v, want %v", m["message"], msg.NoLockHeld)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := lockstate.New()
			if tt.setup != nil {
				tt.setup(m)
			}
			h := New(m)
			req := httptest.NewRequest(http.MethodGet, "/lock"+tt.query, nil)
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
