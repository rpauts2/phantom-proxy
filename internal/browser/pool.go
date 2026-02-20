package browser

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"go.uber.org/zap"
)

// BrowserPool пул браузеров для эмуляции человека
type BrowserPool struct {
	mu          sync.RWMutex
	logger      *zap.Logger
	config      *PoolConfig
	browsers    []*BrowserInstance
	currentIdx  int
	requestChan chan *Request
}

// PoolConfig конфигурация пула
type PoolConfig struct {
	// Размер пула
	MinBrowsers int
	MaxBrowsers int
	
	// Таймауты
	BrowserTimeout time.Duration
	PageTimeout    time.Duration
	
	// Поведение
	HumanizeActions bool
	RandomDelays    bool
	
	// Fingerprints
	RandomUserAgent   bool
	RandomViewport    bool
	RandomTimezone    bool
	
	// Playwright
	Headless bool
	Args     []string
}

// BrowserInstance экземпляр браузера
type BrowserInstance struct {
	ID           string
	Browser      playwright.Browser
	Context      playwright.BrowserContext
	Page         playwright.Page
	LastUsed     time.Time
	RequestCount int
	IsActive     bool
	UserAgent    string
	Viewport     *Viewport
	Timezone     string
}

// Viewport размер окна
type Viewport struct {
	Width  int
	Height int
}

// Request запрос на выполнение
type Request struct {
	URL      string
	Method   string
	Headers  map[string]string
	Body     string
	Response chan *Response
	Error    chan error
}

// Response ответ
type Response struct {
	Status  int
	Headers map[string]string
	Body    string
	Screenshot []byte
}

// DefaultConfig конфигурация по умолчанию
func DefaultConfig() *PoolConfig {
	return &PoolConfig{
		MinBrowsers:     2,
		MaxBrowsers:     10,
		BrowserTimeout:  30 * time.Minute,
		PageTimeout:     60 * time.Second,
		HumanizeActions: true,
		RandomDelays:    true,
		RandomUserAgent: true,
		RandomViewport:  true,
		RandomTimezone:  true,
		Headless:        true, // Для продакшена false
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-web-security",
			"--disable-features=IsolateOrigins,site-per-process",
		},
	}
}

// NewBrowserPool создаёт новый пул
func NewBrowserPool(config *PoolConfig, logger *zap.Logger) (*BrowserPool, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	if config.MinBrowsers < 1 {
		config.MinBrowsers = 1
	}
	if config.MaxBrowsers < config.MinBrowsers {
		config.MaxBrowsers = config.MinBrowsers
	}
	
	pool := &BrowserPool{
		logger:      logger,
		config:      config,
		browsers:    make([]*BrowserInstance, 0),
		requestChan: make(chan *Request, 100),
	}
	
	// Инициализация Playwright
	if err := playwright.Install(); err != nil {
		return nil, fmt.Errorf("failed to install playwright: %w", err)
	}
	
	logger.Info("Browser pool initialized",
		zap.Int("min_browsers", config.MinBrowsers),
		zap.Int("max_browsers", config.MaxBrowsers))
	
	return pool, nil
}

// Start запускает пул
func (p *BrowserPool) Start(ctx context.Context) error {
	p.logger.Info("Starting browser pool")
	
	// Создание минимального количества браузеров
	for i := 0; i < p.config.MinBrowsers; i++ {
		if err := p.createBrowser(ctx); err != nil {
			p.logger.Error("Failed to create browser", zap.Error(err))
			return err
		}
	}
	
	// Фоновые задачи
	go p.healthWorker(ctx)
	go p.requestWorker(ctx)
	go p.scalingWorker(ctx)
	
	return nil
}

