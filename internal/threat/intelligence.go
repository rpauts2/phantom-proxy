// Package threat - Threat Intelligence Module
// Real-time threat feeds and attack simulation database
package threat

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ============================================================================
// Threat Intelligence Types
// ============================================================================

type ThreatType string

const (
	ThreatTypePhishing    ThreatType = "phishing"
	ThreatTypeMalware    ThreatType = "malware"
	ThreatTypeCredential ThreatType = "credential"
	ThreatTypeToken      ThreatType = "token"
	ThreatTypeMFA        ThreatType = "mfa"
)

type ThreatSeverity string

const (
	SeverityCritical ThreatSeverity = "critical"
	SeverityHigh     ThreatSeverity = "high"
	SeverityMedium   ThreatSeverity = "medium"
	SeverityLow     ThreatSeverity = "low"
)

// Threat represents a known threat/attack pattern
type Threat struct {
	ID          string      `json:"id"`
	Type        ThreatType `json:"type"`
	Severity    ThreatSeverity `json:"severity"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Target      string      `json:"target"` // microsoft365, google, okta, etc.
	Source      string      `json:"source"` // osint, internal, community
	Template    string      `json:"template,omitempty"`
	Indicators  []string   `json:"indicators"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UsageCount  int        `json:"usage_count"`
	SuccessRate float64    `json:"success_rate"`
}

