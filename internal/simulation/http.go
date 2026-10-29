package simulation

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type eventRequest struct {
	CampaignID    string `json:"campaign_id"`
	ParticipantID string `json:"participant_id"`
	Type          string `json:"type"`
}

// Handler exposes health, status, emergency stop, and synthetic-event endpoints.
func Handler(store *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		status := http.StatusOK
		if store.IsStopped() {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]any{"status": map[bool]string{true: "stopped", false: "ok"}[store.IsStopped()]})
	})

	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"stopped": store.IsStopped(),
			"events":  len(store.List(500)),
		})
	})

	mux.HandleFunc("/admin/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		store.Stop()
		writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
	})

	mux.HandleFunc("/api/v1/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			limit := 100
			if raw := r.URL.Query().Get("limit"); raw != "" {
				n, err := strconv.Atoi(raw)
				if err != nil || n < 1 || n > 500 {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 500"})
					return
				}
				limit = n
			}
			writeJSON(w, http.StatusOK, map[string]any{"events": store.List(limit)})
		case http.MethodPost:
			r.Body = http.MaxBytesReader(w, r.Body, 4096)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			var req eventRequest
			if err := decoder.Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event request"})
				return
			}
			var extra any
			if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request must contain one JSON object"})
				return
			}
			event, err := store.Add(strings.TrimSpace(req.CampaignID), strings.TrimSpace(req.ParticipantID), req.Type)
			if err != nil {
				status := http.StatusBadRequest
				message := "invalid event"
				switch {
				case errors.Is(err, ErrInvalidID):
					message = "campaign_id and participant_id must use 1-64 letters, digits, underscores, or hyphens"
				case errors.Is(err, ErrInvalidEvent):
					message = "type must be visit, click, or training_complete"
				case errors.Is(err, ErrCampaignNotInScope):
					status = http.StatusForbidden
					message = "campaign is not in the explicit simulation allowlist"
				case errors.Is(err, ErrLabStopped):
					status = http.StatusServiceUnavailable
					message = "simulation lab is stopped"
				}
				writeJSON(w, status, map[string]string{"error": message})
				return
			}
			writeJSON(w, http.StatusCreated, event)
		default:
			methodNotAllowed(w, http.MethodGet+", "+http.MethodPost)
		}
	})

	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
