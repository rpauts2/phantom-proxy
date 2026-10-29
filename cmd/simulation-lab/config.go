package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/simulation"
)

// validateListenAddress permits numeric loopback addresses only. This keeps the
// awareness lab off public and LAN interfaces by default and by explicit flag.
func validateListenAddress(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("simulation lab only binds to a loopback IP; use 127.0.0.1 or ::1")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("listen port must be a number between 1 and 65535")
	}
	return nil
}

// parseCampaignAllowlist returns validated, unique synthetic campaign IDs.
func parseCampaignAllowlist(raw string) ([]string, error) {
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, candidate := range strings.Split(raw, ",") {
		id := strings.TrimSpace(candidate)
		if id == "" {
			continue
		}
		if !simulation.ValidID(id) {
			return nil, fmt.Errorf("invalid synthetic campaign ID %q: use 1-64 letters, digits, underscores, or hyphens", id)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate synthetic campaign ID %q", id)
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("an explicit campaign allowlist is required: use -campaigns demo")
	}
	return ids, nil
}

func validateScopeDuration(d time.Duration) error {
	if d < time.Minute || d > 24*time.Hour {
		return fmt.Errorf("scope duration must be between 1m and 24h")
	}
	return nil
}
