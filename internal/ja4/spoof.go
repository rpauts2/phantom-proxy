// Package ja4 - JA4 Fingerprint Spoofing
// Modern TLS fingerprinting bypass
package ja4

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ============================================================================
// JA4 Fingerprint Types
// ============================================================================

type JA4Fingerprint struct {
	Version    string   // TLS version (13 for TLS 1.3)
	ALPN       string   // Application Layer Protocol
	Ciphers    string   // First 12 cipher suites
	Extensions string   // First 12 extensions
	Signal     string   // Server name indication
	Hash       string   // MD5 of the fingerprint
}

// JA3-like fingerprint (legacy support)
type JA3Fingerprint struct {
	Version    string
	Ciphers    string
	Extensions string
	Elliptic   string
	Cert       string
	Hash       string
}

// Browser profiles for spoofing
type BrowserProfile struct {
	Name         string   `json:"name"`
	JA4Hash      string   `json:"ja4_hash"`
	JA3Hash      string   `json:"ja3_hash"`
	TLSVersion   string   `json:"tls_version"`
	Ciphers     []string `json:"ciphers"`
	ALPN        []string `json:"alpn"`
	Extensions  []string `json:"extensions"`
	Curves      []string `json:"curves"`
	Signature   []string `json:"signature"`
	UserAgent   string   `json:"user_agent"`
}

// ============================================================================
// Browser Profiles (Real JA4 fingerprints)
// ============================================================================

var browserProfiles = map[string]*BrowserProfile{
	"chrome-120": {
		Name:        "Chrome 120",
		TLSVersion:  "13",
		Ciphers: []string{
			"0x1301", "0x1302", "0x1303", "0xc02b", "0xc02f", "0xc02c",
			"0xc030", "0xcca9", "0xcca8", "0xcad9", "0x009d", "0x009c",
		},
		ALPN: []string{"h3", "h2", "http/1.1"},
		Extensions: []string{
			"0x0000", "0x000a", "0x000b", "0x0017", "0x0023",
			"0x002d", "0x002f", "0x0033", "0x003c", "0x0045",
			"0x001b", "0x001d",
		},
		Curves:     []string{"0x001d", "0x0017", "0x001e", "0x0019"},
		Signature:  []string{"0x0403", "0x0503", "0x080a", "0x0804", "0x0807"},
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	},
	"firefox-121": {
		Name:        "Firefox 121",
		TLSVersion:  "13",
		Ciphers: []string{
			"0x1301", "0x1302", "0x1303", "0xcca9", "0xcca8", "0xc02b",
			"0xc02f", "0xc02c", "0xc030", "0x009d", "0x009c", "0x006b",
		},
		ALPN: []string{"h2", "http/1.1"},
		Extensions: []string{
			"0x0000", "0x000a", "0x000b", "0x0017", "0x0023",
			"0x002d", "0x002f", "0x0033", "0x003c", "0x0045",
			"0x001b", "0x001d", "0xff01",
		},
		Curves:     []string{"0x001d", "0x0017", "0x001e", "0x0019", "0x0018"},
		Signature:  []string{"0x0403", "0x0503", "0x080a", "0x0804", "0x0807"},
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	},
	"edge-120": {
		Name:        "Edge 120",
		TLSVersion:  "13",
		Ciphers: []string{
			"0x1301", "0x1302", "0x1303", "0xc02b", "0xc02f", "0xc02c",
			"0xc030", "0xcca9", "0xcca8", "0xcad9", "0x009d", "0x009c",
		},
		ALPN: []string{"h3", "h2", "http/1.1"},
		Extensions: []string{
			"0x0000", "0x000a", "0x000b", "0x0017", "0x0023",
			"0x002d", "0x002f", "0x0033", "0x003c", "0x0045",
			"0x001b", "0x001d",
		},
		Curves:     []string{"0x001d", "0x0017", "0x001e", "0x0019"},
		Signature:  []string{"0x0403", "0x0503", "0x080a", "0x0804", "0x0807"},
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	},
	"safari-17": {
		Name:        "Safari 17",
		TLSVersion:  "13",
		Ciphers: []string{
			"0x1301", "0x1302", "0x1303", "0xcca9", "0xcca8", "0xc02b",
			"0xc02f", "0xc02c", "0xc030", "0x009d", "0x009c", "0x006b",
		},
		ALPN: []string{"h2", "http/1.1"},
		Extensions: []string{
			"0x0000", "0x000a", "0x000b", "0x0017", "0x0023",
			"0x002d", "0x002f", "0x0033", "0x003c", "0x0055",
			"0x001b", "0x001d",
		},
		Curves:     []string{"0x001d", "0x0017", "0x001e"},
		Signature:  []string{"0x0403", "0x0503", "0x080a", "0x0804"},
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	},
}

