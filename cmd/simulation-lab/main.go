// simulation-lab is a local-only, synthetic-event awareness lab.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/simulation"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8088", "HTTP listen address (defaults to localhost only)")
	flag.Parse()

	server := &http.Server{
		Addr:              *addr,
		Handler:           simulation.Handler(simulation.NewStore(500)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("simulation lab listening on http://%s", *addr)
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
