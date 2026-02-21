// Package risk - Human Risk Score Calculator
package risk

import (
	"context"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RiskScore оценка риска
type RiskScore struct {
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email"`
	TenantID     string                 `json:"tenant_id"`
	OverallScore float64                `json:"overall_score"`
	RiskLevel    string                 `json:"risk_level"` // low, medium, high, critical
	Factors      map[string]float64     `json:"factors"`
	BehaviorData map[string]interface{} `json:"behavior_data"`
	LastUpdated  time.Time              `json:"last_updated"`
	Trend        string                 `json:"trend"` // improving, stable, worsening
}

// BehaviorEvent событие поведения
type BehaviorEvent struct {
	UserID     string                 `json:"user_id"`
	EventType  string                 `json:"event_type"` // click, submit, hover, time_spent
	EventData  map[string]interface{} `json:"event_data"`
	Timestamp  time.Time              `json:"timestamp"`
	PhishletID string                 `json:"phishlet_id"`
	CampaignID string                 `json:"campaign_id"`
}

// Calculator калькулятор риска
type Calculator struct {
	mu           sync.RWMutex
	logger       *zap.Logger
	userScores   map[string]*RiskScore
	eventHistory map[string][]BehaviorEvent
	config       *Config
}

// Config конфигурация
type Config struct {
	EnableML    bool    `json:"enable_ml"`
	HistorySize int     `json:"history_size"`
	DecayFactor float64 `json:"decay_factor"`
	Thresholds  Thresholds `json:"thresholds"`
	Weights     Weights    `json:"weights"`
}

// Weights веса факторов
type Weights struct {
	ClickSpeed        float64 `json:"click_speed"`
	FormSubmission    float64 `json:"form_submission"`
	HoverPatterns     float64 `json:"hover_patterns"`
	TimeOnPage        float64 `json:"time_on_page"`
	MouseMovement     float64 `json:"mouse_movement"`
	KeyboardPatterns  float64 `json:"keyboard_patterns"`
	PreviousClicks    float64 `json:"previous_clicks"`
	DeviceFingerprint float64 `json:"device_fingerprint"`
}

// Thresholds пороги
type Thresholds struct {
	Low      float64 `json:"low"`      // 0-30
	Medium   float64 `json:"medium"`   // 30-60
	High     float64 `json:"high"`     // 60-80
	Critical float64 `json:"critical"` // 80-100
}

// DefaultConfig конфигурация по умолчанию
func DefaultConfig() *Config {
	return &Config{
		EnableML:    false,
		HistorySize: 100,
		DecayFactor: 0.95,
		Weights: Weights{
			ClickSpeed:        0.15,
			FormSubmission:    0.20,
			HoverPatterns:     0.10,
			TimeOnPage:        0.10,
			MouseMovement:     0.10,
			KeyboardPatterns:  0.15,
			PreviousClicks:    0.10,
			DeviceFingerprint: 0.10,
		},
		Thresholds: Thresholds{
			Low:      30,
			Medium:   60,
			High:     80,
			Critical: 80,
		},
	}
}

// NewCalculator создает калькулятор риска
func NewCalculator(logger *zap.Logger, config *Config) *Calculator {
	if config == nil {
		config = DefaultConfig()
	}

	return &Calculator{
		logger:       logger,
		userScores:   make(map[string]*RiskScore),
		eventHistory: make(map[string][]BehaviorEvent),
		config:       config,
	}
}

// ProcessEvent обрабатывает событие
func (c *Calculator) ProcessEvent(ctx context.Context, event BehaviorEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Добавить в историю
	if _, ok := c.eventHistory[event.UserID]; !ok {
		c.eventHistory[event.UserID] = make([]BehaviorEvent, 0)
	}

	history := c.eventHistory[event.UserID]
	history = append(history, event)

	// Ограничить размер
	if len(history) > c.config.HistorySize {
		history = history[len(history)-c.config.HistorySize:]
	}
	c.eventHistory[event.UserID] = history

	// Пересчитать риск
	c.recalculateRisk(event.UserID)

	return nil
}

// GetRiskScore возвращает оценку риска
func (c *Calculator) GetRiskScore(userID string) *RiskScore {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if score, ok := c.userScores[userID]; ok {
		return score.clone()
	}

	return &RiskScore{
		UserID:       userID,
		OverallScore: 50,
		RiskLevel:    "medium",
		Factors:      make(map[string]float64),
		LastUpdated:  time.Now(),
		Trend:        "stable",
	}
}

// GetAllScores возвращает все оценки
func (c *Calculator) GetAllScores() map[string]*RiskScore {
	c.mu.RLock()
	defer c.mu.RUnlock()

	scores := make(map[string]*RiskScore)
	for k, v := range c.userScores {
		scores[k] = v.clone()
	}
	return scores
}

// GetRiskDistribution возвращает распределение
func (c *Calculator) GetRiskDistribution() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dist := map[string]int{
		"low":      0,
		"medium":   0,
		"high":     0,
		"critical": 0,
	}

	for _, score := range c.userScores {
		dist[score.RiskLevel]++
	}

	return dist
}

