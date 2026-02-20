package ml

import (
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// BotDetector ML-детектор ботов
type BotDetector struct {
	mu              sync.RWMutex
	enabled         bool
	threshold       float32
	learningMode    bool
	logger          *zap.Logger
	
	// Статистика
	totalRequests   int64
	botDetections   int64
	humanDetections int64
	
	// Паттерны для обнаружения
	botPatterns     []string
	suspiciousUAs   map[string]int
	ipRequestCounts map[string]int
}

// DetectionResult результат детекта
type DetectionResult struct {
	IsBot      bool
	Confidence float32
	Reason     string
	Features   map[string]interface{}
}

// NewBotDetector создаёт новый детектор
func NewBotDetector(logger *zap.Logger, threshold float32) *BotDetector {
	return &BotDetector{
		enabled:         true,
		threshold:       threshold,
		learningMode:    true,
		logger:          logger,
		botPatterns: []string{
			"bot", "spider", "crawler", "scraper",
			"headless", "phantom", "selenium", "puppeteer",
			"python", "curl", "wget", "go-http",
		},
		suspiciousUAs:   make(map[string]int),
		ipRequestCounts: make(map[string]int),
	}
}

// Detect анализирует запрос
func (d *BotDetector) Detect(ua string, ip string, headers map[string]string) *DetectionResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.totalRequests++
	
	result := &DetectionResult{
		IsBot:      false,
		Confidence: 0,
		Reason:     "human",
		Features:   make(map[string]interface{}),
	}
	
	score := 0.0
	reasons := []string{}
	
	// 1. Анализ User-Agent
	uaLower := strings.ToLower(ua)
	for _, pattern := range d.botPatterns {
		if strings.Contains(uaLower, pattern) {
			score += 0.3
			reasons = append(reasons, fmt.Sprintf("bot_pattern:%s", pattern))
		}
	}
	
	// 2. Проверка на отсутствие стандартных заголовков браузера
	if headers["Accept-Language"] == "" {
		score += 0.15
		reasons = append(reasons, "no_accept_language")
	}
	
	if headers["Accept-Encoding"] == "" {
		score += 0.1
		reasons = append(reasons, "no_accept_encoding")
	}
	
	// 3. Проверка на headless браузеры
	if strings.Contains(uaLower, "headless") || 
	   strings.Contains(uaLower, "phantom") ||
	   strings.Contains(uaLower, "selenium") {
		score += 0.5
		reasons = append(reasons, "headless_browser")
	}
	
	// 4. Rate limiting анализ
	d.ipRequestCounts[ip]++
	if d.ipRequestCounts[ip] > 100 {
		score += 0.2
		reasons = append(reasons, "high_request_rate")
	}
	
	// 5. Проверка порядка заголовков (браузеры отправляют в определённом порядке)
	if !d.checkHeaderOrder(headers) {
		score += 0.15
		reasons = append(reasons, "suspicious_header_order")
	}
	
	// 6. Анализ временных паттернов (можно добавить позже)
	
	result.Confidence = float32(score)
	result.Reason = strings.Join(reasons, ",")
	
	if score >= float64(d.threshold) {
		result.IsBot = true
		d.botDetections++
		d.logger.Warn("Bot detected",
			zap.String("ip", ip),
			zap.String("reason", result.Reason),
			zap.Float32("confidence", result.Confidence))
	} else {
		result.IsBot = false
		d.humanDetections++
	}
	
	// Learning mode - запоминаем паттерны
	if d.learningMode {
		d.learn(ua, ip, result.IsBot)
	}
	
	return result
}

// checkHeaderOrder проверяет порядок заголовков
func (d *BotDetector) checkHeaderOrder(headers map[string]string) bool {
	// Браузеры обычно отправляют заголовки в определённом порядке
	// Простая эвристика: наличие ключевых заголовков
	requiredHeaders := []string{
		"Host",
		"Connection",
		"Accept",
		"User-Agent",
	}
	
	for _, h := range requiredHeaders {
		if _, ok := headers[h]; !ok {
			return false
		}
	}
	
	return true
}

// learn запоминает паттерны для улучшения модели
func (d *BotDetector) learn(ua string, ip string, isBot bool) {
	if isBot {
		d.suspiciousUAs[ua]++
	}
}

// GetStats возвращает статистику
func (d *BotDetector) GetStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	botPercentage := 0.0
	if d.totalRequests > 0 {
		botPercentage = float64(d.botDetections) / float64(d.totalRequests) * 100
	}
	
	return map[string]interface{}{
		"enabled":          d.enabled,
		"threshold":        d.threshold,
		"learning_mode":    d.learningMode,
		"total_requests":   d.totalRequests,
		"bot_detections":   d.botDetections,
		"human_detections": d.humanDetections,
		"bot_percentage":   fmt.Sprintf("%.2f%%", botPercentage),
	}
}

// SetThreshold обновляет порог
func (d *BotDetector) SetThreshold(threshold float32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.threshold = threshold
	d.logger.Info("Bot detector threshold updated",
		zap.Float32("threshold", threshold))
}

// Enable включает детектор
func (d *BotDetector) Enable() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.enabled = true
	d.logger.Info("Bot detector enabled")
}

// Disable выключает детектор
func (d *BotDetector) Disable() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.enabled = false
	d.logger.Info("Bot detector disabled")
}

// AddBotPattern добавляет паттерн для обнаружения
func (d *BotDetector) AddBotPattern(pattern string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.botPatterns = append(d.botPatterns, pattern)
	d.logger.Info("Bot pattern added", zap.String("pattern", pattern))
}
