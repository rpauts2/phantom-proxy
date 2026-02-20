package tls

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	tls "github.com/refraction-networking/utls"
)

// SpoofManager управляет TLS fingerprint spoofing
type SpoofManager struct {
	mu        sync.RWMutex
	profiles  map[string]*Profile
	rotator   *ProfileRotator
	currentID string
}

// Profile представляет TLS профиль браузера
type Profile struct {
	ID          string
	ClientHello tls.ClientHelloID
	Priority    int
	JA3         string
	JA3S        string
	SuccessRate float64
	LastUsed    time.Time
}

// ProfileRotator автоматически переключает профили
type ProfileRotator struct {
	mu        sync.RWMutex
	profiles  []*Profile
	blacklist map[string]time.Time
	cooldown  time.Duration
}

// NewSpoofManager создаёт новый SpoofManager
func NewSpoofManager() *SpoofManager {
	sm := &SpoofManager{
		profiles: make(map[string]*Profile),
		rotator:  NewProfileRotator(),
	}
	
	// Регистрация профилей
	sm.RegisterProfile(&Profile{
		ID:          "chrome_133",
		ClientHello: tls.HelloChrome_133,
		Priority:    100,
	})
	
	sm.RegisterProfile(&Profile{
		ID:          "chrome_131",
		ClientHello: tls.HelloChrome_131,
		Priority:    95,
	})
	
	sm.RegisterProfile(&Profile{
		ID:          "chrome_120",
		ClientHello: tls.HelloChrome_120,
		Priority:    90,
	})
	
	sm.RegisterProfile(&Profile{
		ID:          "firefox_120",
		ClientHello: tls.HelloFirefox_120,
		Priority:    85,
	})
	
	sm.RegisterProfile(&Profile{
		ID:          "safari_16",
		ClientHello: tls.HelloSafari_16_0,
		Priority:    80,
	})
	
	sm.RegisterProfile(&Profile{
		ID:          "randomized",
		ClientHello: tls.HelloRandomizedALPN,
		Priority:    50,
	})
	
	return sm
}

// NewProfileRotator создаёт ротатор профилей
func NewProfileRotator() *ProfileRotator {
	return &ProfileRotator{
		profiles:  make([]*Profile, 0),
		blacklist: make(map[string]time.Time),
		cooldown:  5 * time.Minute,
	}
}

// RegisterProfile регистрирует TLS профиль
func (sm *SpoofManager) RegisterProfile(p *Profile) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.profiles[p.ID] = p
	sm.rotator.AddProfile(p)
}

// GetProfile возвращает профиль по ID
func (sm *SpoofManager) GetProfile(id string) (*Profile, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	p, ok := sm.profiles[id]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	
	return p, nil
}

// SelectProfile выбирает лучший профиль для использования
func (sm *SpoofManager) SelectProfile() *Profile {
	return sm.rotator.Select()
}

// SetCurrentProfile устанавливает текущий профиль
func (sm *SpoofManager) SetCurrentProfile(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, ok := sm.profiles[id]; !ok {
		return fmt.Errorf("profile not found: %s", id)
	}
	
	sm.currentID = id
	return nil
}

// GetCurrentProfile возвращает текущий профиль
func (sm *SpoofManager) GetCurrentProfile() *Profile {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	return sm.profiles[sm.currentID]
}

// AddProfile добавляет профиль в ротатор
func (pr *ProfileRotator) AddProfile(p *Profile) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	pr.profiles = append(pr.profiles, p)
}

// Select выбирает профиль для использования
func (pr *ProfileRotator) Select() *Profile {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	// Фильтрация заблокированных профилей
	candidates := make([]*Profile, 0)
	now := time.Now()
	
	for _, p := range pr.profiles {
		if blockedUntil, ok := pr.blacklist[p.ID]; ok {
			if now.Before(blockedUntil) {
				continue // Профиль ещё в cooldown
			}
			// Cooldown истёк, удаляем из blacklist
			delete(pr.blacklist, p.ID)
		}
		candidates = append(candidates, p)
	}
	
	if len(candidates) == 0 {
		// Все профили заблокированы, используем рандомизированный
		return pr.profiles[len(pr.profiles)-1] // randomized
	}
	
	// Выбор на основе priority + success rate
	return weightedRandom(candidates)
}

