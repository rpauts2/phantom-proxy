//go:build ignore
// +build ignore

// Captcha solver временно отключен из-за изменений в playwright-go API
// Требуется рефакторинг для работы с новыми типами

package playwright

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"go.uber.org/zap"
)

// CaptchaSolver решатель капч через Playwright
type CaptchaSolver struct {
	mu            sync.RWMutex
	pw            *playwright.Playwright
	browser       playwright.Browser
	contextPool   chan playwright.BrowserContext
	poolSize      int
	logger        *zap.Logger
	stealthScript string
}

// SolverConfig конфигурация решателя
type SolverConfig struct {
	Headless    bool
	PoolSize    int
	UserAgent   string
	Viewport    *Viewport
	Timeout     time.Duration
}

// Viewport размер окна браузера
type Viewport struct {
	Width  int
	Height int
}

// CaptchaResult результат решения капчи
type CaptchaResult struct {
	Token     string
	ExpiredAt time.Time
	Success   bool
	Error     error
}

// DefaultConfig конфигурация по умолчанию
func DefaultConfig() *SolverConfig {
	return &SolverConfig{
		Headless:  false, // Headful для лучшего обхода детекта
		PoolSize:  3,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		Viewport: &Viewport{
			Width:  1920,
			Height: 1080,
		},
		Timeout: 120 * time.Second,
	}
}

// NewCaptchaSolver создаёт новый решатель капч
func NewCaptchaSolver(logger *zap.Logger, config *SolverConfig) (*CaptchaSolver, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Запуск Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to launch playwright: %w", err)
	}

	// Запуск браузера
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(config.Headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-web-security",
			"--disable-features=IsolateOrigins,site-per-process",
		},
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Создание пула контекстов
	contextPool := make(chan playwright.BrowserContext, config.PoolSize)
	for i := 0; i < config.PoolSize; i++ {
		ctx, err := browser.NewContext(playwright.BrowserNewContextOptions{
			UserAgent: playwright.String(config.UserAgent),
			Viewport: &playwright.Size{
				Width:  config.Viewport.Width,
				Height: config.Viewport.Height,
			},
			IsEnabled:   playwright.Bool(true),
			Screen: &playwright.Size{
				Width: config.Viewport.Width,
				Height: config.Viewport.Height,
			},
		})
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("failed to create browser context: %w", err)
		}
		contextPool <- ctx
	}

	s := &CaptchaSolver{
		pw:            pw,
		browser:       browser,
		contextPool:   contextPool,
		poolSize:      config.PoolSize,
		logger:        logger,
		stealthScript: getStealthScript(),
	}

	logger.Info("CaptchaSolver initialized",
		zap.Int("pool_size", config.PoolSize),
		zap.Bool("headless", config.Headless))

	return s, nil
}