// GetHighRiskUsers возвращает пользователей высокого риска
func (c *Calculator) GetHighRiskUsers() []*RiskScore {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var highRisk []*RiskScore
	for _, score := range c.userScores {
		if score.RiskLevel == "high" || score.RiskLevel == "critical" {
			highRisk = append(highRisk, score.clone())
		}
	}

	return highRisk
}

// recalculateRisk пересчитывает оценку
func (c *Calculator) recalculateRisk(userID string) {
	history := c.eventHistory[userID]
	if len(history) == 0 {
		return
	}

	factors := c.calculateFactors(history)
	score := c.calculateScore(factors)
	level := c.determineLevel(score)
	trend := c.calculateTrend(userID, score)

	existing, exists := c.userScores[userID]
	if !exists {
		existing = &RiskScore{
			UserID:       userID,
			Factors:      make(map[string]float64),
			BehaviorData: make(map[string]interface{}),
			History:      make([]ScoreSnapshot, 0),
		}
	}

	existing.OverallScore = score
	existing.RiskLevel = level
	existing.Factors = factors
	existing.LastUpdated = time.Now()
	existing.Trend = trend

	c.userScores[userID] = existing

	c.logger.Debug("Risk recalculated",
		zap.String("user_id", userID),
		zap.Float64("score", score),
		zap.String("level", level),
		zap.String("trend", trend))
}

// calculateFactors вычисляет факторы
func (c *Calculator) calculateFactors(history []BehaviorEvent) map[string]float64 {
	factors := make(map[string]float64)

	factors["click_speed"] = c.analyzeClickSpeed(history)
	factors["form_submission"] = c.analyzeFormSubmission(history)
	factors["hover_patterns"] = c.analyzeHoverPatterns(history)
	factors["time_on_page"] = c.analyzeTimeOnPage(history)
	factors["mouse_movement"] = c.analyzeMouseMovement(history)
	factors["keyboard_patterns"] = c.analyzeKeyboardPatterns(history)
	factors["previous_clicks"] = c.analyzePreviousClicks(history)
	factors["device_fingerprint"] = c.analyzeDeviceFingerprint(history)

	return factors
}

// calculateScore вычисляет общую оценку
func (c *Calculator) calculateScore(factors map[string]float64) float64 {
	total := 0.0
	totalWeight := 0.0

	w := c.config.Weights

	if v, ok := factors["click_speed"]; ok {
		total += v * w.ClickSpeed
		totalWeight += w.ClickSpeed
	}
	if v, ok := factors["form_submission"]; ok {
		total += v * w.FormSubmission
		totalWeight += w.FormSubmission
	}
	if v, ok := factors["hover_patterns"]; ok {
		total += v * w.HoverPatterns
		totalWeight += w.HoverPatterns
	}
	if v, ok := factors["time_on_page"]; ok {
		total += v * w.TimeOnPage
		totalWeight += w.TimeOnPage
	}
	if v, ok := factors["mouse_movement"]; ok {
		total += v * w.MouseMovement
		totalWeight += w.MouseMovement
	}
	if v, ok := factors["keyboard_patterns"]; ok {
		total += v * w.KeyboardPatterns
		totalWeight += w.KeyboardPatterns
	}
	if v, ok := factors["previous_clicks"]; ok {
		total += v * w.PreviousClicks
		totalWeight += w.PreviousClicks
	}
	if v, ok := factors["device_fingerprint"]; ok {
		total += v * w.DeviceFingerprint
		totalWeight += w.DeviceFingerprint
	}

	if totalWeight == 0 {
		return 50.0
	}

	score := (total / totalWeight) * 100
	return math.Min(math.Max(score, 0), 100)
}

// determineLevel определяет уровень риска
func (c *Calculator) determineLevel(score float64) string {
	t := c.config.Thresholds

	if score >= t.Critical {
		return "critical"
	}
	if score >= t.High {
		return "high"
	}
	if score >= t.Medium {
		return "medium"
	}
	return "low"
}

