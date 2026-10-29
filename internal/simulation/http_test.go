package simulation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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
		for header, want := range map[string]string{
			"Referrer-Policy":                 "no-referrer",
			"Permissions-Policy":               "camera=(), microphone=(), geolocation=()",
			"Cross-Origin-Resource-Policy":     "same-origin",
			"X-Permitted-Cross-Domain-Policies": "none",
		} {
			if got := res.Header().Get(header); got != want {
				t.Errorf("GET %s header %s = %q, want %q", path, header, got, want)
			}
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

func TestHandlerRejectsExpiredScopeAndMissingWriteHeader(t *testing.T) {
	expired := NewStoreWithExpiry(10, time.Now().Add(-time.Second), "demo")
	handler := Handler(expired)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(
		`{"campaign_id":"demo","participant_id":"person_01","type":"visit"}`,
	))
	req.Header.Set("X-Phantom-Lab-Request", "1")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusGone {
		t.Fatalf("expired scope POST status = %d, want %d", res.Code, http.StatusGone)
	}

	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if res.Code != http.StatusGone {
		t.Fatalf("expired scope health status = %d, want %d", res.Code, http.StatusGone)
	}

	active := Handler(NewStore(10, "demo"))
	req = httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(
		`{"campaign_id":"demo","participant_id":"person_01","type":"visit"}`,
	))
	res = httptest.NewRecorder()
	active.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("POST without local request header status = %d, want %d", res.Code, http.StatusForbidden)
	}
}

func TestMetricsEndpointReportsOnlyBoundedSyntheticCounters(t *testing.T) {
	store := NewStore(3, "demo")
	_, _ = store.Add("demo", "person_01", "visit")
	_, _ = store.Add("demo", "person_02", "click")
	handler := Handler(store)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "text/plain; version=0.0.4; charset=utf-8" {
		t.Fatalf("unexpected metrics content type: %q", got)
	}
	for _, expected := range []string{
		"phantom_lab_events_retained 2",
		"phantom_lab_events_by_type_retained{type=\"visit\"} 1",
		"phantom_lab_events_by_type_retained{type=\"click\"} 1",
		"phantom_lab_scope_expired 0",
	} {
		if !bytes.Contains(res.Body.Bytes(), []byte(expected)) {
			t.Fatalf("metrics response missing %q: %s", expected, res.Body.String())
		}
	}
	if bytes.Contains(res.Body.Bytes(), []byte(`\n`)) {
		t.Fatal("metrics output contains literal escaped newlines instead of line breaks")
	}
}

func TestCSVExportContainsOnlySyntheticEventFields(t *testing.T) {
	store := NewStore(10, "demo")
	_, _ = store.Add("demo", "person_01", "visit")
	handler := Handler(store)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/events.csv?limit=10", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("CSV status = %d, want %d: %s", res.Code, http.StatusOK, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("unexpected CSV content type: %q", got)
	}
	if got := res.Header().Get("Content-Disposition"); got != "attachment; filename=phantom-lab-events.csv" {
		t.Fatalf("unexpected content disposition: %q", got)
	}
	for _, expected := range []string{"sequence,campaign_id,participant_id,type,occurred_at", "demo,person_01,visit"} {
		if !bytes.Contains(res.Body.Bytes(), []byte(expected)) {
			t.Fatalf("CSV response missing %q: %s", expected, res.Body.String())
		}
	}
	if bytes.Contains(res.Body.Bytes(), []byte("password")) || bytes.Contains(res.Body.Bytes(), []byte("cookie")) {
		t.Fatal("CSV unexpectedly contains credential/session fields")
	}
}

func TestCSVExportRejectsInvalidLimitAndMethods(t *testing.T) {
	handler := Handler(NewStore(10, "demo"))
	for _, tc := range []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/v1/events.csv?limit=501", http.StatusBadRequest},
		{http.MethodPost, "/api/v1/events.csv", http.StatusMethodNotAllowed},
	} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(tc.method, tc.path, nil))
		if res.Code != tc.want {
			t.Errorf("%s %s status = %d, want %d", tc.method, tc.path, res.Code, tc.want)
		}
	}
}

func TestSummaryEndpointAggregatesOnlyRetainedSyntheticEvents(t *testing.T) {
	store := NewStore(3, "demo", "second")
	_, _ = store.Add("demo", "person_01", "visit")
	_, _ = store.Add("demo", "person_02", "click")
	_, _ = store.Add("second", "person_03", "training_complete")
	handler := Handler(store)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("summary status = %d, want %d: %s", res.Code, http.StatusOK, res.Body.String())
	}
	var payload struct {
		RetainedEvents int            `json:"retained_events"`
		ByType         map[string]int `json:"by_type"`
		ByCampaign     map[string]int `json:"by_campaign"`
		Stopped        bool           `json:"stopped"`
		ScopeExpired   bool           `json:"scope_expired"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.RetainedEvents != 3 || payload.ByType["visit"] != 1 || payload.ByType["click"] != 1 || payload.ByType["training_complete"] != 1 {
		t.Fatalf("unexpected summary counts: %#v", payload)
	}
	if payload.ByCampaign["demo"] != 2 || payload.ByCampaign["second"] != 1 || payload.Stopped || payload.ScopeExpired {
		t.Fatalf("unexpected summary state: %#v", payload)
	}
}

func TestSummaryEndpointRejectsNonGet(t *testing.T) {
	res := httptest.NewRecorder()
	Handler(NewStore(10, "demo")).ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/summary", nil))
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("summary POST status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlerReportsPersistenceFailureAsServerError(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(10, "demo")
	// Replacing a directory with the snapshot file must fail, exercising the
	// durable-write error path without relying on filesystem permissions.
	store.persistencePath = filepath.Join(dir, "snapshot-target")
	if err := os.Mkdir(store.persistencePath, 0o700); err != nil {
		t.Fatal(err)
	}
	handler := Handler(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBufferString(
		`{"campaign_id":"demo","participant_id":"person_01","type":"visit"}`,
	))
	req.Header.Set("X-Phantom-Lab-Request", "1")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("persistence failure status = %d, want %d: %s", res.Code, http.StatusInternalServerError, res.Body.String())
	}
	if got := len(store.List(10)); got != 0 {
		t.Fatalf("event count after failed persistence = %d, want 0", got)
	}
}
