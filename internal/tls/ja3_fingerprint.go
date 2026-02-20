package tls

import (
	"crypto/sha256"
	"encoding/hex"
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

// Fingerprint создаёт JA3 отпечаток из ClientHello
func (j *JA3Fingerprinter) Fingerprint(conn *tls.UConn) string {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	j.totalFingerprints++
	
	spec := conn.GetClientHelloSpec()
	
	// Формирование JA3 строки
	var cipherSuites, extensions, curves, pointFormats string
	
	// Cipher suites
	for _, cs := range spec.CipherSuites {
		if cs != tls.GREASE_PLACEHOLDER {
			cipherSuites += formatUint16(cs) + ","
		}
	}
	
	// Extensions
	for _, ext := range spec.Extensions {
		extType := getExtensionType(ext)
		if extType != tls.GREASE_PLACEHOLDER {
			extensions += formatUint16(extType) + ","
		}
	}
	
	// Curves (Supported Groups)
	for _, ext := range spec.Extensions {
		if curveExt, ok := ext.(*tls.SupportedCurvesExtension); ok {
			for _, curve := range curveExt.Curves {
				if curve != tls.GREASE_PLACEHOLDER {
					curves += formatUint16(uint16(curve)) + ","
				}
			}
			break
		}
	}
	
	// Point formats
	for _, ext := range spec.Extensions {
		if pointExt, ok := ext.(*tls.SupportedPointsExtension); ok {
			for _, point := range pointExt.SupportedPoints {
				pointFormats += formatUint8(point) + ","
			}
			break
		}
	}
	
	// Формирование полной JA3 строки
	ja3String := "771," + cipherSuites + "," + extensions + "," + curves + "," + pointFormats
	
	// SHA256 хеш
	hash := sha256.Sum256([]byte(ja3String))
	ja3Hash := hex.EncodeToString(hash[:])
	
	j.logger.Debug("JA3 fingerprint created",
		zap.String("ja3", ja3Hash))
	
	return ja3Hash
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

// Вспомогательные функции

func formatUint16(v uint16) string {
	return fmt.Sprintf("%d", v)
}

func formatUint8(v uint8) string {
	return fmt.Sprintf("%d", v)
}

func getExtensionType(ext tls.TLSExtension) uint16 {
	switch ext.(type) {
	case *tls.SNIExtension:
		return 0
	case *tls.ExtendedMasterSecretExtension:
		return 23
	case *tls.RenegotiationInfoExtension:
		return 65281
	case *tls.SupportedCurvesExtension:
		return 10
	case *tls.SupportedPointsExtension:
		return 11
	case *tls.SessionTicketExtension:
		return 35
	case *tls.ALPNExtension:
		return 16
	case *tls.StatusRequestExtension:
		return 5
	case *tls.SignatureAlgorithmsExtension:
		return 13
	case *tls.SCTExtension:
		return 18
	case *tls.KeyShareExtension:
		return 51
	case *tls.PSKKeyExchangeModesExtension:
		return 45
	case *tls.SupportedVersionsExtension:
		return 43
	case *tls.UtlsCompressCertExtension:
		return 27
	case *tls.UtlsPaddingExtension:
		return 21
	case *tls.UtlsGREASEExtension:
		return tls.GREASE_PLACEHOLDER
	default:
		return 0
	}
}
