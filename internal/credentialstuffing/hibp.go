package credentialstuffing

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HIBPClient checks passwords against Have I Been Pwned (k-anonymity API)
// API: https://haveibeenpwned.com/API/v3#PwnedPasswords
// Легальное использование: демонстрация риска повторного использования паролей
type HIBPClient struct {
	apiKey     string
	httpClient *http.Client
}

// HIBPResult result of password check
type HIBPResult struct {
	Password   string
	PwnedCount int
	Pwned      bool
	Suffixes   []string // matched suffixes (for debugging)
}

// NewHIBPClient creates HIBP client
func NewHIBPClient(apiKey string) *HIBPClient {
	return &HIBPClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckPassword uses k-anonymity: send first 5 chars of SHA1, receive suffixes
func (c *HIBPClient) CheckPassword(password string) (*HIBPResult, error) {
	hash := sha1.Sum([]byte(password))
	hashStr := strings.ToUpper(hex.EncodeToString(hash[:]))
	prefix := hashStr[:5]

	req, err := http.NewRequest("GET", "https://api.pwnedpasswords.com/range/"+prefix, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PhantomProxy-RedTeam")
	if c.apiKey != "" {
		req.Header.Set("hibp-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hibp API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response: each line is "SUFFIX:count"
	lines := strings.Split(string(body), "\r\n")
	suffixToFind := hashStr[5:]
	var pwnedCount int
	var matched []string
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		suffix := strings.TrimSpace(parts[0])
		var count int
		fmt.Sscanf(parts[1], "%d", &count)
		if suffix == suffixToFind {
			pwnedCount = count
			matched = append(matched, suffix)
		}
	}

	return &HIBPResult{
		Password:   password,
		PwnedCount: pwnedCount,
		Pwned:      pwnedCount > 0,
		Suffixes:   matched,
	}, nil
}

// CheckPasswordsBatch checks multiple passwords (rate-limited)
func (c *HIBPClient) CheckPasswordsBatch(passwords []string) ([]HIBPResult, error) {
	var results []HIBPResult
	for i, p := range passwords {
		if i > 0 {
			time.Sleep(1 * time.Second) // HIBP rate limit: 1.5s between requests without key
		}
		r, err := c.CheckPassword(p)
		if err != nil {
			continue
		}
		results = append(results, *r)
	}
	return results, nil
}

// PasswordSprayConfig for automated spraying
type PasswordSprayConfig struct {
	Usernames     []string
	Passwords     []string
	TargetURL     string
	DelayMinutes  int
	LockoutLimit  int
}

// PasswordSprayResult single attempt result
type PasswordSprayResult struct {
	Username   string
	Password   string
	Success    bool
	StatusCode int
	Error      string
	AttemptAt  time.Time
}

// PasswordSprayEngine - conceptual: выполняет spraying с rate limiting
type PasswordSprayEngine struct {
	config *PasswordSprayConfig
	client *http.Client
}

// NewPasswordSprayEngine creates spray engine
func NewPasswordSprayEngine(cfg *PasswordSprayConfig) *PasswordSprayEngine {
	return &PasswordSprayEngine{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Run executes spray (концептуально - в production: HTTP POST с формами логина)
func (e *PasswordSprayEngine) Run() ([]PasswordSprayResult, error) {
	var results []PasswordSprayResult
	delay := time.Duration(e.config.DelayMinutes) * time.Minute
	if delay < time.Minute {
		delay = time.Minute
	}
	for _, user := range e.config.Usernames {
		for _, pass := range e.config.Passwords {
			// HTTP POST login
			_ = e.config.TargetURL
			results = append(results, PasswordSprayResult{
				Username:  user,
				Password:  pass,
				AttemptAt: time.Now(),
			})
			time.Sleep(delay)
		}
	}
	return results, nil
}

// MarshalJSON for HIBPResult - не логируем сам пароль
func (r HIBPResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"pwned":       r.Pwned,
		"pwned_count": r.PwnedCount,
	})
}