// ============================================================================
// Spoofer Implementation
// ============================================================================

type Spoofer struct {
	logger     *zap.Logger
	profiles   map[string]*BrowserProfile
	stats      *SpoofStats
	mu         sync.RWMutex
}

type SpoofStats struct {
	mu             sync.RWMutex
	Spoofed       int64   `json:"spoofed"`
	Failed        int64   `json:"failed"`
	ByBrowser     map[string]int64 `json:"by_browser"`
	LastSpoofTime time.Time `json:"last_spoof_time"`
}

func NewSpoofer(logger *zap.Logger) *Spoofer {
	return &Spoofer{
		logger:   logger,
		profiles: browserProfiles,
		stats: &SpoofStats{
			ByBrowser: make(map[string]int64),
		},
	}
}

// ============================================================================
// JA4 Fingerprint Generation
// ============================================================================

// GenerateJA4 creates a JA4 fingerprint from a browser profile
func (s *Spoofer) GenerateJA4(profile *BrowserProfile) *JA4Fingerprint {
	// JA4 format: q13h3_000000000000_000000000000_000000000000_0000000000000000000000000000000000
	
	// Version (13 for TLS 1.3)
	version := profile.TLSVersion
	
	// ALPN (first 2 chars = first 2 protocols)
	alpn := ""
	if len(profile.ALPN) > 0 {
		alpn = strings.ReplaceAll(profile.ALPN[0], "-", "")
		if len(alpn) > 2 {
			alpn = alpn[:2]
		}
	}
	if alpn == "" {
		alpn = "00"
	}
	
	// Ciphers (first 12 ciphers)
	ciphers := make([]string, 0)
	for _, c := range profile.Ciphers {
		if len(ciphers) >= 12 {
			break
		}
		// Remove 0x prefix
		c = strings.ToLower(strings.ReplaceAll(c, "0x", ""))
		ciphers = append(ciphers, c)
	}
	cipherStr := strings.Join(ciphers, "")
	if len(cipherStr) < 24 {
		cipherStr += strings.Repeat("0", 24-len(cipherStr))
	}
	
	// Extensions (first 12 extensions)
	extensions := make([]string, 0)
	for _, e := range profile.Extensions {
		if len(extensions) >= 12 {
			break
		}
		e = strings.ToLower(strings.ReplaceAll(e, "0x", ""))
		extensions = append(extensions, e)
	}
	extStr := strings.Join(extensions, "")
	if len(extStr) < 24 {
		extStr += strings.Repeat("0", 24-len(extStr))
	}
	
	// SNI (signal) - just use example.com as placeholder
	signal := "05examplecom"
	
	// Create hash
	raw := fmt.Sprintf("%s%s_%s_%s_%s_%s", version, alpn, cipherStr[:12], cipherStr[12:24], extStr[:12], extStr[12:24])
	hash := sha256.Sum256([]byte(raw))
	hashStr := hex.EncodeToString(hash[:12]) // First 12 bytes
	
	return &JA4Fingerprint{
		Version:    version,
		ALPN:       alpn,
		Ciphers:    cipherStr[:24],
		Extensions:  extStr[:24],
		Signal:     signal,
		Hash:       hashStr,
	}
}

// GenerateJA3 creates a JA3 fingerprint from a browser profile
func (s *Spoofer) GenerateJA3(profile *BrowserProfile) *JA3Fingerprint {
	// JA3 format: version,ciphers,extensions,curves,signatures
	
	version := profile.TLSVersion
	if version == "13" {
		version = "0x0304" // TLS 1.3 maps to 0x0304 in JA3
	}
	
	ciphers := make([]string, 0)
	for _, c := range profile.Ciphers {
		ciphers = append(ciphers, strings.ToLower(c))
	}
	cipherStr := strings.Join(ciphers, ",")
	
	extensions := make([]string, 0)
	for _, e := range profile.Extensions {
		extensions = append(extensions, strings.ToLower(e))
	}
	extStr := strings.Join(extensions, ",")
	
	curves := make([]string, 0)
	for _, c := range profile.Curves {
		curves = append(curves, strings.ToLower(c))
	}
	curveStr := strings.Join(curves, ",")
	
	sig := make([]string, 0)
	for _, s := range profile.Signature {
		sig = append(sig, strings.ToLower(s))
	}
	sigStr := strings.Join(sig, ",")
	
	// Create hash
	raw := fmt.Sprintf("%s-%s-%s-%s-%s", version, cipherStr, extStr, curveStr, sigStr)
	hash := md5.Sum([]byte(raw))
	hashStr := hex.EncodeToString(hash[:])
	
	return &JA3Fingerprint{
		Version:    version,
		Ciphers:    cipherStr,
		Extensions: extStr,
		Elliptic:   curveStr,
		Cert:       sigStr,
		Hash:       hashStr,
	}
}

