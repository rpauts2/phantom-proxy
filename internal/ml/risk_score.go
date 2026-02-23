// Package ml - Human Risk Scoring Module
// Real-time behavioral analysis for human risk assessment
package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ============================================================================
// Risk Score Types
// ============================================================================

// RiskCategory represents the FSTEC classification
type RiskCategory string

const (
	RiskCategoryLow    RiskCategory = "low"     // УЗ-3 (минимальный)
	RiskCategoryMedium RiskCategory = "medium"   // УЗ-2 (средний)
	RiskCategoryHigh   RiskCategory = "high"     // УЗ-1 (высокий)
	RiskCategoryCritical RiskCategory = "critical" // Критический
)

// ============================================================================
// Risk Score Engine
// ============================================================================

type RiskEngine struct {
	mu           sync.RWMutex
	logger       *zap.Logger
	redis        *redis.Client
	enabled      bool
	models       map[string]RiskModel
	thresholds   RiskThresholds
	window       time.Duration
}

// RiskModel represents a scoring model
type RiskModel struct {
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Weights     map[string]float64 `json:"weights"`
	Rules       []RiskRule         `json:"rules"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// RiskRule represents a scoring rule
type RiskRule struct {
	Name       string      `json:"name"`
	Condition  string      `json:"condition"`  // "click_speed < 2.0"
	Weight     float64     `json:"weight"`     // 0.3
	Multiplier float64     `json:"multiplier"` // 2.0
}

// RiskThresholds defines scoring thresholds
type RiskThresholds struct {
	LowThreshold    float32 `json:"low_threshold"`
	MediumThreshold float32 `json:"medium_threshold"`
	HighThreshold   float32 `json:"high_threshold"`
	CriticalLimit  float32 `json:"critical_limit"`
	MaxScore       float32 `json:"max_score"`
}

// SessionMetrics represents behavioral metrics for a session
type SessionMetrics struct {
	SessionID        string    `json:"session_id"`
	TenantID         string    `json:"tenant_id"`
	UserID          string    `json:"user_id"`
	
	// Timing metrics
	TimeOnPage       float64   `json:"time_on_page"`       // seconds
	ClickSpeed       float64   `json:"click_speed"`        // clicks per second
	MouseMovements   int       `json:"mouse_movements"`
	ScrollDepth      float64   `json:"scroll_depth"`       // 0-100%
	KeystrokeSpeed   float64   `json:"keystroke_speed"`    // chars per second
	
	// Pattern metrics
	CopyPasteCount   int       `json:"copy_paste_count"`
	TabSwitches     int       `json:"tab_switches"`
	BackButtonUsed  bool      `json:"back_button_used"`
	
	// Historical
	PreviousScore    float32   `json:"previous_score"`
	HistoricalBreaches int     `json:"historical_breaches"`
	TrainingCount   int       `json:"training_count"`
	
	// Device
	DeviceType       string    `json:"device_type"` // desktop, mobile, tablet
	BrowserType      string    `json:"browser_type"`
	OS              string    `json:"os"`
	
	// Timestamp
	Timestamp       time.Time `json:"timestamp"`
}

// RiskScore represents the calculated risk score
type RiskScore struct {
	SessionID    string       `json:"session_id"`
	TenantID     string       `json:"tenant_id"`
	UserID       string       `json:"user_id"`
	
	Score        float32      `json:"score"`         // 0.0 - 100.0
	Category     RiskCategory `json:"category"`      // low, medium, high, critical
	Confidence   float32      `json:"confidence"`    // 0.0 - 1.0
	
	// Breakdown
	Factors       []RiskFactor `json:"factors"`
	
	// FSTEC
	FSTECCategory string       `json:"fstec_category"` // УЗ-1, УЗ-2, УЗ-3
	
	// Metadata
	ModelVersion  string      `json:"model_version"`
	CalculatedAt  time.Time   `json:"calculated_at"`
	ExpiresAt     time.Time   `json:"expires_at"`
}

// RiskFactor represents an individual risk factor
type RiskFactor struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Weight      float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
	Description string  `json:"description"`
}

// ============================================================================
// New Risk Engine
// ============================================================================

func NewRiskEngine(logger *zap.Logger, redisClient *redis.Client) *RiskEngine {
	return &RiskEngine{
		logger:    logger,
		redis:     redisClient,
		enabled:   true,
		models:    make(map[string]RiskModel),
		window:    24 * time.Hour,
		thresholds: RiskThresholds{
			LowThreshold:    20.0,
			MediumThreshold: 50.0,
			HighThreshold:   75.0,
			CriticalLimit:   90.0,
			MaxScore:        100.0,
		},
	}
}

// ============================================================================
// Calculate Risk Score
// ============================================================================

func (e *RiskEngine) CalculateScore(ctx context.Context, metrics *SessionMetrics) (*RiskScore, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.enabled {
		return &RiskScore{
			SessionID:   metrics.SessionID,
			TenantID:    metrics.TenantID,
			Score:       0,
			Category:    RiskCategoryLow,
			Confidence:  0,
		}, nil
	}

	// Get or create model
	model := e.getOrCreateModel(metrics.TenantID)

	// Calculate individual factors
	factors := e.calculateFactors(metrics, &model)

	// Calculate total score
	score := e.calculateTotalScore(factors)

	// Determine category
	category := e.determineCategory(score)

	// Calculate confidence
	confidence := e.calculateConfidence(metrics, factors)

	// Map to FSTEC category
	fstecCategory := e.mapToFSTEC(score)

	result := &RiskScore{
		SessionID:    metrics.SessionID,
		TenantID:     metrics.TenantID,
		UserID:       metrics.UserID,
		Score:        score,
		Category:     category,
		Confidence:  confidence,
		Factors:     factors,
		FSTECCategory: fstecCategory,
		ModelVersion: model.Version,
		CalculatedAt: time.Now(),
		ExpiresAt:    time.Now().Add(e.window),
	}

	// Store in Redis
	e.storeScore(ctx, result)

	// Emit event for real-time updates
	e.logger.Info("Risk score calculated",
		zap.String("session_id", metrics.SessionID),
		zap.Float32("score", score),
		zap.String("category", string(category)))

	return result, nil
}

// ============================================================================
// Factor Calculation
// ============================================================================

func (e *RiskEngine) calculateFactors(metrics *SessionMetrics, model *RiskModel) []RiskFactor {
	factors := make([]RiskFactor, 0)

	// 1. Time on page factor
	timeFactor := e.calculateTimeFactor(metrics)
	factors = append(factors, timeFactor)

	// 2. Click speed factor
	clickFactor := e.calculateClickSpeedFactor(metrics)
	factors = append(factors, clickFactor)

	// 3. Mouse movement factor
	mouseFactor := e.calculateMouseFactor(metrics)
	factors = append(factors, mouseFactor)

	// 4. Copy/paste factor
	copyPasteFactor := e.calculateCopyPasteFactor(metrics)
	factors = append(factors, copyPasteFactor)

	// 5. Scroll depth factor
	scrollFactor := e.calculateScrollFactor(metrics)
	factors = append(factors, scrollFactor)

	// 6. Tab switching factor
	tabFactor := e.calculateTabSwitchFactor(metrics)
	factors = append(factors, tabFactor)

	// 7. Historical factor
	historyFactor := e.calculateHistoricalFactor(metrics)
	factors = append(factors, historyFactor)

	// 8. Device factor
	deviceFactor := e.calculateDeviceFactor(metrics)
	factors = append(factors, deviceFactor)

	// Apply model weights
	for i := range factors {
		if weight, ok := model.Weights[factors[i].Name]; ok {
			factors[i].Weight = weight
			factors[i].Contribution = factors[i].Value * weight
		}
	}

	return factors
}

func (e *RiskEngine) calculateTimeFactor(m *SessionMetrics) RiskFactor {
	// Very fast time on page (< 1 second) is suspicious
	value := 0.0
	description := "Normal time on page"

	if m.TimeOnPage < 1.0 {
		value = 1.0
		description = "Suspiciously fast time on page (< 1s)"
	} else if m.TimeOnPage < 3.0 {
		value = 0.7
		description = "Very fast time on page (< 3s)"
	} else if m.TimeOnPage < 5.0 {
		value = 0.4
		description = "Fast time on page (< 5s)"
	} else if m.TimeOnPage > 300 {
		value = 0.3 // Longer time might mean hesitation
		description = "Long time on page (potential hesitation)"
	}

	return RiskFactor{
		Name:         "time_on_page",
		Value:        value,
		Weight:       0.15,
		Contribution: value * 0.15,
		Description:  description,
	}
}

func (e *RiskEngine) calculateClickSpeedFactor(m *SessionMetrics) RiskFactor {
	// Very fast clicking is suspicious
	value := 0.0
	description := "Normal click speed"

	if m.ClickSpeed > 10.0 {
		value = 1.0
		description = "Automated click pattern detected"
	} else if m.ClickSpeed > 5.0 {
		value = 0.7
		description = "Very fast clicking"
	} else if m.ClickSpeed > 3.0 {
		value = 0.4
		description = "Fast clicking"
	}

	return RiskFactor{
		Name:         "click_speed",
		Value:        value,
		Weight:       0.15,
		Contribution: value * 0.15,
		Description:  description,
	}
}

func (e *RiskEngine) calculateMouseFactor(m *SessionMetrics) RiskFactor {
	// Lack of mouse movements is suspicious
	value := 0.0
	description := "Normal mouse behavior"

	if m.MouseMovements == 0 && m.DeviceType == "desktop" {
		value = 0.8
		description = "No mouse movements (possible automation)"
	} else if m.MouseMovements < 5 {
		value = 0.5
		description = "Very few mouse movements"
	} else if m.MouseMovements < 20 {
		value = 0.3
		description = "Few mouse movements"
	}

	return RiskFactor{
		Name:         "mouse_movements",
		Value:        value,
		Weight:       0.10,
		Contribution: value * 0.10,
		Description:  description,
	}
}

func (e *RiskEngine) calculateCopyPasteFactor(m *SessionMetrics) RiskFactor {
	// Copying credentials is suspicious (could be password manager)
	value := 0.0
	description := "Normal input behavior"

	if m.CopyPasteCount > 3 {
		value = 0.6
		description = "Multiple copy/paste actions"
	} else if m.CopyPasteCount > 1 {
		value = 0.4
		description = "Some copy/paste actions"
	}

	return RiskFactor{
		Name:         "copy_paste",
		Value:        value,
		Weight:       0.10,
		Contribution: value * 0.10,
		Description:  description,
	}
}

func (e *RiskEngine) calculateScrollFactor(m *SessionMetrics) RiskFactor {
	// Not scrolling is suspicious
	value := 0.0
	description := "Normal scroll behavior"

	if m.ScrollDepth == 0 {
		value = 0.5
		description = "No scrolling (didn't review content)"
	} else if m.ScrollDepth < 25 {
		value = 0.3
		description = "Minimal scrolling"
	}

	return RiskFactor{
		Name:         "scroll_depth",
		Value:        value,
		Weight:       0.05,
		Contribution: value * 0.05,
		Description:  description,
	}
}

func (e *RiskEngine) calculateTabSwitchFactor(m *SessionMetrics) RiskFactor {
	// Tab switching could indicate multi-tasking or automation
	value := 0.0
	description := "Normal tab usage"

	if m.TabSwitches > 10 {
		value = 0.7
		description = "Excessive tab switching"
	} else if m.TabSwitches > 5 {
		value = 0.4
		description = "Multiple tab switches"
	}

	return RiskFactor{
		Name:         "tab_switches",
		Value:        value,
		Weight:       0.05,
		Contribution: value * 0.05,
		Description:  description,
	}
}

func (e *RiskEngine) calculateHistoricalFactor(m *SessionMetrics) RiskFactor {
	// Previous breaches increase risk
	value := 0.0
	description := "No historical risk factors"

	if m.HistoricalBreaches > 3 {
		value = 1.0
		description = "Multiple previous breaches"
	} else if m.HistoricalBreaches > 1 {
		value = 0.7
		description = "Previous breaches detected"
	} else if m.HistoricalBreaches == 1 {
		value = 0.5
		description = "One previous breach"
	}

	// Training helps reduce risk
	if m.TrainingCount > 10 {
		value -= 0.3
		description += ", high training count"
	}

	value = math.Max(0, math.Min(1, value))

	return RiskFactor{
		Name:         "historical",
		Value:        value,
		Weight:       0.20,
		Contribution: value * 0.20,
		Description:  description,
	}
}

func (e *RiskEngine) calculateDeviceFactor(m *SessionMetrics) RiskFactor {
	// Some devices are riskier
	value := 0.0
	description := "Standard device"

	switch m.DeviceType {
	case "mobile":
		value = 0.2
		description = "Mobile device (limited visibility)"
	case "tablet":
		value = 0.1
		description = "Tablet device"
	}

	// Unknown OS is riskier
	if m.OS == "" || m.OS == "unknown" {
		value += 0.3
		description += ", unknown OS"
	}

	return RiskFactor{
		Name:         "device",
		Value:        value,
		Weight:       0.10,
		Contribution: value * 0.10,
		Description:  description,
	}
}

// ============================================================================
// Score Calculation
// ============================================================================

func (e *RiskEngine) calculateTotalScore(factors []RiskFactor) float32 {
	var total float64

	for _, factor := range factors {
		total += factor.Contribution
	}

	// Normalize to 0-100
	score := float32(total * 100)
	if score > e.thresholds.MaxScore {
		score = e.thresholds.MaxScore
	}

	return score
}

func (e *RiskEngine) determineCategory(score float32) RiskCategory {
	switch {
	case score >= e.thresholds.CriticalLimit:
		return RiskCategoryCritical
	case score >= e.thresholds.HighThreshold:
		return RiskCategoryHigh
	case score >= e.thresholds.MediumThreshold:
		return RiskCategoryMedium
	default:
		return RiskCategoryLow
	}
}

func (e *RiskEngine) calculateConfidence(metrics *SessionMetrics, factors []RiskFactor) float32 {
	// Confidence is based on data completeness and consistency
	confidence := 0.5 // Base

	// More data = higher confidence
	if metrics.MouseMovements > 0 {
		confidence += 0.1
	}
	if metrics.TimeOnPage > 0 {
		confidence += 0.1
	}
	if metrics.DeviceType != "" {
		confidence += 0.1
	}
	if metrics.BrowserType != "" {
		confidence += 0.1
	}
	if metrics.HistoricalBreaches > 0 || metrics.TrainingCount > 0 {
		confidence += 0.1
	}

	return float32(math.Min(1.0, confidence))
}

func (e *RiskEngine) mapToFSTEC(score float32) string {
	switch {
	case score >= e.thresholds.CriticalLimit:
		return "УЗ-1" // Высокий риск
	case score >= e.thresholds.HighThreshold:
		return "УЗ-1" // Высокий риск
	case score >= e.thresholds.MediumThreshold:
		return "УЗ-2" // Средний риск
	default:
		return "УЗ-3" // Минимальный риск
	}
}

// ============================================================================
// Model Management
// ============================================================================

func (e *RiskEngine) getOrCreateModel(tenantID string) RiskModel {
	if model, ok := e.models[tenantID]; ok {
		return model
	}

	// Create default model
	model := RiskModel{
		Name:    "default",
		Version: "1.0.0",
		Weights: map[string]float64{
			"time_on_page":  0.15,
			"click_speed":   0.15,
			"mouse_movements": 0.10,
			"copy_paste":    0.10,
			"scroll_depth":  0.05,
			"tab_switches":  0.05,
			"historical":    0.20,
			"device":        0.10,
		},
		Rules:       []RiskRule{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	e.models[tenantID] = model
	return model
}

// ============================================================================
// Storage
// ============================================================================

func (e *RiskEngine) storeScore(ctx context.Context, score *RiskScore) error {
	key := fmt.Sprintf("risk:%s:%s", score.TenantID, score.SessionID)
	data, err := json.Marshal(score)
	if err != nil {
		return err
	}

	return e.redis.Set(ctx, key, data, e.window).Err()
}

func (e *RiskEngine) GetScore(ctx context.Context, tenantID, sessionID string) (*RiskScore, error) {
	key := fmt.Sprintf("risk:%s:%s", tenantID, sessionID)
	data, err := e.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var score RiskScore
	err = json.Unmarshal(data, &score)
	return &score, err
}

// ============================================================================
// Stats
// ============================================================================

func (e *RiskEngine) GetStats(ctx context.Context) (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	keys, err := e.redis.Keys(ctx, "risk:*").Result()
	if err != nil {
		return nil, err
	}

	// Count by category
	categories := map[string]int{
		"low":      0,
		"medium":   0,
		"high":     0,
		"critical": 0,
	}

	for _, key := range keys {
		data, err := e.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var score RiskScore
		if err := json.Unmarshal(data, &score); err == nil {
			categories[string(score.Category)]++
		}
	}

	return map[string]interface{}{
		"enabled":      e.enabled,
		"total_scores": len(keys),
		"by_category":  categories,
		"thresholds":   e.thresholds,
	}, nil
}

// ============================================================================
// Configuration
// ============================================================================

func (e *RiskEngine) Enable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = true
	e.logger.Info("Risk engine enabled")
}

func (e *RiskEngine) Disable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = false
	e.logger.Info("Risk engine disabled")
}

func (e *RiskEngine) SetThresholds(thresholds RiskThresholds) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.thresholds = thresholds
	e.logger.Info("Risk thresholds updated", zap.Any("thresholds", thresholds))
}
