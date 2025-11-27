package lockstatehttp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shadyabhi/foolock/lockstate"
	"github.com/shadyabhi/foolock/lockstate/msg"
)

type LockResponse struct {
	Success    bool   `json:"success"`
	Holder     string `json:"holder"`
	Message    string `json:"message,omitempty"`
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
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "client parameter required"}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	ttlStr := r.URL.Query().Get("ttl")
	ttl := 30 * time.Second
	if ttlStr != "" {
		parsedTTL, err := time.ParseDuration(ttlStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid ttl format"}); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			return
		}
		ttl = parsedTTL
	}

	result := h.state.Acquire(client, ttl)

	if result.Success {
		log.Printf("Lock %s by %s until %s (in %s)", result.Message, client, result.ExpiresAt.Format(time.RFC3339), time.Until(result.ExpiresAt).Round(time.Second))
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(LockResponse{
			Success:   true,
			Holder:    result.Holder,
			ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
			Message:   result.Message,
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	if result.Message == msg.GracePeriodActive {
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: msg.GracePeriodActive}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusConflict)
	if err := json.NewEncoder(w).Encode(LockResponse{
		Success:   result.Success,
		Holder:    result.Holder,
		Message:   result.Message,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *Handler) handleRelease(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")
	if client == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "client parameter required"}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	result := h.state.Release(client)

	if !result.Success {
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: result.Message}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	log.Printf("Lock released by %s (held for %s)", client, result.HeldFor.Round(time.Second))
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(LockResponse{
		Success: true,
		Message: result.Message,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *Handler) handleStatus(w http.ResponseWriter, _ *http.Request) {
	status := h.state.Status()

	response := LockResponse{
		Success:   true,
		Holder:    status.Holder,
		IsExpired: status.IsExpired,
	}

	if status.Holder != "" {
		response.Message = msg.LockHeld
		response.ExpiresAt = status.ExpiresAt.Format(time.RFC3339)
		if status.InGrace {
			response.GraceUntil = status.GraceUntil.Format(time.RFC3339)
		}
	} else {
		response.Message = msg.NoLockHeld
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