// ============================================================================
// Spoofing
// ============================================================================

// SpoofRequest spoofs a request to look like a specific browser
func (s *Spoofer) SpoofRequest(browserName string) (*JA4Fingerprint, *JA3Fingerprint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	profile, ok := s.profiles[browserName]
	if !ok {
		// Default to Chrome
		profile = s.profiles["chrome-120"]
	}
	
	ja4 := s.GenerateJA4(profile)
	ja3 := s.GenerateJA3(profile)
	
	// Update stats
	s.stats.mu.Lock()
	s.stats.Spoofed++
	s.stats.ByBrowser[browserName]++
	s.stats.LastSpoofTime = time.Now()
	s.stats.mu.Unlock()
	
	s.logger.Debug("Request spoofed",
		zap.String("browser", browserName),
		zap.String("ja4", ja4.Hash),
		zap.String("ja3", ja3.Hash))
	
	return ja4, ja3, nil
}

// SpoofRandom spoofs a request with a random browser profile
func (s *Spoofer) SpoofRandom() (*JA4Fingerprint, *JA3Fingerprint, string, error) {
	s.mu.RLock()
	profiles := make([]string, 0, len(s.profiles))
	for name := range s.profiles {
		profiles = append(profiles, name)
	}
	s.mu.RUnlock()
	
	if len(profiles) == 0 {
		return nil, nil, "", fmt.Errorf("no profiles available")
	}
	
	browserName := profiles[rand.Intn(len(profiles))]
	ja4, ja3, err := s.SpoofRequest(browserName)
	
	return ja4, ja3, browserName, err
}

// GetProfile returns a browser profile by name
func (s *Spoofer) GetProfile(name string) *BrowserProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if profile, ok := s.profiles[name]; ok {
		return profile
	}
	
	return s.profiles["chrome-120"] // Default
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns spoofing statistics
func (s *Spoofer) GetStats() map[string]interface{} {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()
	
	return map[string]interface{}{
		"total_spoofed":    s.stats.Spoofed,
		"total_failed":     s.stats.Failed,
		"last_spoof_time":  s.stats.LastSpoofTime,
		"by_browser":       s.stats.ByBrowser,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// DetectFingerprint detects the browser from a JA4/JA3 fingerprint
func (s *Spoofer) DetectFingerprint(ja4Hash, ja3Hash string) (string, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Compare against known fingerprints
	for name, profile := range s.profiles {
		ja4 := s.GenerateJA4(profile)
		if ja4.Hash == ja4Hash {
			return name, 1.0
		}
		
		ja3 := s.GenerateJA3(profile)
		if ja3.Hash == ja3Hash {
			return name, 0.9
		}
	}
	
	return "unknown", 0.0
}

// GetRandomUserAgent returns a random user agent
func (s *Spoofer) GetRandomUserAgent() string {
	profile := s.GetProfile("chrome-120") // Default
	return profile.UserAgent
}

// GetUserAgent returns user agent for a specific browser
func (s *Spoofer) GetUserAgent(browserName string) string {
	profile := s.GetProfile(browserName)
	return profile.UserAgent
}

// ============================================================================
// TLS Config Generation
// ============================================================================

// GetTLSConfig generates TLS config for spoofing
func (s *Spoofer) GetTLSConfig(browserName string) map[string]interface{} {
	profile := s.GetProfile(browserName)
	
	return map[string]interface{}{
		"min_version":           "1.3",
		"max_version":           "1.3",
		"cipher_suites":         profile.Ciphers,
		"curves":                profile.Curves,
		"signature_algorithms":   profile.Signature,
		"alpn":                 profile.ALPN,
		"server_name":           "example.com",
		"insecure_skip_verify":  false,
	}
}
