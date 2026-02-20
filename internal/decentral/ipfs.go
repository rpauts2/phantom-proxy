package decentral

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// IPFSClient клиент для IPFS
type IPFSClient struct {
	mu           sync.RWMutex
	logger       *zap.Logger
	config       *IPFSConfig
	httpClient   *http.Client
	pinnedCIDs  map[string]string // name -> CID
}

// IPFSConfig конфигурация IPFS
type IPFSConfig struct {
	// Pinning service
	PinataAPIKey    string
	PinataSecretKey string
	
	// Или локальный IPFS нод
	LocalNodeURL string
	
	// Кэширование
	CacheDir string
}

// PinResponse ответ от Pinata
type PinResponse struct {
	CID             string `json:"IpfsHash"`
	PinSize         int    `json:"PinSize"`
	Timestamp       string `json:"Timestamp"`
	IsDuplicate     bool   `json:"IsDuplicate"`
}

// IPFSMetadata метаданные для Pinata
type IPFSMetadata struct {
	Name      string            `json:"name"`
	Keyvalues map[string]string `json:"keyvalues"`
}

// NewIPFSClient создаёт новый IPFS клиент
func NewIPFSClient(config *IPFSConfig, logger *zap.Logger) (*IPFSClient, error) {
	if config.CacheDir == "" {
		config.CacheDir = "./ipfs_cache"
	}
	
	// Создание директории кэша
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}
	
	return &IPFSClient{
		logger:      logger,
		config:      config,
		httpClient:  &http.Client{Timeout: 120 * time.Second},
		pinnedCIDs:  make(map[string]string),
	}, nil
}

// UploadFile загружает файл в IPFS
func (c *IPFSClient) UploadFile(ctx context.Context, filePath string, name string) (string, error) {
	c.logger.Info("Uploading file to IPFS",
		zap.String("file", filePath),
		zap.String("name", name))
	
	// Чтение файла
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	// Загрузка через Pinata
	if c.config.PinataAPIKey != "" && c.config.PinataSecretKey != "" {
		return c.uploadToPinata(ctx, fileData, name)
	}
	
	// Или локальный IPFS
	return c.uploadToLocalNode(ctx, fileData, name)
}

// UploadDirectory загружает директорию в IPFS
func (c *IPFSClient) UploadDirectory(ctx context.Context, dirPath string, name string) (string, error) {
	c.logger.Info("Uploading directory to IPFS",
		zap.String("dir", dirPath),
		zap.String("name", name))
	
	// Сбор всех файлов
	files := make(map[string][]byte)
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			
			// Относительный путь
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			
			files[relPath] = data
		}
		
		return nil
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}
	
	// Загрузка через Pinata
	if c.config.PinataAPIKey != "" && c.config.PinataSecretKey != "" {
		return c.uploadDirToPinata(ctx, files, name)
	}
	
	return "", fmt.Errorf("directory upload only supported via Pinata")
}

// uploadToPinata загружает файл в Pinata
func (c *IPFSClient) uploadToPinata(ctx context.Context, data []byte, name string) (string, error) {
	// Создание multipart формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Добавление файла
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return "", err
	}
	
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	
	// Добавление метаданных
	metadata := IPFSMetadata{
		Name: name,
		Keyvalues: map[string]string{
			"uploaded_by": "phantomproxy",
			"upload_date": time.Now().Format(time.RFC3339),
		},
	}
	
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	
	if err := writer.WriteField("pinataMetadata", string(metadataJSON)); err != nil {
		return "", err
	}
	
	writer.Close()
	
	// HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.pinata.cloud/pinning/pinFileToIPFS", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("pinata_api_key", c.config.PinataAPIKey)
	req.Header.Set("pinata_secret_api_key", c.config.PinataSecretKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Pinata API error: %d - %s", resp.StatusCode, string(respBody))
	}
	
	// Парсинг ответа
	var pinResp PinResponse
	if err := json.Unmarshal(respBody, &pinResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Сохранение в кэш
	c.mu.Lock()
	c.pinnedCIDs[name] = pinResp.CID
	c.mu.Unlock()
	
	c.logger.Info("File uploaded to Pinata",
		zap.String("name", name),
		zap.String("cid", pinResp.CID),
		zap.Int("size", pinResp.PinSize))
	
	return pinResp.CID, nil
}

