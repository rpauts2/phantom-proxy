package tls

import (
	"sync"

	tls "github.com/refraction-networking/utls"
	"go.uber.org/zap"
)

// JA3Fingerprinter анализатор JA3 отпечатков
type JA3Fingerprinter struct {
	mu            sync.RWMutex
	enabled       bool
	blockKnownBots bool
	logger        *zap.Logger
	
	// Базы данных отпечатков
	knownBots     map[string]bool  // JA3 хеш -> true если бот
	customSigs    map[string]bool  // Пользовательские сигнатуры
	
	// Статистика
	totalFingerprints int64
	blockedCount      int64
}

// NewJA3Fingerprinter создаёт анализатор
func NewJA3Fingerprinter(logger *zap.Logger) *JA3Fingerprinter {
	return &JA3Fingerprinter{
		enabled:        true,
		blockKnownBots: true,
		logger:         logger,
		knownBots:      make(map[string]bool),
		customSigs:     make(map[string]bool),
	}
}

// Fingerprint создаёт JA3 отпечаток из ClientHello.
// Из-за ограничений текущей версии utls подробное вычисление JA3 отключено.
// Возвращаем пустую строку, сохраняя счётчики и логику блокировок.
func (j *JA3Fingerprinter) Fingerprint(conn *tls.UConn) string {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.totalFingerprints++
	j.logger.Debug("JA3 fingerprinting disabled for current utls version")
	return ""
}

// IsBlocked проверяет заблокирован ли отпечаток
func (j *JA3Fingerprinter) IsBlocked(ja3Hash string) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	if !j.enabled {
		return false
	}
	
	if j.blockKnownBots && j.knownBots[ja3Hash] {
		j.blockedCount++
		j.logger.Warn("Known bot JA3 blocked",
			zap.String("ja3", ja3Hash))
		return true
	}
	
	if j.customSigs[ja3Hash] {
		j.blockedCount++
		j.logger.Warn("Custom signature blocked",
			zap.String("ja3", ja3Hash))
		return true
	}
	
	return false
}

// AddKnownBot добавляет известный бот отпечаток
func (j *JA3Fingerprinter) AddKnownBot(ja3Hash string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.knownBots[ja3Hash] = true
	j.logger.Info("Known bot JA3 added",
		zap.String("ja3", ja3Hash))
}

// AddCustomSignature добавляет пользовательскую сигнатуру
func (j *JA3Fingerprinter) AddCustomSignature(ja3Hash string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.customSigs[ja3Hash] = true
	j.logger.Info("Custom signature added",
		zap.String("ja3", ja3Hash))
}

// GetStats возвращает статистику
func (j *JA3Fingerprinter) GetStats() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	return map[string]interface{}{
		"enabled":            j.enabled,
		"block_known_bots":   j.blockKnownBots,
		"total_fingerprints": j.totalFingerprints,
		"blocked_count":      j.blockedCount,
		"known_bots_count":   len(j.knownBots),
		"custom_sigs_count":  len(j.customSigs),
	}
}

// Enable включает fingerprinting
func (j *JA3Fingerprinter) Enable() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.enabled = true
	j.logger.Info("JA3 fingerprinting enabled")
}

// Disable выключает fingerprinting
func (j *JA3Fingerprinter) Disable() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.enabled = false
	j.logger.Info("JA3 fingerprinting disabled")
}

// SetBlockKnownBots включает/выключает блокировку известных ботов
func (j *JA3Fingerprinter) SetBlockKnownBots(block bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.blockKnownBots = block
	j.logger.Info("Block known bots updated",
		zap.Bool("block", block))
}

// Вспомогательные функции больше не используются и удалены для совместимости.
