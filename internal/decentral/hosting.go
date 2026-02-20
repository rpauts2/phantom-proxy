package decentral

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DecentralizedHosting модуль децентрализованного хостинга
type DecentralizedHosting struct {
	mu        sync.RWMutex
	logger    *zap.Logger
	config    *HostingConfig
	ipfs      *IPFSClient
	ens       *ENSClient
	pages     map[string]*HostingPage
}

// HostingConfig конфигурация хостинга
type HostingConfig struct {
	// IPFS
	PinataAPIKey    string
	PinataSecretKey string
	LocalIPFSNode   string
	
	// ENS
	EthereumRPC      string
	EthereumPrivateKey string
	ENSRegistryAddress string
	
	// Кэширование
	CacheDir string
	
	// Автообновление
	AutoUpdateInterval time.Duration
}

// HostingPage страница для хостинга
type HostingPage struct {
	Name        string    `json:"name"`
	SourcePath  string    `json:"source_path"`
	IPFSCID     string    `json:"ipfs_cid"`
	ENSName     string    `json:"ens_name"`
	GatewayURL  string    `json:"gateway_url"`
	ENSURL      string    `json:"ens_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AutoUpdate  bool      `json:"auto_update"`
}

// NewDecentralizedHosting создаёт новый модуль хостинга
func NewDecentralizedHosting(config *HostingConfig, logger *zap.Logger) (*DecentralizedHosting, error) {
	if config.CacheDir == "" {
		config.CacheDir = "./decentral_cache"
	}
	
	// Создание директории
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, err
	}
	
	// Создание IPFS клиента
	ipfs, err := NewIPFSClient(&IPFSConfig{
		PinataAPIKey:    config.PinataAPIKey,
		PinataSecretKey: config.PinataSecretKey,
		LocalNodeURL:    config.LocalIPFSNode,
		CacheDir:        config.CacheDir,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPFS client: %w", err)
	}
	
	// Создание ENS клиента (опционально)
	var ens *ENSClient
	if config.EthereumPrivateKey != "" {
		ens, err = NewENSClient(&ENSConfig{
			RPCURL:             config.EthereumRPC,
			PrivateKey:         config.EthereumPrivateKey,
			ENSRegistryAddress: config.ENSRegistryAddress,
		}, logger)
		if err != nil {
			logger.Warn("Failed to create ENS client", zap.Error(err))
			// Продолжаем без ENS
		}
	}
	
	h := &DecentralizedHosting{
		logger:  logger,
		config:  config,
		ipfs:    ipfs,
		ens:     ens,
		pages:   make(map[string]*HostingPage),
	}
	
	// Загрузка существующих страниц
	if err := h.loadPages(); err != nil {
		logger.Warn("Failed to load existing pages", zap.Error(err))
	}
	
	return h, nil
}

// HostPage публикует страницу в децентрализованной сети
func (h *DecentralizedHosting) HostPage(ctx context.Context, name string, sourcePath string, ensName string) (*HostingPage, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.logger.Info("Hosting page",
		zap.String("name", name),
		zap.String("source", sourcePath),
		zap.String("ens", ensName))
	
	// Проверка пути
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source path does not exist: %s", sourcePath)
	}
	
	// Загрузка в IPFS
	var cid string
	var err error
	
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, err
	}
	
	if fileInfo.IsDir() {
		cid, err = h.ipfs.UploadDirectory(ctx, sourcePath, name)
	} else {
		cid, err = h.ipfs.UploadFile(ctx, sourcePath, name)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	
	// Обновление ENS (если есть)
	if h.ens != nil && ensName != "" {
		contentHash := fmt.Sprintf("ipfs://%s", cid)
		if err := h.ens.RegisterENS(ctx, ensName, contentHash); err != nil {
			h.logger.Warn("Failed to register ENS", zap.Error(err))
		}
	}
	
	// Создание страницы
	page := &HostingPage{
		Name:        name,
		SourcePath:  sourcePath,
		IPFSCID:     cid,
		ENSName:     ensName,
		GatewayURL:  h.ipfs.GetGatewayURL(cid),
		ENSURL:      fmt.Sprintf("https://%s.limo", ensName), // eth.limo шлюз
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AutoUpdate:  false,
	}
	
	h.pages[name] = page
	
	// Сохранение
	if err := h.savePages(); err != nil {
		h.logger.Warn("Failed to save pages", zap.Error(err))
	}
	
	h.logger.Info("Page hosted successfully",
		zap.String("name", name),
		zap.String("cid", cid),
		zap.String("gateway", page.GatewayURL))
	
	return page, nil
}

// UpdatePage обновляет страницу
func (h *DecentralizedHosting) UpdatePage(ctx context.Context, name string) (*HostingPage, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	page, ok := h.pages[name]
	if !ok {
		return nil, fmt.Errorf("page not found: %s", name)
	}
	
	h.logger.Info("Updating page",
		zap.String("name", name))
	
	// Повторная загрузка в IPFS
	var cid string
	var err error
	
	fileInfo, err := os.Stat(page.SourcePath)
	if err != nil {
		return nil, err
	}
	
	if fileInfo.IsDir() {
		cid, err = h.ipfs.UploadDirectory(ctx, page.SourcePath, name)
	} else {
		cid, err = h.ipfs.UploadFile(ctx, page.SourcePath, name)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Обновление CID
	oldCID := page.IPFSCID
	page.IPFSCID = cid
	page.GatewayURL = h.ipfs.GetGatewayURL(cid)
	page.UpdatedAt = time.Now()
	
	// Обновление ENS
	if h.ens != nil && page.ENSName != "" {
		contentHash := fmt.Sprintf("ipfs://%s", cid)
		if err := h.ens.UpdateENS(ctx, page.ENSName, contentHash); err != nil {
			h.logger.Warn("Failed to update ENS", zap.Error(err))
		}
	}
	
	// Unpin старой версии
	if oldCID != "" {
		if err := h.ipfs.Unpin(ctx, oldCID); err != nil {
			h.logger.Warn("Failed to unpin old CID", zap.Error(err))
		}
	}
	
	h.logger.Info("Page updated",
		zap.String("name", name),
		zap.String("old_cid", oldCID),
		zap.String("new_cid", cid))
	
	return page, nil
}

// GetPage получает информацию о странице
func (h *DecentralizedHosting) GetPage(name string) (*HostingPage, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	page, ok := h.pages[name]
	if !ok {
		return nil, fmt.Errorf("page not found: %s", name)
	}
	
	return page, nil
}

// ListPages возвращает список всех страниц
func (h *DecentralizedHosting) ListPages() []*HostingPage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	pages := make([]*HostingPage, 0, len(h.pages))
	for _, page := range h.pages {
		pages = append(pages, page)
	}
	
	return pages
}

// DeletePage удаляет страницу
func (h *DecentralizedHosting) DeletePage(ctx context.Context, name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	page, ok := h.pages[name]
	if !ok {
		return fmt.Errorf("page not found: %s", name)
	}
	
	h.logger.Info("Deleting page",
		zap.String("name", name))
	
	// Unpin из IPFS
	if page.IPFSCID != "" {
		if err := h.ipfs.Unpin(ctx, page.IPFSCID); err != nil {
			h.logger.Warn("Failed to unpin", zap.Error(err))
		}
	}
	
	// Удаление из ENS
	if h.ens != nil && page.ENSName != "" {
		// TODO: Удаление ENS записи
	}
	
	delete(h.pages, name)
	
	// Сохранение
	if err := h.savePages(); err != nil {
		h.logger.Warn("Failed to save pages", zap.Error(err))
	}
	
	return nil
}

// EnableAutoUpdate включает автообновление
func (h *DecentralizedHosting) EnableAutoUpdate(name string, interval time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	page, ok := h.pages[name]
	if !ok {
		return fmt.Errorf("page not found: %s", name)
	}
	
	page.AutoUpdate = true
	
	h.logger.Info("Auto-update enabled",
		zap.String("name", name),
		zap.Duration("interval", interval))
	
	return nil
}

// Start начинает фоновые задачи
func (h *DecentralizedHosting) Start(ctx context.Context) error {
	h.logger.Info("Starting decentralized hosting")
	
	// Запуск IPFS
	if err := h.ipfs.Start(ctx); err != nil {
		return err
	}
	
	// Запуск ENS
	if h.ens != nil {
		if err := h.ens.Start(ctx); err != nil {
			h.logger.Warn("Failed to start ENS", zap.Error(err))
		}
	}
	
	// Фоновая задача автообновления
	go h.autoUpdateWorker(ctx)
	
	return nil
}

// autoUpdateWorker фоновая задача автообновления
func (h *DecentralizedHosting) autoUpdateWorker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkAutoUpdates()
		}
	}
}

// checkAutoUpdates проверяет страницы с автообновлением
func (h *DecentralizedHosting) checkAutoUpdates() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, page := range h.pages {
		if page.AutoUpdate {
			// Проверка изменений в исходном пути
			// TODO: Сравнение хешей файлов
		}
	}
}

// loadPages загружает страницы из файла
func (h *DecentralizedHosting) loadPages() error {
	// TODO: Загрузка из JSON файла
	return nil
}

// savePages сохраняет страницы в файл
func (h *DecentralizedHosting) savePages() error {
	// TODO: Сохранение в JSON файл
	cacheFile := filepath.Join(h.config.CacheDir, "pages.json")
	
	// Упрощённая реализация
	_ = cacheFile
	
	return nil
}
