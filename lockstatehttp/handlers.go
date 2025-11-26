package lockstatehttp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shadyabhi/foolock/lockstate"
)

type LockResponse struct {
	Holder     string `json:"holder"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	IsExpired  bool   `json:"is_expired,omitempty"`
	GraceUntil string `json:"grace_until,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Handler struct {
	state *lockstate.LockState
}

func NewHandler(state *lockstate.LockState) *Handler {
	return &Handler{state: state}
}

func (h *Handler) HandleLock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		h.handleAcquire(w, r)
	case http.MethodDelete:
		h.handleRelease(w, r)
	case http.MethodGet:
		h.handleStatus(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleAcquire(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")
	if client == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "client parameter required"})
		return
	}

	ttlStr := r.URL.Query().Get("ttl")
	ttl := lockstate.DefaultTTL
	if ttlStr != "" {
		parsedTTL, err := time.ParseDuration(ttlStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid ttl format"})
			return
		}
		ttl = parsedTTL
	}

	result := h.state.Acquire(client, ttl)

	if result.Success {
		log.Printf("Lock %s by %s until %s", result.Message, client, result.ExpiresAt.Format(time.RFC3339))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LockResponse{
			Holder:    result.Holder,
			ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
		})
		return
	}

	if result.Message == "grace period active" {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "grace period active"})
		return
	}

	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w).Encode(LockResponse{
		Holder:    result.Holder,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *Handler) handleRelease(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")
	if client == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "client parameter required"})
		return
	}

	result := h.state.Release(client)

	if !result.Success {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: result.Message})
		return
	}

	log.Printf("Lock released by %s", client)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "lock released"})
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := h.state.Status()

	response := LockResponse{
		Holder:    status.Holder,
		IsExpired: status.IsExpired,
	}

	if status.Holder != "" {
		response.ExpiresAt = status.ExpiresAt.Format(time.RFC3339)
		if status.InGrace {
			response.GraceUntil = status.GraceUntil.Format(time.RFC3339)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
