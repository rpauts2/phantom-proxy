package campaign

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Tracker система отслеживания
type Tracker struct {
	mu sync.RWMutex
	// In-memory storage for demo - in production use database
	opens   map[string][]OpenEvent
	clicks  map[string][]ClickEvent
	submits map[string][]SubmitEvent
}

// OpenEvent событие открытия письма
type OpenEvent struct {
	ID         string
	CampaignID string
	Email      string
	Timestamp time.Time
	IP        string
	UserAgent string
}

// ClickEvent событие клика
type ClickEvent struct {
	ID         string
	CampaignID string
	Email      string
	URL        string
	Timestamp  time.Time
	IP         string
}

// SubmitEvent событие отправки формы
type SubmitEvent struct {
	ID         string
	CampaignID string
	Email      string
	Data       map[string]string
	Timestamp  time.Time
	IP         string
}

// NewTracker создает новый трекер
func NewTracker() *Tracker {
	return &Tracker{
		opens:   make(map[string][]OpenEvent),
		clicks:  make(map[string][]ClickEvent),
		submits: make(map[string][]SubmitEvent),
	}
}

// TrackOpen записывает открытие письма
func (t *Tracker) TrackOpen(campaignID, email, ip, userAgent string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	event := OpenEvent{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Email:      email,
		Timestamp:  time.Now(),
		IP:         ip,
		UserAgent:  userAgent,
	}

	t.opens[campaignID] = append(t.opens[campaignID], event)
	return nil
}

// TrackClick записывает клик
func (t *Tracker) TrackClick(campaignID, email, url, ip string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	event := ClickEvent{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Email:      email,
		URL:        url,
		Timestamp:  time.Now(),
		IP:         ip,
	}

	t.clicks[campaignID] = append(t.clicks[campaignID], event)
	return nil
}

// TrackSubmit записывает отправку формы
func (t *Tracker) TrackSubmit(campaignID, email string, data map[string]string, ip string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	event := SubmitEvent{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Email:      email,
		Data:       data,
		Timestamp:  time.Now(),
		IP:         ip,
	}

	t.submits[campaignID] = append(t.submits[campaignID], event)
	return nil
}

// GetOpens получает все открытия
func (t *Tracker) GetOpens(campaignID string) []OpenEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.opens[campaignID]
}

// GetClicks получает все клики
func (t *Tracker) GetClicks(campaignID string) []ClickEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.clicks[campaignID]
}

// GetSubmits получает все отправки форм
func (t *Tracker) GetSubmits(campaignID string) []SubmitEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.submits[campaignID]
}

// GenerateOpenPixel генерирует HTML для трекинг-пикселя
func (t *Tracker) GenerateOpenPixel(campaignID, email string) template.HTML {
	trackingURL := fmt.Sprintf("/track/open?c=%s&e=%s", campaignID, email)
	pixel := fmt.Sprintf(`<img src="%s" width="1" height="1" style="display:none" />`, trackingURL)
	return template.HTML(pixel)
}

// GenerateClickLink генерирует трекинг-ссылку
func (t *Tracker) GenerateClickLink(campaignID, email, originalURL string) string {
	return fmt.Sprintf("/track/click?c=%s&e=%s&u=%s", campaignID, email, originalURL)
}

// InjectTrackingPixel внедряет трекинг-пиксель в HTML
func (t *Tracker) InjectTrackingPixel(htmlContent, campaignID, email string) string {
	pixel := string(t.GenerateOpenPixel(campaignID, email))
	return htmlContent + pixel
}

// InjectAllTrackingLinks заменяет все ссылки на трекинг-ссылки
func (t *Tracker) InjectAllTrackingLinks(htmlContent, campaignID, email string) string {
	// Простая замена href - в реальном проекте использовать regex
	// Это заглушка - нужно доработать
	return htmlContent
}

// GetStats получает статистику кампании
func (t *Tracker) GetStats(campaignID string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	opens := t.opens[campaignID]
	clicks := t.clicks[campaignID]
	submits := t.submits[campaignID]

	// Уникальные открытия
	uniqueOpens := make(map[string]bool)
	for _, o := range opens {
		uniqueOpens[o.Email] = true
	}

	// Уникальные клики
	uniqueClicks := make(map[string]bool)
	for _, c := range clicks {
		uniqueClicks[c.Email] = true
	}

	// Уникальные отправки
	uniqueSubmits := make(map[string]bool)
	for _, s := range submits {
		uniqueSubmits[s.Email] = true
	}

	return map[string]interface{}{
		"total_opens":       len(opens),
		"unique_opens":      len(uniqueOpens),
		"total_clicks":      len(clicks),
		"unique_clicks":     len(uniqueClicks),
		"total_submits":     len(submits),
		"unique_submits":    len(uniqueSubmits),
		"open_rate":         0.0,
		"click_rate":        0.0,
		"submission_rate":   0.0,
	}
}

// CalculateRates вычисляет метрики
func (t *Tracker) CalculateRates(campaignID string, totalSent int) map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	opens := t.opens[campaignID]
	clicks := t.clicks[campaignID]
	submits := t.submits[campaignID]

	uniqueOpens := len(opens)
	uniqueClicks := len(clicks)
	uniqueSubmits := len(submits)

	if totalSent == 0 {
		return map[string]float64{
			"open_rate":       0,
			"click_rate":      0,
			"submission_rate": 0,
		}
	}

	return map[string]float64{
		"open_rate":       float64(uniqueOpens) / float64(totalSent) * 100,
		"click_rate":      float64(uniqueClicks) / float64(totalSent) * 100,
		"submission_rate": float64(uniqueSubmits) / float64(totalSent) * 100,
	}
}

// Flush очищает данные кампании
func (t *Tracker) Flush(campaignID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.opens, campaignID)
	delete(t.clicks, campaignID)
	delete(t.submits, campaignID)
}

// Global tracker instance
var globalTracker = NewTracker()

// GetGlobalTracker возвращает глобальный трекер
func GetGlobalTracker() *Tracker {
	return globalTracker
}
