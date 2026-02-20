package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Proxy WebSocket прокси
type Proxy struct {
	mu         sync.RWMutex
	upgrader   websocket.Upgrader
	mapper     *DomainMapper
	logger     *zap.Logger
	sessions   map[string]*Session
	sessionCount int64
}

// Session представляет WebSocket сессию
type Session struct {
	ID         string
	ClientIP   string
	TargetURL  string
	StartTime  time.Time
	LastActive time.Time
	MessageCount int64
	ClientConn *websocket.Conn
	ServerConn *websocket.Conn
}

// DomainMapper для ремаппинга доменов в сообщениях
type DomainMapper struct {
	ClientDomain string
	ServerDomain string
}

// NewProxy создаёт новый WebSocket прокси
func NewProxy(logger *zap.Logger) *Proxy {
	return &Proxy{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Разрешаем все origins (для продакшена нужно ограничить)
				return true
			},
			// Разрешаем все подзаголовки
			Subprotocols: []string{"*"},
		},
		logger:   logger,
		sessions: make(map[string]*Session),
		mapper:   &DomainMapper{},
	}
}

// SetDomainMapper устанавливает маппер доменов
func (p *Proxy) SetDomainMapper(clientDomain, serverDomain string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.mapper = &DomainMapper{
		ClientDomain: clientDomain,
		ServerDomain: serverDomain,
	}
}

// HandleWS обрабатывает WebSocket подключение
func (p *Proxy) HandleWS(w http.ResponseWriter, r *http.Request) {
	clientIP := p.getClientIP(r)
	
	// Upgrade клиентского соединения до WebSocket
	clientConn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Error("WebSocket upgrade failed", 
			zap.Error(err),
			zap.String("ip", clientIP))
		return
	}
	
	// Создание сессии
	session := p.createSession(clientIP, r)
	session.ClientConn = clientConn
	
	p.logger.Info("WebSocket connection established",
		zap.String("session_id", session.ID),
		zap.String("ip", clientIP),
		zap.String("target", session.TargetURL))
	
	// Подключение к целевому серверу
	serverConn, _, err := websocket.DefaultDialer.Dial(session.TargetURL, p.buildHeaders(r))
	if err != nil {
		p.logger.Error("Failed to connect to target",
			zap.Error(err),
			zap.String("target", session.TargetURL))
		clientConn.Close()
		return
	}
	session.ServerConn = serverConn
	
	// Запуск двусторонней пересылки
	done := make(chan struct{}, 2)
	
	go func() {
		p.relayClientToServer(session)
		done <- struct{}{}
	}()
	
	go func() {
		p.relayServerToClient(session)
		done <- struct{}{}
	}()
	
	// Ожидание завершения
	<-done
	
	// Очистка
	p.cleanupSession(session)
}

// createSession создаёт новую WebSocket сессию
func (p *Proxy) createSession(clientIP string, r *http.Request) *Session {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	id := fmt.Sprintf("ws-%d", p.sessionCount)
	p.sessionCount++
	
	// Определение целевого URL
	targetURL := p.buildTargetURL(r)
	
	session := &Session{
		ID:         id,
		ClientIP:   clientIP,
		TargetURL:  targetURL,
		StartTime:  time.Now(),
		LastActive: time.Now(),
	}
	
	p.sessions[id] = session
	
	return session
}

// buildTargetURL строит целевой URL
func (p *Proxy) buildTargetURL(r *http.Request) string {
	scheme := "wss"
	if r.TLS == nil {
		scheme = "ws"
	}
	
	// Замена домена на оригинальный
	host := r.Host
	if p.mapper.ServerDomain != "" {
		host = strings.Replace(host, p.mapper.ClientDomain, p.mapper.ServerDomain, 1)
	}
	
	return fmt.Sprintf("%s://%s%s", scheme, host, r.URL.Path)
}

// buildHeaders строит заголовки для подключения к серверу
func (p *Proxy) buildHeaders(r *http.Request) http.Header {
	headers := http.Header{}
	
	// Копирование заголовков
	for key, values := range r.Header {
		for _, value := range values {
			headers.Add(key, value)
		}
	}
	
	// Замена Origin
	if origin := headers.Get("Origin"); origin != "" {
		if p.mapper.ServerDomain != "" {
			newOrigin := strings.Replace(origin, p.mapper.ClientDomain, p.mapper.ServerDomain, 1)
			headers.Set("Origin", newOrigin)
		}
	}
	
	return headers
}