// createBrowser создаёт новый браузер
func (p *BrowserPool) createBrowser(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.browsers) >= p.config.MaxBrowsers {
		return fmt.Errorf("max browsers limit reached")
	}
	
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright: %w", err)
	}
	
	// Генерация fingerprint
	userAgent := p.generateUserAgent()
	viewport := p.generateViewport()
	timezone := p.generateTimezone()
	
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(p.config.Headless),
		Args:     p.config.Args,
	})
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent:   playwright.String(userAgent),
		Viewport:    &playwright.ViewportSize{Width: viewport.Width, Height: viewport.Height},
		TimezoneID:  playwright.String(timezone),
		Locale:      playwright.String("en-US"),
		JavaScript:  playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}
	
	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	
	// Инъекция stealth скриптов
	if err := p.injectStealth(page); err != nil {
		p.logger.Warn("Failed to inject stealth", zap.Error(err))
	}
	
	instance := &BrowserInstance{
		ID:        fmt.Sprintf("browser-%d", len(p.browsers)),
		Browser:   browser,
		Context:   context,
		Page:      page,
		LastUsed:  time.Now(),
		IsActive:  true,
		UserAgent: userAgent,
		Viewport:  viewport,
		Timezone:  timezone,
	}
	
	p.browsers = append(p.browsers, instance)
	
	p.logger.Info("Browser created",
		zap.String("id", instance.ID),
		zap.String("user_agent", userAgent),
		zap.Int("viewport_width", viewport.Width),
		zap.String("timezone", timezone))
	
	return nil
}

// Execute выполняет запрос через браузер
func (p *BrowserPool) Execute(ctx context.Context, url string, method string, headers map[string]string, body string) (*Response, error) {
	browser := p.getNextBrowser()
	if browser == nil {
		return nil, fmt.Errorf("no available browsers")
	}
	
	p.logger.Debug("Executing request",
		zap.String("browser_id", browser.ID),
		zap.String("url", url))
	
	// Human-like задержка
	if p.config.RandomDelays {
		delay := time.Duration(rand.Intn(2000)+1000) * time.Millisecond
		time.Sleep(delay)
	}
	
	// Навигация
	page := browser.Page
	
	// Установка заголовков
	if headers != nil {
		// TODO: Установка заголовков
	}
	
	// Выполнение запроса
	var response *playwright.Response
	var err error
	
	switch method {
	case "GET":
		response, err = page.Goto(url, playwright.PageGotoOptions{
			Timeout: playwright.Float64(float64(p.config.PageTimeout.Milliseconds())),
		})
	case "POST":
		response, err = page.Post(url, body, playwright.PagePostOptions{
			Timeout: playwright.Float64(float64(p.config.PageTimeout.Milliseconds())),
		})
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
	
	if err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}
	
	// Human-like поведение после загрузки
	if p.config.HumanizeActions {
		go p.humanBehavior(browser.Page)
	}
	
	// Получение ответа
	status := response.Status()
	respHeaders := make(map[string]string)
	for k, v := range response.Headers() {
		respHeaders[k] = v
	}
	
	respBody, err := response.Body()
	if err != nil {
		return nil, err
	}
	
	browser.LastUsed = time.Now()
	browser.RequestCount++
	
	return &Response{
		Status:  status,
		Headers: respHeaders,
		Body:    string(respBody),
	}, nil
}

// Screenshot делает скриншот страницы
func (p *BrowserPool) Screenshot(browserID string) ([]byte, error) {
	browser := p.getBrowserByID(browserID)
	if browser == nil {
		return nil, fmt.Errorf("browser not found")
	}
	
	screenshot, err := browser.Page.Screenshot()
	if err != nil {
		return nil, err
	}
	
	return screenshot, nil
}

// getNextBrowser получает следующий доступный браузер
func (p *BrowserPool) getNextBrowser() *BrowserInstance {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if len(p.browsers) == 0 {
		return nil
	}
	
	// Round-robin
	browser := p.browsers[p.currentIdx]
	p.currentIdx = (p.currentIdx + 1) % len(p.browsers)
	
	return browser
}

// getBrowserByID получает браузер по ID
func (p *BrowserPool) getBrowserByID(id string) *BrowserInstance {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	for _, browser := range p.browsers {
		if browser.ID == id {
			return browser
		}
	}
	
	return nil
}

