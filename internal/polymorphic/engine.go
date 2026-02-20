package polymorphic

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Engine полиморфный JS движок
type Engine struct {
	mu            sync.RWMutex
	enabled       bool
	mutationLevel string // low, medium, high
	seedRotation  int    // минут
	currentSeed   int64
	lastRotation  time.Time
	
	// Статистика
	mutationCount int64
}

// MutationResult результат мутации
type MutationResult struct {
	OriginalHash string
	MutatedCode  string
	Mutations    []string
	Seed         int64
}

// NewEngine создаёт новый движок
func NewEngine(mutationLevel string, seedRotation int) *Engine {
	if mutationLevel == "" {
		mutationLevel = "high"
	}
	if seedRotation == 0 {
		seedRotation = 15
	}
	
	return &Engine{
		enabled:       true,
		mutationLevel: mutationLevel,
		seedRotation:  seedRotation,
		currentSeed:   time.Now().UnixNano(),
		lastRotation:  time.Now(),
	}
}

// Mutate применяет мутации к JavaScript коду
func (e *Engine) Mutate(code string) *MutationResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.enabled {
		return &MutationResult{
			OriginalHash: e.hash(code),
			MutatedCode:  code,
			Mutations:    []string{"none"},
		}
	}
	
	e.rotateSeed()
	
	result := &MutationResult{
		OriginalHash: e.hash(code),
		MutatedCode:  code,
		Mutations:    make([]string, 0),
		Seed:         e.currentSeed,
	}
	
	// Уровень 1: Переименование переменных (medium, high)
	if e.mutationLevel != "low" {
		result.MutatedCode = e.renameVariables(result.MutatedCode)
		result.Mutations = append(result.Mutations, "variable_renaming")
	}
	
	// Уровень 2: Трансформация строк (все уровни)
	result.MutatedCode = e.transformStrings(result.MutatedCode)
	result.Mutations = append(result.Mutations, "string_transformation")
	
	// Уровень 3: Base64 мутация (все уровни)
	result.MutatedCode = e.mutateBase64(result.MutatedCode)
	result.Mutations = append(result.Mutations, "base64_mutation")
	
	// Уровень 4: Мёртвый код (high)
	if e.mutationLevel == "high" {
		result.MutatedCode = e.addDeadCode(result.MutatedCode)
		result.Mutations = append(result.Mutations, "dead_code_injection")
	}
	
	// Уровень 5: Изменение порядка операций (high)
	if e.mutationLevel == "high" {
		result.MutatedCode = e.reorderOperations(result.MutatedCode)
		result.Mutations = append(result.Mutations, "operation_reordering")
	}
	
	e.mutationCount++
	
	return result
}

// renameVariables переименовывает переменные
func (e *Engine) renameVariables(code string) string {
	varRegex := regexp.MustCompile(`(var|let|const)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=`)
	
	varMap := make(map[string]string)
	
	return varRegex.ReplaceAllStringFunc(code, func(match string) string {
		parts := varRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		
		oldName := parts[2]
		
		if _, exists := varMap[oldName]; !exists {
			varMap[oldName] = e.generateRandomName()
		}
		newName := varMap[oldName]
		
		result := strings.Replace(match, oldName, newName, 1)
		result = strings.ReplaceAll(result, oldName, newName)
		
		return result
	})
}

// generateRandomName генерирует случайное имя
func (e *Engine) generateRandomName() string {
	prefixes := []string{"_0x", "_var", "_tmp", "_ctx", "_obj"}
	chars := "abcdef0123456789"
	
	prefix := prefixes[e.randomInt(0, len(prefixes))]
	name := prefix
	
	for i := 0; i < 8; i++ {
		name += string(chars[e.randomInt(0, len(chars))])
	}
	
	return name
}

// transformStrings трансформирует строковые литералы
func (e *Engine) transformStrings(code string) string {
	stringRegex := regexp.MustCompile(`"([^"\\]*(\\.[^"\\]*)*)"`)
	
	return stringRegex.ReplaceAllStringFunc(code, func(s string) string {
		if e.randomBool() {
			content := s[1 : len(s)-1]
			return e.toFromCharCode(content)
		}
		return s
	})
}

// toFromCharCode преобразует строку в fromCharCode
func (e *Engine) toFromCharCode(s string) string {
	var codes []string
	for _, c := range s {
		codes = append(codes, fmt.Sprintf("%d", c))
	}
	return fmt.Sprintf("String.fromCharCode(%s)", strings.Join(codes, ","))
}

// mutateBase64 изменяет способы генерации base64
func (e *Engine) mutateBase64(code string) string {
	btoaRegex := regexp.MustCompile(`btoa\(([^)]+)\)`)
	
	alternatives := []string{
		`Buffer.from($1).toString('base64')`,
		`window.btoa.call(null, $1)`,
		`Function("return btoa")()($1)`,
		`atob(String.fromCharCode.apply(null, $1.split('').map(c => c.charCodeAt(0))))`,
		`btoa.call(window, $1)`,
	}
	
	return btoaRegex.ReplaceAllString(code, e.randomChoice(alternatives))
}

// addDeadCode добавляет бесполезный код
func (e *Engine) addDeadCode(code string) string {
	deadCodes := []string{
		`void 0;`,
		`!function(){};`,
		`Math.random()>2&&0;`,
		`for(let i=0;i<0;i++);`,
		`undefined&&undefined;`,
		`!function(){return!1}();`,
		`0&&(function(){})(),`,
		`NaN<2&&0,`,
	}
	
	pos := e.randomInt(0, len(code))
	insertion := e.randomChoice(deadCodes)
	
	return code[:pos] + insertion + code[pos:]
}

// reorderOperations изменяет порядок операций
func (e *Engine) reorderOperations(code string) string {
	propRegex := regexp.MustCompile(`\.([a-zA-Z_$][a-zA-Z0-9_$]*)`)
	
	return propRegex.ReplaceAllStringFunc(code, func(match string) string {
		if e.randomBool() {
			prop := match[1:]
			return fmt.Sprintf(`["%s"]`, prop)
		}
		return match
	})
}

// rotateSeed обновляет seed
func (e *Engine) rotateSeed() {
	now := time.Now()
	if now.Sub(e.lastRotation).Minutes() >= float64(e.seedRotation) {
		e.currentSeed = now.UnixNano()
		e.lastRotation = now
	}
}

// hash вычисляет хеш кода
func (e *Engine) hash(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

// randomBool возвращает случайный boolean
func (e *Engine) randomBool() bool {
	n, _ := rand.Int(rand.Reader, big.NewInt(2))
	return n.Int64() == 1
}

// randomInt возвращает случайное число в диапазоне [min, max)
func (e *Engine) randomInt(min, max int) int {
	if min >= max {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return min + int(n.Int64())
}

// randomChoice выбирает случайный элемент
func (e *Engine) randomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	idx := e.randomInt(0, len(choices))
	return choices[idx]
}

// GetStats возвращает статистику
func (e *Engine) GetStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return map[string]interface{}{
		"enabled":        e.enabled,
		"mutation_level": e.mutationLevel,
		"seed_rotation":  e.seedRotation,
		"mutation_count": e.mutationCount,
		"current_seed":   e.currentSeed,
	}
}

// Enable включает движок
func (e *Engine) Enable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = true
}

// Disable выключает движок
func (e *Engine) Disable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = false
}

// SetMutationLevel устанавливает уровень мутации
func (e *Engine) SetMutationLevel(level string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mutationLevel = level
}

// SetSeedRotation устанавливает интервал ротации seed
func (e *Engine) SetSeedRotation(minutes int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.seedRotation = minutes
}
