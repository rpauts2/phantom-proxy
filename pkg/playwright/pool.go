// PhantomProxy v14.0 - Playwright Browser Pool
// Package playwright provides browser automation for captcha solving
package playwright

import (
	"fmt"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"go.uber.org/zap"
)

// BrowserPool manages browser instances
type BrowserPool struct {
	mu        sync.RWMutex
	config    *Config
	logger    *zap.Logger
	browsers  []*BrowserInstance
	maxSize   int
	current   int
}

// BrowserInstance представляет экземпляр браузера
type BrowserInstance struct {
	ID        string
	Browser   playwright.Browser
	Context   playwright.BrowserContext
	Page      playwright.Page
	InUse     bool
	LastUsed  time.Time
	CreatedAt time.Time
}

// Config конфигурация пула
type Config struct {
	Headless    bool
	PoolSize    int
	Timeout     time.Duration
	UserAgent   string
	Proxy       string
	Stealth     bool
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Headless:  true,
		PoolSize:  5,
		Timeout:   30 * time.Second,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		Stealth:   true,
	}
}

// NewBrowserPool создает пул браузеров
func NewBrowserPool(config *Config, logger *zap.Logger) (*BrowserPool, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize Playwright
	err := playwright.Install()
	if err != nil {
		return nil, fmt.Errorf("failed to install playwright: %w", err)
	}

	pool := &BrowserPool{
		config:  config,
		logger:  logger,
		browsers: make([]*BrowserInstance, 0),
		maxSize: config.PoolSize,
	}

	// Pre-warm pool
	if err := pool.warmup(); err != nil {
		logger.Warn("Failed to warmup browser pool", zap.Error(err))
	}

	logger.Info("Browser pool initialized",
		zap.Int("size", config.PoolSize),
		zap.Bool("headless", config.Headless),
		zap.Bool("stealth", config.Stealth))

	return pool, nil
}

// warmup предварительно создает браузеры
func (p *BrowserPool) warmup() error {
	for i := 0; i < p.config.PoolSize; i++ {
		if err := p.createBrowser(); err != nil {
			return err
		}
	}
	return nil
}

// createBrowser создает новый браузер
func (p *BrowserPool) createBrowser() error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(p.config.Headless),
	})
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(p.config.UserAgent),
	})
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	instance := &BrowserInstance{
		ID:        fmt.Sprintf("browser_%d", len(p.browsers)),
		Browser:   browser,
		Context:   context,
		Page:      page,
		CreatedAt: time.Now(),
	}

	p.browsers = append(p.browsers, instance)
	p.logger.Debug("Browser created", zap.String("id", instance.ID))

	return nil
}

// GetBrowser получает свободный браузер
func (p *BrowserPool) GetBrowser() (*BrowserInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Найти свободный
	for _, browser := range p.browsers {
		if !browser.InUse {
			browser.InUse = true
			browser.LastUsed = time.Now()
			return browser, nil
		}
	}

	// Если все заняты - создать новый
	if len(p.browsers) < p.maxSize {
		if err := p.createBrowser(); err != nil {
			return nil, err
		}
		last := p.browsers[len(p.browsers)-1]
		last.InUse = true
		return last, nil
	}

	return nil, fmt.Errorf("browser pool exhausted")
}

// ReleaseBrowser освобождает браузер
func (p *BrowserPool) ReleaseBrowser(instance *BrowserInstance) {
	p.mu.Lock()
	defer p.mu.Unlock()

	instance.InUse = false
	instance.LastUsed = time.Now()
}

// NavigateTo переходит на страницу
func (p *BrowserPool) NavigateTo(url string) (string, error) {
	browser, err := p.GetBrowser()
	if err != nil {
		return "", err
	}
	defer p.ReleaseBrowser(browser)

	_, err = browser.Page.Goto(url)
	if err != nil {
		return "", err
	}

	return browser.Page.Content()
}

// SolveReCAPTCHA решает reCAPTCHA
func (p *BrowserPool) SolveReCAPTCHA(pageURL, siteKey string) (string, error) {
	browser, err := p.GetBrowser()
	if err != nil {
		return "", err
	}
	defer p.ReleaseBrowser(browser)

	// Перейти на страницу
	_, err = browser.Page.Goto(pageURL)
	if err != nil {
		return "", err
	}

	// Найти и кликнуть reCAPTCHA
	frame := browser.Page.FrameLocator("iframe[src*='recaptcha']")
	if frame == nil {
		return "", fmt.Errorf("recaptcha iframe not found")
	}

	// Кликнуть на чекбокс
	err = frame.Locator("input[type='checkbox']").Click()
	if err != nil {
		return "", err
	}

	// Подождать токен
	time.Sleep(2 * time.Second)

	// Получить токен
	token, err := browser.Page.Evaluate(`() => {
		return document.querySelector('[name="g-recaptcha-response"]').value
	}`)

	if err != nil || token == nil {
		return "", fmt.Errorf("failed to get recaptcha token")
	}

	return token.(string), nil
}

// Screenshot делает скриншот
func (p *BrowserPool) Screenshot(url string) ([]byte, error) {
	browser, err := p.GetBrowser()
	if err != nil {
		return nil, err
	}
	defer p.ReleaseBrowser(browser)

	_, err = browser.Page.Goto(url)
	if err != nil {
		return nil, err
	}

	return browser.Page.Screenshot()
}

// Close закрывает пул
func (p *BrowserPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, browser := range p.browsers {
		if err := browser.Browser.Close(); err != nil {
			p.logger.Error("Failed to close browser", zap.Error(err))
		}
	}

	p.logger.Info("Browser pool closed")
	return nil
}

// GetStats возвращает статистику
func (p *BrowserPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inUse := 0
	for _, b := range p.browsers {
		if b.InUse {
			inUse++
		}
	}

	return map[string]interface{}{
		"total_browsers": len(p.browsers),
		"in_use":         inUse,
		"available":      len(p.browsers) - inUse,
		"max_size":       p.maxSize,
	}
}
