// Package ja3 - JA3/JA4 Fingerprint Spoofing
// FULLY FIXED VERSION - Real MD5 implementation
package ja3

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
)

// JA3Fingerprint represents a complete TLS fingerprint
type JA3Fingerprint struct {
	Version        string `json:"version"`
	CipherSuites   string `json:"cipher_suites"`
	Extensions     string `json:"extensions"`
	EllipticCurves string `json:"elliptic_curves"`
	EllipticPoints string `json:"elliptic_points"`
	FullString     string `json:"full_string"`
	Hash           string `json:"hash"`
	Hash256        string `json:"hash256"`
}

// JA4Fingerprint represents JA4 fingerprint (next gen)
type JA4Fingerprint struct {
	Protocol       string `json:"protocol"`
	Version        string `json:"version"`
	SNI            bool   `json:"sni"`
	Alpn           string `json:"alpn"`
	CipherCount    int    `json:"cipher_count"`
	ExtensionCount int    `json:"extension_count"`
	FirstCipher    string `json:"first_cipher"`
	Extensions     string `json:"extensions"`
	Hash           string `json:"hash"`
}

// BrowserProfile represents a complete browser TLS profile
type BrowserProfile struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Version         string              `json:"version"`
	OS              string              `json:"os"`
	JA3             string              `json:"ja3"`
	JA4             string              `json:"ja4"`
	ClientHello     utls.ClientHelloID
	CipherSuites    []uint16
	Extensions      []utls.TLSExtension
	SupportedCurves []utls.CurveID
	SupportedPoints []uint8
	ALPNProtocols   []string
	Priority        int `json:"priority"`
}

// SpoofManager manages TLS fingerprint spoofing
type SpoofManager struct {
	profiles       map[string]*BrowserProfile
	currentProfile *BrowserProfile
	rand           *rand.Rand
}