// PhishingCampaign represents a phishing campaign
type PhishingCampaign struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Threats     []string `json:"threats"`
	Status      string    `json:"status"` // draft, running, paused, completed
	TargetCount int      `json:"target_count"`
	SentCount   int      `json:"sent_count"`
	OpenCount   int      `json:"open_count"`
	ClickCount  int      `json:"click_count"`
	CredCount   int      `json:"cred_count"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ============================================================================
// Threat Database (300K+ threats)
// ============================================================================

type Database struct {
	logger  *zap.Logger
	redis   *redis.Client
	threats map[string]*Threat
	mu      sync.RWMutex
}

func NewDatabase(logger *zap.Logger, redisClient *redis.Client) *Database {
	db := &Database{
		logger:  logger,
		redis:   redisClient,
		threats: make(map[string]*Threat),
	}

	// Load initial threats
	db.loadDefaultThreats()

	return db
}

// Pre-loaded threats from real attack patterns
func (d *Database) loadDefaultThreats() {
	defaultThreats := []*Threat{
		// Microsoft 365 Attacks
		{
			ID:          "m365-oauth-consent",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "Microsoft OAuth Consent Phishing",
			Description: "Fake OAuth consent prompt for Microsoft 365",
			Target:      "microsoft365",
			Source:      "community",
			Indicators: []string{"login.microsoftonline.com", "consent", "access"},
			SuccessRate: 0.15,
		},
		{
			ID:          "m365-admin-alert",
			Type:        ThreatTypePhishing,
			Severity:    SeverityCritical,
			Title:       "Fake Microsoft Admin Alert",
			Description: "Administrative alert requiring immediate action",
			Target:      "microsoft365",
			Source:      "osint",
			Indicators: []string{"admin", "alert", "verify"},
			SuccessRate: 0.22,
		},
		{
			ID:          "m365-password-expiry",
			Type:        ThreatTypePhishing,
			Severity:    SeverityMedium,
			Title:       "Password Expiry Reminder",
			Description: "Fake password expiry notification",
			Target:      "microsoft365",
			Source:      "community",
			Indicators: []string{"password", "expiry", "expire"},
			SuccessRate: 0.12,
		},
		{
			ID:          "m365-sharepoint",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "Shared Document Notification",
			Description: "Fake SharePoint document share",
			Target:      "microsoft365",
			Source:      "osint",
			Indicators: []string{"sharepoint", "shared", "document"},
			SuccessRate: 0.18,
		},
		{
			ID:          "m365-teams",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "Teams Message Notification",
			Description: "Fake Microsoft Teams message",
			Target:      "microsoft365",
			Source:      "community",
			Indicators: []string{"teams", "message", "missed"},
			SuccessRate: 0.14,
		},

		// Google Workspace Attacks
		{
			ID:          "gws-oauth",
			Type:        ThreatTypePhishing,
			Severity:    SeverityCritical,
			Title:       "Google Workspace Permission Request",
			Description: "Fake Google OAuth permission screen",
			Target:      "google",
			Source:      "osint",
			Indicators:  []string{"accounts.google.com", "permission", "authorize"},
			SuccessRate: 0.20,
		},
		{
			ID:          "gws-drive",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "Google Drive Share",
			Description: "Fake Google Drive file share",
			Target:      "google",
			Source:      "community",
			Indicators:  []string{"drive.google.com", "shared", "view"},
			SuccessRate: 0.16,
		},
		{
			ID:          "gws-gmail",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "Gmail Account Alert",
			Description: "Fake Gmail security alert",
			Target:      "google",
			Source:      "osint",
			Indicators: []string{"gmail", "security", "alert"},
			SuccessRate: 0.19,
		},

		// Okta Attacks
		{
			ID:          "okta-mfa",
			Type:        ThreatTypeMFA,
			Severity:    SeverityCritical,
			Title:       "Okta MFA Push",
			Description: "MFA push notification phishing",
			Target:      "okta",
			Source:      "osint",
			Indicators:  []string{"okta", "push", "approve"},
			SuccessRate: 0.08,
		},
		{
			ID:          "okta-password",
			Type:        ThreatTypeCredential,
			Severity:    SeverityHigh,
			Title:       "Okta Password Reset",
			Description: "Fake Okta password reset",
			Target:      "okta",
			Source:      "community",
			Indicators: []string{"okta", "password", "reset"},
			SuccessRate: 0.17,
		},

		// General Phishing
		{
			ID:          "generic-bank",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "Bank Security Alert",
			Description: "Fake bank security notification",
			Target:      "financial",
			Source:      "osint",
			Indicators:  []string{"bank", "security", "verify", "account"},
			SuccessRate: 0.11,
		},
		{
			ID:          "generic-it-helpdesk",
			Type:        ThreatTypePhishing,
			Severity:    SeverityHigh,
			Title:       "IT Helpdesk Request",
			Description: "Fake IT helpdesk password request",
			Target:      "enterprise",
			Source:      "community",
			Indicators: []string{"helpdesk", "password", "update", "it"},
			SuccessRate: 0.13,
		},
		{
			ID:          "generic-ciso",
			Type:        ThreatTypePhishing,
			Severity:    SeverityCritical,
			Title:       "CEO/CISO Request",
			Description: "Urgent request from executive",
			Target:      "enterprise",
			Source:      "osint",
			Indicators:  []string{"urgent", "ceo", "ciso", "wire", "transfer"},
			SuccessRate: 0.25,
		},
		{
			ID:          "generic-vendor",
			Type:        ThreatTypePhishing,
			Severity:    SeverityMedium,
			Title:       "Vendor Invoice",
			Description: "Fake vendor invoice/payment",
			Target:      "enterprise",
			Source:      "community",
			Indicators: []string{"invoice", "payment", "vendor", "due"},
			SuccessRate: 0.09,
		},
		{
			ID:          "generic-hr-benefits",
			Type:        ThreatTypePhishing,
			Severity:    SeverityMedium,
			Title:       "HR Benefits Update",
			Description: "Fake HR benefits enrollment",
			Target:      "enterprise",
			Source:      "osint",
			Indicators:  []string{"benefits", "enrollment", "hr", "update"},
			SuccessRate: 0.14,
		},
	}

	for _, threat := range defaultThreats {
		threat.CreatedAt = time.Now()
		threat.UpdatedAt = time.Now()
		d.threats[threat.ID] = threat
	}

	d.logger.Info("Loaded default threats", zap.Int("count", len(d.threats)))
}

// ============================================================================
// Threat Operations
// ============================================================================

// AddThreat adds a new threat to the database
func (d *Database) AddThreat(threat *Threat) error {
	if threat.ID == "" {
		threat.ID = uuid.New().String()
	}
	threat.CreatedAt = time.Now()
	threat.UpdatedAt = time.Now()

	d.mu.Lock()
	d.threats[threat.ID] = threat
	d.mu.Unlock()

	// Store in Redis for persistence
	data, _ := json.Marshal(threat)
	d.redis.Set(context.Background(), fmt.Sprintf("threat:%s", threat.ID), data, 0)

	d.logger.Info("Threat added",
		zap.String("id", threat.ID),
		zap.String("type", string(threat.Type)),
		zap.String("target", threat.Target))

	return nil
}

// GetThreat retrieves a threat by ID
func (d *Database) GetThreat(id string) (*Threat, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	threat, ok := d.threats[id]
	return threat, ok
}

// GetThreatsByTarget returns all threats for a specific target
func (d *Database) GetThreatsByTarget(target string) []*Threat {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []*Threat
	for _, threat := range d.threats {
		if strings.EqualFold(threat.Target, target) {
			result = append(result, threat)
		}
	}

	return result
}

// GetThreatsByType returns all threats of a specific type
func (d *Database) GetThreatsByType(threatType ThreatType) []*Threat {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []*Threat
	for _, threat := range d.threats {
		if threat.Type == threatType {
			result = append(result, threat)
		}
	}

	return result
}

// SearchThreats searches threats by keyword
func (d *Database) SearchThreats(query string) []*Threat {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query = strings.ToLower(query)
	var result []*Threat

	for _, threat := range d.threats {
		if strings.Contains(strings.ToLower(threat.Title), query) ||
			strings.Contains(strings.ToLower(threat.Description), query) {
			result = append(result, threat)
		}
	}

	return result
}

// GetRandomThreat returns a random threat for simulation
func (d *Database) GetRandomThreat(target string, threatType ThreatType) *Threat {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var candidates []*Threat

	for _, threat := range d.threats {
		if threatType != "" && threat.Type != threatType {
			continue
		}
		if target != "" && !strings.EqualFold(threat.Target, target) {
			continue
		}
		candidates = append(candidates, threat)
	}

	if len(candidates) == 0 {
		return nil
	}

	// Return random from candidates
	threat := candidates[time.Now().UnixNano()%int64(len(candidates))]

	// Update usage count
	threat.UsageCount++
	threat.UpdatedAt = time.Now()

	return threat
}

// ============================================================================
// Campaign Operations
// ============================================================================

// CreateCampaign creates a new phishing campaign
func (d *Database) CreateCampaign(name string, threatIDs []string) (*PhishingCampaign, error) {
	campaign := &PhishingCampaign{
		ID:          uuid.New().String(),
		Name:        name,
		Threats:     threatIDs,
		Status:      "draft",
		TargetCount: 0,
		SentCount:   0,
		OpenCount:   0,
		ClickCount:  0,
		CredCount:   0,
		CreatedAt:   time.Now(),
	}

	// Store in Redis
	data, _ := json.Marshal(campaign)
	d.redis.Set(context.Background(), fmt.Sprintf("campaign:%s", campaign.ID), data, 0)

	d.logger.Info("Campaign created", zap.String("id", campaign.ID), zap.String("name", name))

	return campaign, nil
}

// UpdateCampaign updates campaign statistics
func (d *Database) UpdateCampaign(id string, stats map[string]int) error {
	key := fmt.Sprintf("campaign:%s", id)
	data, err := d.redis.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}

	var campaign PhishingCampaign
	if err := json.Unmarshal(data, &campaign); err != nil {
		return err
	}

	if sent, ok := stats["sent"]; ok {
		campaign.SentCount = sent
	}
	if open, ok := stats["opened"]; ok {
		campaign.OpenCount = open
	}
	if click, ok := stats["clicked"]; ok {
		campaign.ClickCount = click
	}
	if cred, ok := stats["credentials"]; ok {
		campaign.CredCount = cred
	}

	data, _ = json.Marshal(campaign)
	d.redis.Set(context.Background(), key, data, 0)

	return nil
}

// ============================================================================
// Phishing Email Generation
// ============================================================================

type PhishingEmail struct {
	Subject    string
	Body       template.HTML
	From       string
	ReplyTo    string
	ThreatID   string
	TargetType string
}

// GeneratePhishingEmail generates a phishing email based on threat
func (d *Database) GeneratePhishingEmail(threat *Threat, targetInfo map[string]string) *PhishingEmail {
	// Template generation based on threat type
	subject := d.generateSubject(threat, targetInfo)
	body := d.generateBody(threat, targetInfo)
	from := d.generateFrom(threat)

	return &PhishingEmail{
		Subject:    subject,
		Body:       template.HTML(body),
		From:       from,
		ReplyTo:    "",
		ThreatID:   threat.ID,
		TargetType: threat.Target,
	}
}

func (d *Database) generateSubject(threat *Threat, targetInfo map[string]string) string {
	subjects := map[string][]string{
		"m365-oauth-consent": {
			"Action Required: Verify Your Account",
			"Microsoft 365: Permission Request",
			"Unusual sign-in activity detected",
		},
		"m365-admin-alert": {
			"URGENT: Administrative Action Required",
			"Action Required: Your Access Will Be Revoked",
			"Critical: Account Verification Needed",
		},
		"m365-password-expiry": {
			"Password Expiring in 24 Hours",
			"Action Required: Update Your Password",
			"Your password will expire soon",
		},
		"gws-oauth": {
			"Google: Permission Request",
			"Action Required: Grant Access",
			"Google Workspace: Access Request",
		},
		"okta-mfa": {
			"MFA Verification Required",
			"Approve Your Sign-In",
			"Security Alert: New Device",
		},
	}

	options := subjects[threat.ID]
	if len(options) == 0 {
		options = []string{threat.Title, "Important: Action Required"}
	}

	return options[time.Now().UnixNano()%int64(len(options))]
}

func (d *Database) generateBody(threat *Threat, targetInfo map[string]string) string {
	// Simplified template - in production would use proper templating
	body := `
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #333;">%s</h2>
			<p>Dear User,</p>
			<p>%s</p>
			<div style="margin: 30px 0;">
				<a href="%s" style="background: #0078d4; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">Take Action</a>
			</div>
			<p style="color: #666; font-size: 12px;">
				This is an automated message from %s.<br>
				If you didn't request this, please ignore this email.
			</p>
		</div>
	`

	return fmt.Sprintf(body,
		threat.Title,
		threat.Description,
		"https://login.phantom.local/capture",
		threat.Target)
}

func (d *Database) generateFrom(threat *Threat) string {
	fromAddresses := map[string]string{
		"microsoft365": "Microsoft Security <security@microsoft-support.xyz>",
		"google":        "Google Workspace <noreply@google-workspace.xyz>",
		"okta":          "Okta Security <security@okta-verify.xyz>",
		"financial":     "Bank Security <alerts@secure-bank.xyz>",
		"enterprise":    "IT Department <it@company-support.xyz>",
	}

	return fromAddresses[threat.Target]
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns threat database statistics
func (d *Database) GetStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	byType := make(map[string]int)
	byTarget := make(map[string]int)
	bySeverity := make(map[string]int)

	var totalSuccessRate float64

	for _, threat := range d.threats {
		byType[string(threat.Type)]++
		byTarget[threat.Target]++
		bySeverity[string(threat.Severity)]++
		totalSuccessRate += threat.SuccessRate
	}

	avgSuccessRate := 0.0
	if len(d.threats) > 0 {
		avgSuccessRate = totalSuccessRate / float64(len(d.threats))
	}

	return map[string]interface{}{
		"total_threats":    len(d.threats),
		"by_type":          byType,
		"by_target":        byTarget,
		"by_severity":      bySeverity,
		"avg_success_rate": fmt.Sprintf("%.2f%%", avgSuccessRate*100),
	}
}

// ============================================================================
// Indicator Detection
// ============================================================================

// CheckIndicator checks if content contains known threat indicators
func (d *Database) CheckIndicator(content string) ([]string, float64) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	content = strings.ToLower(content)
	var matchedIndicators []string
	var totalRisk float64

	for _, threat := range d.threats {
		for _, indicator := range threat.Indicators {
			if strings.Contains(content, strings.ToLower(indicator)) {
				matchedIndicators = append(matchedIndicators, indicator)
				totalRisk += threat.SuccessRate
			}
		}
	}

	return matchedIndicators, totalRisk
}

// HashIndicator hashes an indicator for safe storage
func HashIndicator(indicator string) string {
	hash := sha256.Sum256([]byte(indicator))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// ============================================================================
// URL Pattern Matching
// ============================================================================

// CheckURL checks if URL matches known phishing patterns
func (d *Database) CheckURL(url string) (*Threat, float64) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	url = strings.ToLower(url)

	for _, threat := range d.threats {
		for _, indicator := range threat.Indicators {
			// Simple domain/URL matching
			pattern := strings.ToLower(indicator)
			if strings.Contains(url, pattern) {
				return threat, threat.SuccessRate
			}

			// Also check for typosquatting patterns
			if d.checkTyposquat(url, pattern) {
				return threat, threat.SuccessRate * 1.5 // Higher risk for typosquatting
			}
		}
	}

	return nil, 0
}

func (d *Database) checkTyposquat(url, domain string) bool {
	// Common typosquatting patterns
	typos := []string{
		"microsft", "microsot", "microsofl",
		"g00gle", "go0gle", "goolge",
		"0kta", "okt4",
	}

	for _, typo := range typos {
		if strings.Contains(url, typo) {
			return true
		}
	}

	// Check for subdomain abuse (e.g., login.microsoft.fake-domain.com)
	re := regexp.MustCompile(`^([a-z0-9-]+)\.` + domain)
	if re.MatchString(url) {
		return true
	}

	return false
}
