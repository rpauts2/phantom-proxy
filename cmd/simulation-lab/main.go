// simulation-lab is a local-only, synthetic-event awareness lab.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/simulation"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8088", "HTTP listen address (must be a loopback IP)")
	dataPath := flag.String("data", ".phantom-lab/events.json", "path to the local synthetic-event snapshot")
	campaigns := flag.String("campaigns", "", "comma-separated explicit allowlist of synthetic campaign IDs (required)")
	scopeDuration := flag.Duration("scope-duration", 8*time.Hour, "maximum simulation scope lifetime (1m to 24h)")
	flag.Parse()

	if err := validateListenAddress(*addr); err != nil {
		log.Fatal(err)
	}
	allowlist, err := parseCampaignAllowlist(*campaigns)
	if err != nil {
		log.Fatal(err)
	}
	if err := validateScopeDuration(*scopeDuration); err != nil {
		log.Fatal(err)
	}
	scopeExpiresAt := time.Now().Add(*scopeDuration)
	store, err := simulation.OpenPersistentStore(*dataPath, 500, scopeExpiresAt, allowlist...)
	if err != nil {
		log.Fatalf("open persistent simulation store: %v", err)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           simulation.Handler(store),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("simulation lab listening on http://%s", *addr)
		log.Printf("allowed synthetic campaign IDs: %s", strings.Join(allowlist, ","))
		log.Printf("simulation scope expires at %s", store.ScopeExpiresAt().UTC().Format(time.RFC3339))
		log.Printf("synthetic event snapshot: %s", *dataPath)
		log.Printf("synthetic events only; no credential or session data is accepted")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = server.Close()
	}
}
