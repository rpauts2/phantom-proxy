// simulation-lab is a local-only, synthetic-event awareness lab.
package main

import (
	"context"
	"flag"
	"log"
	"net"
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
	campaigns := flag.String("campaigns", "", "comma-separated explicit allowlist of synthetic campaign IDs (required)")
	flag.Parse()

	host, _, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalf("invalid listen address: %v", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		log.Fatal("simulation lab only binds to a loopback IP; use 127.0.0.1 or ::1")
	}

	allowlist := make([]string, 0)
	for _, id := range strings.Split(*campaigns, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			allowlist = append(allowlist, id)
		}
	}
	if len(allowlist) == 0 {
		log.Fatal("an explicit campaign allowlist is required: use -campaigns demo")
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           simulation.Handler(simulation.NewStore(500, allowlist...)),
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