// relayClientToServer пересылает сообщения от клиента к серверу
func (p *Proxy) relayClientToServer(session *Session) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("Panic in client->server relay", 
				zap.Any("recover", r),
				zap.String("session_id", session.ID))
		}
	}()
	
	for {
		msgType, message, err := session.ClientConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				p.logger.Debug("Client->Server read error",
					zap.String("session_id", session.ID),
					zap.Error(err))
			}
			return
		}
		
		// Обновление активности
		session.LastActive = time.Now()
		session.MessageCount++
		
		// Логирование
		p.logger.Debug("Client->Server",
			zap.String("session_id", session.ID),
			zap.Int("type", msgType),
			zap.Int("size", len(message)))
		
		// Ремаппинг доменов
		modifiedMessage := p.mapper.Replace(message)
		
		// Отправка серверу
		err = session.ServerConn.WriteMessage(msgType, modifiedMessage)
		if err != nil {
			p.logger.Debug("Client->Server write error",
				zap.String("session_id", session.ID),
				zap.Error(err))
			return
		}
	}
}

// relayServerToClient пересылает сообщения от сервера к клиенту
func (p *Proxy) relayServerToClient(session *Session) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("Panic in server->client relay",
				zap.Any("recover", r),
				zap.String("session_id", session.ID))
		}
	}()
	
	for {
		msgType, message, err := session.ServerConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				p.logger.Debug("Server->Client read error",
					zap.String("session_id", session.ID),
					zap.Error(err))
			}
			return
		}
		
		// Обновление активности
		session.LastActive = time.Now()
		session.MessageCount++
		
		// Логирование
		p.logger.Debug("Server->Client",
			zap.String("session_id", session.ID),
			zap.Int("type", msgType),
			zap.Int("size", len(message)))
		
		// Ремаппинг доменов
		modifiedMessage := p.mapper.Replace(message)
		
		// Отправка клиенту
		err = session.ClientConn.WriteMessage(msgType, modifiedMessage)
		if err != nil {
			p.logger.Debug("Server->Client write error",
				zap.String("session_id", session.ID),
				zap.Error(err))
			return
		}
	}
}

// cleanupSession очищает сессию
func (p *Proxy) cleanupSession(session *Session) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	delete(p.sessions, session.ID)
	
	if session.ClientConn != nil {
		session.ClientConn.Close()
	}
	if session.ServerConn != nil {
		session.ServerConn.Close()
	}
	
	p.logger.Info("WebSocket session closed",
		zap.String("session_id", session.ID),
		zap.Duration("duration", time.Since(session.StartTime)),
		zap.Int64("messages", session.MessageCount))
}

// getClientIP извлекает IP клиента
func (p *Proxy) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// Replace заменяет домены в сообщении
func (m *DomainMapper) Replace(data []byte) []byte {
	if m.ClientDomain == "" || m.ServerDomain == "" {
		return data
	}
	
	// Попытка парсинга как JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err == nil {
		m.replaceInJSON(jsonData)
		result, _ := json.Marshal(jsonData)
		return result
	}
	
	// Замена в строке
	text := string(data)
	text = strings.ReplaceAll(text, m.ServerDomain, m.ClientDomain)
	text = strings.ReplaceAll(text, m.ClientDomain, m.ServerDomain)
	return []byte(text)
}

// replaceInJSON рекурсивно заменяет домены в JSON
func (m *DomainMapper) replaceInJSON(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = strings.ReplaceAll(v, m.ServerDomain, m.ClientDomain)
		case map[string]interface{}:
			m.replaceInJSON(v)
		case []interface{}:
			for i, item := range v {
				if str, ok := item.(string); ok {
					v[i] = strings.ReplaceAll(str, m.ServerDomain, m.ClientDomain)
				} else if obj, ok := item.(map[string]interface{}); ok {
					m.replaceInJSON(obj)
				}
			}
		}
	}
}

// GetSession получает сессию по ID
func (p *Proxy) GetSession(id string) (*Session, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	session, ok := p.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	
	return session, nil
}

// ListSessions возвращает список всех сессий
func (p *Proxy) ListSessions() []*Session {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	sessions := make([]*Session, 0, len(p.sessions))
	for _, s := range p.sessions {
		sessions = append(sessions, s)
	}
	
	return sessions
}

// GetStats возвращает статистику
func (p *Proxy) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	totalMessages := int64(0)
	for _, s := range p.sessions {
		totalMessages += s.MessageCount
	}
	
	return map[string]interface{}{
		"active_sessions": len(p.sessions),
		"total_messages":  totalMessages,
		"client_domain":   p.mapper.ClientDomain,
		"server_domain":   p.mapper.ServerDomain,
	}
}

// Close закрывает все сессии
func (p *Proxy) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, session := range p.sessions {
		if session.ClientConn != nil {
			session.ClientConn.WriteMessage(websocket.CloseMessage, 
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutting down"))
			session.ClientConn.Close()
		}
		if session.ServerConn != nil {
			session.ServerConn.Close()
		}
	}
	
	p.sessions = make(map[string]*Session)
}
