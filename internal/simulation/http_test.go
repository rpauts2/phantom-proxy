package simulation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerAcceptsOnlySyntheticEventSchema(t *testing.T) {
	handler := Handler(NewStore(10, "demo"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(
		`{"campaign_id":"demo","participant_id":"person_1","type":"click"}`,
	))
	req.Header.Set("X-Phantom-Lab-Request", "1")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d: %s", res.Code, http.StatusCreated, res.Body.String())
	}
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("security header missing")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(
		`{"campaign_id":"demo","participant_id":"person_1","type":"click","password":"secret"}`,
	))
	req.Header.Set("X-Phantom-Lab-Request", "1")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandlerListsEventsAndHealth(t *testing.T) {
	store := NewStore(10, "demo")
	_, _ = store.Add("demo", "person_1", "visit")
	handler := Handler(store)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/events?limit=10", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", res.Code, http.StatusOK)
	}
	var payload struct {
		Events []Event `json:"events"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Events) != 1 || payload.Events[0].Type != "visit" {
		t.Fatalf("unexpected events response: %#v", payload.Events)
	}

	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", res.Code, http.StatusOK)
	}
}

func TestHandlerEmergencyStopRejectsNewEvents(t *testing.T) {
	handler := Handler(NewStore(10, "demo"))

	res := httptest.NewRecorder()
	stopReq := httptest.NewRequest(http.MethodPost, "/admin/stop", nil)
	stopReq.Header.Set("X-Phantom-Lab-Action", "stop")
	handler.ServeHTTP(res, stopReq)
	if res.Code != http.StatusOK {
		t.Fatalf("stop status = %d, want %d", res.Code, http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(
		`{"campaign_id":"demo","participant_id":"person_1","type":"visit"}`,
	))
	req.Header.Set("X-Phantom-Lab-Request", "1")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("event after stop status = %d, want %d", res.Code, http.StatusServiceUnavailable)
	}

	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("health after stop status = %d, want %d", res.Code, http.StatusServiceUnavailable)
	}
}

func TestDashboardServesStaticAssetsWithSecurityHeaders(t *testing.T) {
	handler := Handler(NewStore(10, "demo"))
	for _, path := range []string{"/", "/style.css", "/app.js"} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, res.Code, http.StatusOK)
		}
		if res.Header().Get("Content-Security-Policy") == "" {
			t.Fatalf("GET %s missing Content-Security-Policy", path)
		}
		if res.Body.Len() == 0 {
			t.Fatalf("GET %s returned empty body", path)
		}
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/unknown-route", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("unknown route status = %d, want %d", res.Code, http.StatusNotFound)
	}
}