// weightedRandom выбирает профиль с учётом веса
func weightedRandom(profiles []*Profile) *Profile {
	totalWeight := 0
	for _, p := range profiles {
		totalWeight += p.Priority
	}
	
	r := rand.Intn(totalWeight)
	runningSum := 0
	
	for _, p := range profiles {
		runningSum += p.Priority
		if r < runningSum {
			return p
		}
	}
	
	return profiles[len(profiles)-1]
}

// BlockProfile временно блокирует профиль
func (pr *ProfileRotator) BlockProfile(id string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	pr.blacklist[id] = time.Now().Add(pr.cooldown)
}

// SpoofedConnection обёртка над utls.UConn
type SpoofedConnection struct {
	*tls.UConn
	Profile *Profile
}

// Dial создаёт TLS соединение со spoofed fingerprint
func (sm *SpoofManager) Dial(network, addr string) (*SpoofedConnection, error) {
	tcpConn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	
	profile := sm.SelectProfile()
	
	config := &tls.Config{
		ServerName: getServerName(addr),
		MinVersion: tls.VersionTLS12,
	}
	
	uConn := tls.UClient(tcpConn, config, profile.ClientHello)
	
	err = uConn.Handshake()
	if err != nil {
		// Блокируем профиль и пробуем другой
		sm.rotator.BlockProfile(profile.ID)
		return nil, err
	}
	
	return &SpoofedConnection{
		UConn:   uConn,
		Profile: profile,
	}, nil
}

// GetJA3Fingerprint возвращает JA3 отпечаток соединения
func (sc *SpoofedConnection) GetJA3Fingerprint() string {
	// Для получения spec нужно использовать внутренние методы utls
	// Упрощённая версия - возвращаем ID профиля
	return sc.Profile.JA3
}

// getServerName извлекает ServerName из адреса
func getServerName(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// CalculateJA3 вычисляет JA3 fingerprint из ClientHelloSpec
func CalculateJA3(spec tls.ClientHelloSpec) string {
	var cipherSuites, extensions, curves, pointFormats string
	
	// Cipher suites
	for _, cs := range spec.CipherSuites {
		if cs != tls.GREASE_PLACEHOLDER {
			cipherSuites += fmt.Sprintf("%04x,", cs)
		}
	}
	
	// Extensions
	for _, ext := range spec.Extensions {
		// Получаем тип расширения
		extType := getExtensionType(ext)
		if extType != tls.GREASE_PLACEHOLDER {
			extensions += fmt.Sprintf("%d,", extType)
		}
	}
	
	// Curves (Supported Groups)
	for _, ext := range spec.Extensions {
		if curveExt, ok := ext.(*tls.SupportedCurvesExtension); ok {
			for _, curve := range curveExt.Curves {
				if curve != tls.GREASE_PLACEHOLDER {
					curves += fmt.Sprintf("%d,", curve)
				}
			}
			break
		}
	}
	
	// Point formats
	for _, ext := range spec.Extensions {
		if pointExt, ok := ext.(*tls.SupportedPointsExtension); ok {
			for _, point := range pointExt.SupportedPoints {
				pointFormats += fmt.Sprintf("%d,", point)
			}
			break
		}
	}
	
	// Формирование JA3 строки
	ja3String := fmt.Sprintf("771,%s,%s,%s,%s", 
		cipherSuites, extensions, curves, pointFormats)
	
	// SHA256 хеш
	hash := sha256.Sum256([]byte(ja3String))
	return hex.EncodeToString(hash[:])
}

// getExtensionType возвращает тип TLS расширения
func getExtensionType(ext tls.TLSExtension) uint16 {
	// Используем рефлексию или type assertion для получения типа
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

// UpdateProfileSuccessRate обновляет статистику успешности профиля
func (sm *SpoofManager) UpdateProfileSuccessRate(id string, success bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if p, ok := sm.profiles[id]; ok {
		if success {
			p.SuccessRate = (p.SuccessRate*float64(p.Priority) + 1.0) / float64(p.Priority+1)
		} else {
			p.SuccessRate = (p.SuccessRate * float64(p.Priority)) / float64(p.Priority+1)
		}
		p.LastUsed = time.Now()
	}
}

// GetStats возвращает статистику по профилям
func (sm *SpoofManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	profiles := make([]map[string]interface{}, 0)
	
	for id, p := range sm.profiles {
		profiles = append(profiles, map[string]interface{}{
			"id":           id,
			"priority":     p.Priority,
			"success_rate": p.SuccessRate,
			"last_used":    p.LastUsed.Format(time.RFC3339),
		})
	}
	
	stats["profiles"] = profiles
	stats["current_profile"] = sm.currentID
	
	return stats
}