// SolveReCAPTCHA решает reCAPTCHA v2/v3
func (s *CaptchaSolver) SolveReCAPTCHA(pageURL, siteKey string) (*CaptchaResult, error) {
	start := time.Now()
	s.logger.Info("Solving reCAPTCHA",
		zap.String("url", pageURL),
		zap.String("sitekey", siteKey))

	// Получение контекста из пула
	ctx := <-s.contextPool
	defer func() { s.contextPool <- ctx }()

	page, err := ctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Инъекция stealth скриптов
	if err := s.injectStealth(page); err != nil {
		s.logger.Warn("Failed to inject stealth script", zap.Error(err))
	}

	// Переход на страницу
	if _, err := page.Goto(pageURL); err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	s.logger.Debug("Page loaded, waiting for reCAPTCHA")

	// Ожидание iframe reCAPTCHA
	frameSelector := fmt.Sprintf("iframe[src*='recaptcha'][src*='%s']", siteKey)

	// Ждём появления фрейма
	var captchaFrame playwright.Frame
	for i := 0; i < 30; i++ {
		frames := page.Frames()
		for _, f := range frames {
			if strings.Contains(f.URL(), "recaptcha") && strings.Contains(f.URL(), siteKey) {
				captchaFrame = f
				break
			}
		}
		if captchaFrame != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if captchaFrame == nil {
		return &CaptchaResult{
			Success: false,
			Error:   fmt.Errorf("reCAPTCHA iframe not found after 30s"),
		}, fmt.Errorf("iframe not found")
	}

	s.logger.Debug("reCAPTCHA iframe found, clicking checkbox")

	// Клик по чекбоксу
	checkboxSelector := "#recaptcha-anchor"
	if err := captchaFrame.Click(checkboxSelector); err != nil {
		return &CaptchaResult{
			Success: false,
			Error:   fmt.Errorf("failed to click checkbox: %w", err),
		}, err
	}

	s.logger.Debug("Waiting for reCAPTCHA token")

	// Ожидание токена
	token, err := s.waitForToken(page, siteKey)
	if err != nil {
		return &CaptchaResult{
			Success: false,
			Error:   err,
		}, err
	}

	elapsed := time.Since(start)
	s.logger.Info("reCAPTCHA solved",
		zap.Duration("duration", elapsed),
		zap.String("token_prefix", token[:20]))

	return &CaptchaResult{
		Token:     token,
		Success:   true,
		ExpiredAt: start.Add(120 * time.Second), // reCAPTCHA токены живут 2 минуты
	}, nil
}

// SolveHCaptcha решает hCaptcha
func (s *CaptchaSolver) SolveHCaptcha(pageURL, siteKey string) (*CaptchaResult, error) {
	start := time.Now()
	s.logger.Info("Solving hCaptcha",
		zap.String("url", pageURL),
		zap.String("sitekey", siteKey))

	ctx := <-s.contextPool
	defer func() { s.contextPool <- ctx }()

	page, err := ctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	if err := s.injectStealth(page); err != nil {
		s.logger.Warn("Failed to inject stealth script", zap.Error(err))
	}

	if _, err := page.Goto(pageURL); err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	// Ожидание iframe hCaptcha
	frameSelector := fmt.Sprintf("iframe[src*='hcaptcha'][src*='%s']", siteKey)

	// Ждём появления фрейма
	var captchaFrame playwright.Frame
	for i := 0; i < 30; i++ {
		frames := page.Frames()
		for _, f := range frames {
			if strings.Contains(f.URL(), "hcaptcha") && strings.Contains(f.URL(), siteKey) {
				captchaFrame = f
				break
			}
		}
		if captchaFrame != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if captchaFrame == nil {
		return &CaptchaResult{
			Success: false,
			Error:   fmt.Errorf("hCaptcha iframe not found after 30s"),
		}, fmt.Errorf("iframe not found")
	}

	// Клик по чекбоксу
	if err := captchaFrame.Click("#checkbox"); err != nil {
		return &CaptchaResult{
			Success: false,
			Error:   fmt.Errorf("failed to click checkbox: %w", err),
		}, err
	}

	token, err := s.waitForHCaptchaToken(page)
	if err != nil {
		return &CaptchaResult{
			Success: false,
			Error:   err,
		}, err
	}

	s.logger.Info("hCaptcha solved",
		zap.Duration("duration", time.Since(start)))

	return &CaptchaResult{
		Token:     token,
		Success:   true,
		ExpiredAt: start.Add(120 * time.Second),
	}, nil
}

// waitForToken ожидает токен reCAPTCHA
func (s *CaptchaSolver) waitForToken(page playwright.Page, siteKey string) (string, error) {
	timeout := time.After(120 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for reCAPTCHA token")
		case <-ticker.C:
			token, err := page.Evaluate(fmt.Sprintf(`() => {
				const textarea = document.querySelector('textarea[name="g-recaptcha-response"]');
				return textarea ? textarea.value : null;
			}`))
			if err != nil {
				continue
			}

			if tokenStr, ok := token.(string); ok && tokenStr != "" {
				return tokenStr, nil
			}
		}
	}
}

// waitForHCaptchaToken ожидает токен hCaptcha
func (s *CaptchaSolver) waitForHCaptchaToken(page playwright.Page) (string, error) {
	timeout := time.After(120 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for hCaptcha token")
		case <-ticker.C:
			token, err := page.Evaluate(`() => {
				const textarea = document.querySelector('[name="h-captcha-response"]');
				return textarea ? textarea.value : null;
			}`)
			if err != nil {
				continue
			}

			if tokenStr, ok := token.(string); ok && tokenStr != "" {
				return tokenStr, nil
			}
		}
	}
}

// injectStealth внедряет скрипты для скрытия headless
func (s *CaptchaSolver) injectStealth(page playwright.Page) error {
	_, err := page.AddInitScript(s.stealthScript)
	return err
}

// Close закрывает решатель капч
func (s *CaptchaSolver) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Закрытие всех контекстов
	close(s.contextPool)
	for ctx := range s.contextPool {
		ctx.Close()
	}

	if s.browser != nil {
		s.browser.Close()
	}

	if s.pw != nil {
		s.pw.Stop()
	}

	s.logger.Info("CaptchaSolver closed")
	return nil
}

// getStealthScript возвращает скрипт для скрытия headless
func getStealthScript() string {
	return `
// Скрытие webdriver флага
Object.defineProperty(navigator, 'webdriver', {
	get: () => false
});

// Подмена plugins
Object.defineProperty(navigator, 'plugins', {
	get: () => [1, 2, 3, 4, 5]
});

// Подмена languages
Object.defineProperty(navigator, 'languages', {
	get: () => ['en-US', 'en']
});

// Добавление chrome объекта
window.chrome = {
	runtime: {},
	loadTimes: function() {},
	csi: function() {}
};

// Подмена permissions
const originalQuery = window.navigator.permissions.query;
window.navigator.permissions.query = (parameters) => (
	parameters.name === 'notifications' ?
		Promise.resolve({ state: Notification.permission }) :
		originalQuery(parameters)
);

// Скрытие headless признаков
Object.defineProperty(navigator, 'userAgent', {
	get: () => 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36'
});

// Удаление navigator.webdriver
delete navigator.__proto__.webdriver;

// Подмена appVersion
Object.defineProperty(navigator, 'appVersion', {
	get: () => '5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36'
});

// Подмена platform
Object.defineProperty(navigator, 'platform', {
	get: () => 'Win32'
});

// Подмена hardwareConcurrency
Object.defineProperty(navigator, 'hardwareConcurrency', {
	get: () => 8
});

// Подмена deviceMemory
Object.defineProperty(navigator, 'deviceMemory', {
	get: () => 8
});

// Подмена maxTouchPoints
Object.defineProperty(navigator, 'maxTouchPoints', {
	get: () => 0
});

// Генерация случайных отпечатков
const originalGetPropertyValue = CSSStyleDeclaration.prototype.getPropertyValue;
CSSStyleDeclaration.prototype.getPropertyValue = function(property) {
	if (property === '--headless') {
		return '';
	}
	return originalGetPropertyValue.call(this, property);
};
`
}

// GetStats возвращает статистику решателя
func (s *CaptchaSolver) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"pool_size":   s.poolSize,
		"available":   len(s.contextPool),
		"browser":     "chromium",
		"headless":    false,
	}
}