// calculateTrend вычисляет тренд
func (c *Calculator) calculateTrend(userID string, currentScore float64) string {
	existing, ok := c.userScores[userID]
	if !ok || len(existing.History) < 3 {
		return "stable"
	}

	h := existing.History
	if len(h) < 3 {
		return "stable"
	}

	avg := (h[len(h)-3].Score + h[len(h)-2].Score + h[len(h)-1].Score) / 3

	diff := currentScore - avg
	if diff > 5 {
		return "worsening"
	}
	if diff < -5 {
		return "improving"
	}
	return "stable"
}

// ScoreSnapshot снимок оценки
type ScoreSnapshot struct {
	Score  float64   `json:"score"`
	Time   time.Time `json:"time"`
	Reason string    `json:"reason"`
}

// Анализаторы (заглушки - в production использовать ML)

func (c *Calculator) analyzeClickSpeed(history []BehaviorEvent) float64 {
	clicks := filterEvents(history, "click")
	if len(clicks) < 2 {
		return 50
	}

	var totalDelay float64
	for i := 1; i < len(clicks); i++ {
		delay := clicks[i].Timestamp.Sub(clicks[i-1].Timestamp).Seconds()
		totalDelay += delay
	}
	avgDelay := totalDelay / float64(len(clicks)-1)

	if avgDelay < 0.5 {
		return 80
	}
	if avgDelay < 1.0 {
		return 60
	}
	if avgDelay > 3.0 {
		return 30
	}
	return 50
}

func (c *Calculator) analyzeFormSubmission(history []BehaviorEvent) float64 {
	submits := filterEvents(history, "submit")
	if len(submits) == 0 {
		return 50
	}

	for _, s := range submits {
		if delay, ok := s.EventData["time_to_submit"].(float64); ok {
			if delay < 2.0 {
				return 80
			}
			if delay < 5.0 {
				return 60
			}
		}
	}

	return 40
}

func (c *Calculator) analyzeHoverPatterns(history []BehaviorEvent) float64 {
	hovers := filterEvents(history, "hover")
	if len(hovers) == 0 {
		return 70
	}

	if len(hovers) > 50 {
		return 60
	}

	return 40
}

func (c *Calculator) analyzeTimeOnPage(history []BehaviorEvent) float64 {
	for _, e := range history {
		if e.EventType == "time_spent" {
			if seconds, ok := e.EventData["seconds"].(float64); ok {
				if seconds < 2 {
					return 80
				}
				if seconds < 5 {
					return 60
				}
				if seconds > 300 {
					return 60
				}
			}
		}
	}
	return 40
}

func (c *Calculator) analyzeMouseMovement(history []BehaviorEvent) float64 {
	movements := filterEvents(history, "mouse_move")
	if len(movements) == 0 {
		return 70
	}

	return 40
}

func (c *Calculator) analyzeKeyboardPatterns(history []BehaviorEvent) float64 {
	keypresses := filterEvents(history, "keypress")
	inputs := filterEvents(history, "input")

	if len(inputs) > 0 && len(keypresses) == 0 {
		return 80
	}

	return 40
}

func (c *Calculator) analyzePreviousClicks(history []BehaviorEvent) float64 {
	clicks := filterEvents(history, "click")
	if len(clicks) < 5 {
		return 50
	}

	coordCount := make(map[string]int)
	for _, click := range clicks {
		if x, ok := click.EventData["x"].(float64); ok {
			if y, ok := click.EventData["y"].(float64); ok {
				key := fmt.Sprintf("%.0f,%.0f", x, y)
				coordCount[key]++
			}
		}
	}

	for _, count := range coordCount {
		if count > 5 {
			return 70
		}
	}

	return 40
}

func (c *Calculator) analyzeDeviceFingerprint(history []BehaviorEvent) float64 {
	return 50
}

func filterEvents(history []BehaviorEvent, eventType string) []BehaviorEvent {
	var result []BehaviorEvent
	for _, e := range history {
		if e.EventType == eventType {
			result = append(result, e)
		}
	}
	return result
}

func (r *RiskScore) clone() *RiskScore {
	if r == nil {
		return nil
	}

	clone := &RiskScore{
		UserID:       r.UserID,
		Email:        r.Email,
		TenantID:     r.TenantID,
		OverallScore: r.OverallScore,
		RiskLevel:    r.RiskLevel,
		Factors:      make(map[string]float64),
		BehaviorData: make(map[string]interface{}),
		LastUpdated:  r.LastUpdated,
		Trend:        r.Trend,
		History:      make([]ScoreSnapshot, len(r.History)),
	}

	for k, v := range r.Factors {
		clone.Factors[k] = v
	}

	for k, v := range r.BehaviorData {
		clone.BehaviorData[k] = v
	}

	copy(clone.History, r.History)

	return clone
}