// injectStealth внедряет stealth скрипты
func (p *BrowserPool) injectStealth(page playwright.Page) error {
	stealthScript := `
	// Скрытие webdriver
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
	
	// Добавление chrome
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
	`
	
	_, err := page.AddInitScript(playwright.PageAddInitScriptOptions{
		Content: playwright.String(stealthScript),
	})
	
	return err
}

// humanBehavior эмулирует человеческое поведение
func (p *BrowserPool) humanBehavior(page playwright.Page) {
	// Случайные движения мыши
	go func() {
		for i := 0; i < rand.Intn(5)+3; i++ {
			x := rand.Intn(1000)
			y := rand.Intn(800)
			page.Mouse.Move(float64(x), float64(y))
			time.Sleep(time.Duration(rand.Intn(500)+200) * time.Millisecond)
		}
	}()
	
	// Случайные клики
	go func() {
		time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
		x := rand.Intn(800)
		y := rand.Intn(600)
		page.Mouse.Click(float64(x), float64(y))
	}()
	
	// Скроллинг
	go func() {
		for i := 0; i < rand.Intn(3)+1; i++ {
			page.Evaluate(fmt.Sprintf("window.scrollBy(0, %d)", rand.Intn(500)+100))
			time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
		}
	}()
}

// generateUserAgent генерирует случайный User-Agent
func (p *BrowserPool) generateUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
	}
	
	return userAgents[rand.Intn(len(userAgents))]
}

// generateViewport генерирует случайный viewport
func (p *BrowserPool) generateViewport() *Viewport {
	viewports := []Viewport{
		{Width: 1920, Height: 1080},
		{Width: 1366, Height: 768},
		{Width: 1536, Height: 864},
		{Width: 1440, Height: 900},
		{Width: 1280, Height: 720},
	}
	
	return &viewports[rand.Intn(len(viewports))]
}

// generateTimezone генерирует случайную timezone
func (p *BrowserPool) generateTimezone() string {
	timezones := []string{
		"America/New_York",
		"America/Chicago",
		"America/Los_Angeles",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
	}
	
	return timezones[rand.Intn(len(timezones))]
}

// healthWorker фоновая задача проверки здоровья
func (p *BrowserPool) healthWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkHealth()
		}
	}
}

// checkHealth проверяет здоровье браузеров
func (p *BrowserPool) checkHealth() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	for _, browser := range p.browsers {
		if !browser.IsActive {
			continue
		}
		
		// Проверка на таймаут
		if time.Since(browser.LastUsed) > p.config.BrowserTimeout {
			p.logger.Info("Browser timeout",
				zap.String("id", browser.ID),
				zap.Duration("idle", time.Since(browser.LastUsed)))
			// TODO: Пересоздание браузера
		}
	}
}

// requestWorker обрабатывает запросы
func (p *BrowserPool) requestWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-p.requestChan:
			// Обработка запроса
			// TODO: Реализация
		}
	}
}

// scalingWorker масштабирует пул
func (p *BrowserPool) scalingWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.scale()
		}
	}
}

// scale масштабирует пул
func (p *BrowserPool) scale() {
	// TODO: Логика масштабирования
}

// Stop останавливает пул
func (p *BrowserPool) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, browser := range p.browsers {
		if browser.Page != nil {
			browser.Page.Close()
		}
		if browser.Context != nil {
			browser.Context.Close()
		}
		if browser.Browser != nil {
			browser.Browser.Close()
		}
	}
	
	p.logger.Info("Browser pool stopped")
	
	return nil
}

// GetStats возвращает статистику
func (p *BrowserPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	totalRequests := 0
	for _, browser := range p.browsers {
		totalRequests += browser.RequestCount
	}
	
	return map[string]interface{}{
		"total_browsers":  len(p.browsers),
		"active_browsers": len(p.browsers),
		"total_requests":  totalRequests,
		"min_browsers":    p.config.MinBrowsers,
		"max_browsers":    p.config.MaxBrowsers,
	}
}
