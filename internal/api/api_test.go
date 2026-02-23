package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"

    "github.com/gofiber/fiber/v2"
    "github.com/phantom-proxy/phantom-proxy/internal/config"
    "github.com/phantom-proxy/phantom-proxy/internal/database"
    "github.com/phantom-proxy/phantom-proxy/internal/proxy"
    "go.uber.org/zap"
)

// helper to send a request to a Fiber app with API key header
func sendRequest(t *testing.T, app *fiber.App, method, path string, body []byte, apiKey string) *http.Response {
    req := httptest.NewRequest(method, path, bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    if len(body) > 0 {
        req.Header.Set("Content-Type", "application/json")
    }
    resp, err := app.Test(req, -1)
    if err != nil {
        t.Fatalf("failed to perform request %s %s: %v", method, path, err)
    }
    return resp
}

func TestPhishletEndpoints(t *testing.T) {
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()

    // prepare temporary phishlet directory
    tmpDir, err := os.MkdirTemp("", "phishlets")
    if err != nil {
        t.Fatalf("failed to create tmp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)

    content := `id: sample
proxy_hosts:
  - phish_sub: ""
    orig_sub: "example"
    domain: "com"
js_inject: []`
    if err := os.WriteFile(filepath.Join(tmpDir, "sample.yaml"), []byte(content), 0644); err != nil {
        t.Fatalf("failed to write phishlet: %v", err)
    }

    db, err := database.NewDatabase(&database.DatabaseConfig{
        Type:       database.DatabaseSQLite,
        SQLitePath: ":memory:",
    }, logger)
    if err != nil {
        t.Logf("unable to create sqlite DB (CGO?): %v, proceeding with nil db", err)
        db = nil
    } else {
        defer db.Close()
    }

    cfg := &config.Config{PhishletsPath: tmpDir}
    proxyObj, err := proxy.NewHTTPProxy(cfg, db, nil, logger)
    if err != nil {
        t.Fatalf("NewHTTPProxy error: %v", err)
    }

    apiKey := "secret"
    server := NewAPIServer(proxyObj, db, logger, apiKey)

    // list phishlets should contain sample
    resp := sendRequest(t, server.app, "GET", "/api/v1/phishlets", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status code listing phishlets: %d", resp.StatusCode)
    }

    var listResp struct {
        Phishlets []map[string]interface{} `json:"phishlets"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
        t.Fatalf("failed to decode list response: %v", err)
    }
    found := false
    for _, p := range listResp.Phishlets {
        if p["id"] == "sample" || p["ID"] == "sample" {
            found = true
            break
        }
    }
    if !found {
        t.Errorf("sample phishlet not present in list")
    }

    // sessions endpoint should return empty list (db nil or no sessions)
    resp = sendRequest(t, server.app, "GET", "/api/v1/sessions", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status code listing sessions: %d", resp.StatusCode)
    }
    var sessResp struct {
        Sessions []map[string]interface{} `json:"sessions"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&sessResp); err != nil {
        t.Fatalf("failed to decode sessions response: %v", err)
    }
    if len(sessResp.Sessions) != 0 {
        t.Errorf("expected zero sessions, got %d", len(sessResp.Sessions))
    }

    // credentials endpoint should also return empty list
    resp = sendRequest(t, server.app, "GET", "/api/v1/credentials", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status code listing credentials: %d", resp.StatusCode)
    }
    var credResp struct {
        Credentials []map[string]interface{} `json:"credentials"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&credResp); err != nil {
        t.Fatalf("failed to decode credentials response: %v", err)
    }
    if len(credResp.Credentials) != 0 {
        t.Errorf("expected zero credentials, got %d", len(credResp.Credentials))
    }

    // stats endpoint should return JSON with expected keys
    resp = sendRequest(t, server.app, "GET", "/api/v1/stats", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status code stats: %d", resp.StatusCode)
    }
    var statsResp map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
        t.Fatalf("failed to decode stats response: %v", err)
    }
    for _, key := range []string{"total_sessions", "active_sessions", "total_requests"} {
        if _, ok := statsResp[key]; !ok {
            t.Errorf("stats missing key %s", key)
        }
    }

    // enable endpoint
    resp = sendRequest(t, server.app, "POST", "/api/v1/phishlets/sample/enable", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("enable endpoint returned %d", resp.StatusCode)
    }
    ph, err := proxyObj.GetPhishlet("sample")
    if err != nil {
        t.Fatalf("GetPhishlet error: %v", err)
    }
    if !ph.Enabled {
        t.Error("proxy did not enable phishlet")
    }

    // disable endpoint
    resp = sendRequest(t, server.app, "POST", "/api/v1/phishlets/sample/disable", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("disable endpoint returned %d", resp.StatusCode)
    }
    ph, err = proxyObj.GetPhishlet("sample")
    if err != nil {
        t.Fatalf("GetPhishlet error: %v", err)
    }
    if ph.Enabled {
        t.Error("proxy did not disable phishlet")
    }

    // health check should succeed (no targets defined so unknown status)
    resp = sendRequest(t, server.app, "GET", "/api/v1/phishlets/sample/health", nil, apiKey)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("health endpoint returned %d", resp.StatusCode)
    }
    var health map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        t.Fatalf("failed to decode health response: %v", err)
    }
    if health["phishlet_id"] != "sample" {
        t.Errorf("unexpected health phishlet_id: %v", health["phishlet_id"])
    }
}