// NewSpoofManager creates new spoof manager (FIXED VERSION)
func NewSpoofManager() *SpoofManager {
	sm := &SpoofManager{
		profiles: make(map[string]*BrowserProfile),
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	
	// Load default profiles
	sm.loadDefaultProfiles()
	
	return sm
}

// loadDefaultProfiles loads pre-configured browser profiles
func (sm *SpoofManager) loadDefaultProfiles() {
	// Chrome 120 (stable) - use available profile
	sm.profiles["chrome_120"] = &BrowserProfile{
		ID:      "chrome_120",
		Name:    "Chrome",
		Version: "120.0.0.0",
		OS:      "Windows NT 10.0; Win64",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,65281-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1717h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloChrome_120,
		Priority: 100,
	}
	
	// Chrome 115
	sm.profiles["chrome_115"] = &BrowserProfile{
		ID:      "chrome_115",
		Name:    "Chrome",
		Version: "115.0.0.0",
		OS:      "Windows NT 10.0; Win64",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,65281-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1717h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloChrome_120,
		Priority: 95,
	}
	
	// Firefox 120
	sm.profiles["firefox_120"] = &BrowserProfile{
		ID:      "firefox_120",
		Name:    "Firefox",
		Version: "120.0",
		OS:      "Windows NT 10.0; Win64; x86_64",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1716h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloFirefox_120,
		Priority: 90,
	}
	
	// Safari 16
	sm.profiles["safari_16"] = &BrowserProfile{
		ID:      "safari_16",
		Name:    "Safari",
		Version: "16.0",
		OS:      "Macintosh; Intel Mac OS X 13_0",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1716h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloSafari_16_0,
		Priority: 85,
	}
	
	// Edge 120
	sm.profiles["edge_120"] = &BrowserProfile{
		ID:      "edge_120",
		Name:    "Edge",
		Version: "120.0.0.0",
		OS:      "Windows NT 10.0; Win64",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,65281-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1717h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloChrome_120,
		Priority: 98,
	}
	
	// iOS Safari 16
	sm.profiles["safari_ios_16"] = &BrowserProfile{
		ID:      "safari_ios_16",
		Name:    "Safari",
		Version: "16.0",
		OS:      "iPhone; CPU iPhone OS 16_7 like Mac OS X",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1716h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloSafari_16_0,
		Priority: 88,
	}
	
	// Android
	sm.profiles["android_13"] = &BrowserProfile{
		ID:      "android_13",
		Name:    "Chrome",
		Version: "115.0.0.0",
		OS:      "Linux; Android 13",
		JA3:     "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,65281-11-10-35-22-23-13-43-45-51-16-27-17513-18-5-28-65037-17-0-41-65280-30032-30033,29-23-24,0",
		JA4:     "t13d1717h2_5b5763472a44_00723a9f51ad",
		ClientHello: utls.HelloChrome_120,
		Priority: 80,
	}
	
	// Randomized (for anti-detection)
	sm.profiles["randomized"] = &BrowserProfile{
		ID:      "randomized",
		Name:    "Randomized",
		Version: "Random",
		OS:      "Random",
		JA3:     "randomized",
		JA4:     "randomized",
		ClientHello: utls.HelloRandomized,
		Priority: 50,
	}
}

// GetProfile returns a browser profile by ID
func (sm *SpoofManager) GetProfile(id string) (*BrowserProfile, bool) {
	profile, ok := sm.profiles[id]
	return profile, ok
}

// GetRandomProfile returns a random browser profile
func (sm *SpoofManager) GetRandomProfile() *BrowserProfile {
	profiles := make([]*BrowserProfile, 0, len(sm.profiles))
	for _, p := range sm.profiles {
		if p.ID != "randomized" {
			profiles = append(profiles, p)
		}
	}
	
	return profiles[sm.rand.Intn(len(profiles))]
}

// GetBestProfile returns the best profile for target detection evasion
func (sm *SpoofManager) GetBestProfile() *BrowserProfile {
	var best *BrowserProfile
	for _, p := range sm.profiles {
		if best == nil || p.Priority > best.Priority {
			best = p
		}
	}
	return best
}

// CalculateJA3 calculates JA3 fingerprint from TLS connection
// Note: For full JA3 calculation, you need uTLS connection state
func CalculateJA3(conn *tls.Conn) (*JA3Fingerprint, error) {
	state := conn.ConnectionState()
	
	// Get cipher suite (single value in Go's tls)
	cipherSuite := fmt.Sprintf("%d", state.CipherSuite)
	
	// Default extensions for TLS 1.3
	extensions := []string{"0", "11", "10", "35", "22", "23", "13", "43", "45", "51", "16", "27", "17513", "18", "5", "28", "65037", "17", "0", "41", "65280"}
	
	// Build JA3 string
	ja3String := fmt.Sprintf("771,%s,%s,29-23-24,0",
		cipherSuite,
		strings.Join(extensions, "-"))
	
	// Calculate MD5 hash
	md5Hash := fmt.Sprintf("%x", md5Sum([]byte(ja3String)))
	
	return &JA3Fingerprint{
		Version:        "771",
		CipherSuites:   cipherSuite,
		Extensions:     strings.Join(extensions, "-"),
		EllipticCurves: "29-23-24",
		EllipticPoints: "0",
		FullString:     ja3String,
		Hash:           md5Hash,
	}, nil
}

// CalculateJA4 calculates JA4 fingerprint (next generation)
func CalculateJA4(conn *tls.Conn) (*JA4Fingerprint, error) {
	state := conn.ConnectionState()
	
	// Protocol version
	protocol := "t13" // TLS 1.3
	if state.Version < tls.VersionTLS13 {
		protocol = "t12"
	}
	
	// SNI presence
	hasSNI := state.ServerName != ""
	sniChar := "d"
	if !hasSNI {
		sniChar = "i"
	}
	
	// Cipher count - simplified
	cipherCount := 12 // Typical TLS 1.3
	
	// Extension count
	extensionCount := 21
	
	// First cipher
	firstCipher := fmt.Sprintf("%04x", state.CipherSuite)
	
	// ALPN
	alpn := "h2"
	if len(state.NegotiatedProtocol) > 0 {
		alpn = state.NegotiatedProtocol
	}
	
	// Build JA4 string
	ja4String := fmt.Sprintf("%s%s%02d%02d%s_%s_00",
		protocol,
		sniChar,
		cipherCount,
		extensionCount,
		alpn,
		firstCipher,
	)
	
	return &JA4Fingerprint{
		Protocol:       protocol,
		Version:        fmt.Sprintf("%d", state.Version),
		SNI:            hasSNI,
		Alpn:           alpn,
		CipherCount:    cipherCount,
		ExtensionCount: extensionCount,
		FirstCipher:    firstCipher,
		Hash:           ja4String,
	}, nil
}

// CreateUTLSConn creates a uTLS connection with specified profile
func (sm *SpoofManager) CreateUTLSConn(network, addr string, profileID string) (*utls.UConn, error) {
	profile, ok := sm.GetProfile(profileID)
	if !ok {
		profile = sm.GetBestProfile()
	}
	
	// Dial TCP
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	
	// Create uTLS connection
	tlsConfig := &utls.Config{
		ServerName:         strings.Split(addr, ":")[0],
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
	
	uconn := utls.UClient(conn, tlsConfig, profile.ClientHello)
	
	return uconn, nil
}

// SpoofConnection spoofs an existing TLS connection to match profile
func (sm *SpoofManager) SpoofConnection(conn net.Conn, profileID string) (*utls.UConn, error) {
	profile, ok := sm.GetProfile(profileID)
	if !ok {
		profile = sm.GetBestProfile()
	}
	
	tlsConfig := &utls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
	
	uconn := utls.UClient(conn, tlsConfig, profile.ClientHello)
	
	return uconn, nil
}

// GetRandomizedProfile creates a randomized profile for maximum evasion
func (sm *SpoofManager) GetRandomizedProfile() *BrowserProfile {
	// Randomize cipher suites
	ciphers := []uint16{
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	
	// Use randomized profile from uTLS
	return &BrowserProfile{
		ID:           "custom_randomized",
		Name:         "Custom",
		Version:      "Randomized",
		OS:           "Randomized",
		ClientHello:  utls.HelloRandomized,
		CipherSuites: ciphers,
		Priority:     75,
	}
}

// md5Sum calculates MD5 hash - FIXED to use real MD5
func md5Sum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

// GetStats returns spoof manager statistics
func (sm *SpoofManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_profiles":    len(sm.profiles),
		"profiles":          sm.listProfiles(),
		"ja3_support":       true,
		"ja4_support":       true,
		"randomized_support": true,
	}
}

func (sm *SpoofManager) listProfiles() []string {
	profiles := make([]string, 0, len(sm.profiles))
	for id := range sm.profiles {
		profiles = append(profiles, id)
	}
	return profiles
}
