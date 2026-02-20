package exfiltration

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExfilConfig simulation configuration
type ExfilConfig struct {
	Enabled       bool
	TargetTypes   []string // file extensions: .docx, .xlsx, .pdf
	MaxSizeMB     int
	Compression   bool
	Encryption    bool
	CloudProvider string   // gdrive, dropbox, onedrive, s3
	ScopePath     string   // allowed paths for simulation
}

// ExfilResult result of simulated exfiltration
type ExfilResult struct {
	FilesCount   int
	TotalBytes   int64
	Duration     time.Duration
	Compressed   bool
	Encrypted    bool
	Destination  string
	MetadataOnly bool // если true — только метаданные, без реальной отправки
}

// Simulator simulates data exfiltration for DLP testing
type Simulator struct {
	config *ExfilConfig
}

// NewSimulator creates exfiltration simulator
func NewSimulator(cfg *ExfilConfig) *Simulator {
	return &Simulator{config: cfg}
}

// Simulate runs exfiltration simulation
func (s *Simulator) Simulate(ctx context.Context, basePath string) (*ExfilResult, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	start := time.Now()
	var totalBytes int64
	var filesCount int

	allowed := make(map[string]bool)
	for _, ext := range s.config.TargetTypes {
		allowed[strings.ToLower(ext)] = true
	}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !allowed[ext] && len(allowed) > 0 {
			return nil
		}
		if s.config.MaxSizeMB > 0 && info.Size() > int64(s.config.MaxSizeMB)*1024*1024 {
			return nil
		}
		// Metadata-only: не читаем содержимое, только считаем
		totalBytes += info.Size()
		filesCount++
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ExfilResult{
		FilesCount:   filesCount,
		TotalBytes:   totalBytes,
		Duration:     time.Since(start),
		Compressed:   s.config.Compression,
		Encrypted:    s.config.Encryption,
		Destination:  s.config.CloudProvider,
		MetadataOnly: true,
	}, nil
}

// GenerateFakePayload creates fake data for DLP trigger testing
func (s *Simulator) GenerateFakePayload(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}

// CloudClient interface for cloud storage (Google Drive, Dropbox, OneDrive)
type CloudClient interface {
	Upload(ctx context.Context, path string, data []byte) (string, error)
}

// MockCloudClient for testing
type MockCloudClient struct{}

func (m *MockCloudClient) Upload(ctx context.Context, path string, data []byte) (string, error) {
	_ = ctx
	_ = path
	_ = data
	return "mock://uploaded/" + fmt.Sprintf("%d", time.Now().Unix()), nil
}