// uploadDirToPinata загружает директорию в Pinata
func (c *IPFSClient) uploadDirToPinata(ctx context.Context, files map[string][]byte, name string) (string, error) {
	// Создание временной директории
	tmpDir, err := os.MkdirTemp("", "pinata-upload-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)
	
	// Копирование файлов
	for path, data := range files {
		fullPath := filepath.Join(tmpDir, path)
		
		// Создание директорий
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return "", err
		}
		
		// Запись файла
		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return "", err
		}
	}
	
	// Архивация в ZIP
	zipPath := tmpDir + ".zip"
	if err := createZip(zipPath, tmpDir); err != nil {
		return "", err
	}
	
	// Чтение ZIP
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return "", err
	}
	
	// Загрузка ZIP в Pinata
	cid, err := c.uploadToPinata(ctx, zipData, name+".zip")
	if err != nil {
		return "", err
	}
	
	return cid, nil
}

// uploadToLocalNode загружает файл в локальный IPFS нод
func (c *IPFSClient) uploadToLocalNode(ctx context.Context, data []byte, name string) (string, error) {
	if c.config.LocalNodeURL == "" {
		c.config.LocalNodeURL = "http://localhost:5001"
	}
	
	// Запрос к локальному IPFS API
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.LocalNodeURL+"/api/v0/add", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	// Парсинг ответа (упрощённо)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	cid, ok := result["Hash"].(string)
	if !ok {
		return "", fmt.Errorf("no Hash in response")
	}
	
	c.logger.Info("File uploaded to local IPFS",
		zap.String("name", name),
		zap.String("cid", cid))
	
	return cid, nil
}

// GetCID получает CID по имени
func (c *IPFSClient) GetCID(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	cid, ok := c.pinnedCIDs[name]
	if !ok {
		return "", fmt.Errorf("CID not found for name: %s", name)
	}
	
	return cid, nil
}

// GetGatewayURL возвращает URL для доступа через шлюз
func (c *IPFSClient) GetGatewayURL(cid string) string {
	return fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
}

// Unpin удаляет файл из Pinata
func (c *IPFSClient) Unpin(ctx context.Context, cid string) error {
	c.logger.Info("Unpinning from IPFS",
		zap.String("cid", cid))
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", 
		fmt.Sprintf("https://api.pinata.cloud/pinning/%s", cid), nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("pinata_api_key", c.config.PinataAPIKey)
	req.Header.Set("pinata_secret_api_key", c.config.PinataSecretKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Pinata API error: %d", resp.StatusCode)
	}
	
	c.mu.Lock()
	// Удаление из кэша
	for name, c := range c.pinnedCIDs {
		if c == cid {
			delete(c.pinnedCIDs, name)
			break
		}
	}
	c.mu.Unlock()
	
	return nil
}

// createZip создаёт ZIP архив
func createZip(zipPath string, dir string) error {
	// Упрощённая реализация
	// В продакшене использовать archive/zip
	return fmt.Errorf("not implemented")
}

// Start начинает фоновые задачи
func (c *IPFSClient) Start(ctx context.Context) error {
	c.logger.Info("Starting IPFS client")
	
	// Загрузка кэша из файла
	if err := c.loadCache(); err != nil {
		c.logger.Warn("Failed to load cache", zap.Error(err))
	}
	
	return nil
}

// loadCache загружает кэш из файла
func (c *IPFSClient) loadCache() error {
	cacheFile := filepath.Join(c.config.CacheDir, "pinned_cids.json")
	
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	
	var cids map[string]string
	if err := json.Unmarshal(data, &cids); err != nil {
		return err
	}
	
	c.mu.Lock()
	c.pinnedCIDs = cids
	c.mu.Unlock()
	
	return nil
}

// saveCache сохраняет кэш в файл
func (c *IPFSClient) saveCache() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	cacheFile := filepath.Join(c.config.CacheDir, "pinned_cids.json")
	
	data, err := json.MarshalIndent(c.pinnedCIDs, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(cacheFile, data, 0644)
}
