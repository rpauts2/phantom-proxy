package simulation

import (
	"encoding/json"
	"fmt"
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

// Handler exposes the local dashboard, health, status, emergency stop, and synthetic-event endpoints.
func Handler(store *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveDashboard(w, r)
	})
	mux.HandleFunc("/style.css", serveDashboardCSS)
	mux.HandleFunc("/app.js", serveDashboardJS)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		events := store.List(500)
		counts := map[string]int{"visit": 0, "click": 0, "training_complete": 0}
		for _, event := range events {
			counts[event.Type]++
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "# HELP phantom_lab_events_retained Number of synthetic events currently retained in memory.\n# TYPE phantom_lab_events_retained gauge\nphantom_lab_events_retained %d\n", len(events))
		_, _ = fmt.Fprintf(w, "# HELP phantom_lab_events_by_type_retained Synthetic events retained by type.\n# TYPE phantom_lab_events_by_type_retained gauge\n")
		for _, eventType := range []string{"visit", "click", "training_complete"} {
			_, _ = fmt.Fprintf(w, "phantom_lab_events_by_type_retained{type=%q} %d\n", eventType, counts[eventType])
		}
		stopped := 0
		if store.IsStopped() {
			stopped = 1
		}
		expired := 0
		if store.IsScopeExpired() {
			expired = 1
		}
		_, _ = fmt.Fprintf(w, "phantom_lab_stopped %d\nphantom_lab_scope_expired %d\n", stopped, expired)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		status := http.StatusOK
		state := "ok"
		if store.IsStopped() {
			status = http.StatusServiceUnavailable
			state = "stopped"
		} else if store.IsScopeExpired() {
			status = http.StatusGone
			state = "scope_expired"
		}
		writeJSON(w, status, map[string]string{"status": state})
	})

	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		expiresAt := store.ScopeExpiresAt()
		var expiresAtValue any
		if !expiresAt.IsZero() {
			expiresAtValue = expiresAt.UTC().Format("2006-01-02T15:04:05Z")
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"stopped":          store.IsStopped(),
			"scope_expired":    store.IsScopeExpired(),
			"scope_expires_at": expiresAtValue,
			"events":           len(store.List(500)),
		})
	})

	mux.HandleFunc("/admin/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		if r.Header.Get("X-Phantom-Lab-Action") != "stop" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "missing local action header"})
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
			if r.Header.Get("X-Phantom-Lab-Request") != "1" {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "missing local request header"})
				return
			}
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
				case errors.Is(err, ErrScopeExpired):
					status = http.StatusGone
					message = "simulation scope has expired; restart with a new explicit scope"
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
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
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
