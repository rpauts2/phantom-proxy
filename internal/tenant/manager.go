// Package tenant - Multi-tenant Manager
package tenant

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tenant представляет клиента
type Tenant struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Enabled     bool      `json:"enabled"`
	Plan        string    `json:"plan"` // free, pro, enterprise
	MaxSessions int       `json:"max_sessions"`
	MaxUsers    int       `json:"max_users"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Metadata    string    `json:"metadata"`
}

// User представляет пользователя
type User struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // admin, operator, viewer
	Enabled   bool      `json:"enabled"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
}

// Manager управляет tenant'ами
type Manager struct {
	mu      sync.RWMutex
	db      *sql.DB
	logger  *zap.Logger
	tenants map[string]*Tenant
	users   map[string]*User
}

// NewManager создает менеджер tenant'ов
func NewManager(db *sql.DB, logger *zap.Logger) *Manager {
	return &Manager{
		db:      db,
		logger:  logger,
		tenants: make(map[string]*Tenant),
		users:   make(map[string]*User),
	}
}

// Init инициализирует схему БД
func (m *Manager) Init(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tenants (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		enabled BOOLEAN DEFAULT true,
		plan TEXT DEFAULT 'free',
		max_sessions INTEGER DEFAULT 100,
		max_users INTEGER DEFAULT 10,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		metadata TEXT
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL,
		email TEXT NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'viewer',
		enabled BOOLEAN DEFAULT true,
		last_login DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id),
		UNIQUE(tenant_id, email)
	);

	CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
	`

	_, err := m.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("schema init failed: %w", err)
	}

	m.logger.Info("Multi-tenant schema initialized")
	return nil
}

// CreateTenant создает клиента
func (m *Manager) CreateTenant(ctx context.Context, name, slug, plan string) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	tenant := &Tenant{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Enabled:     true,
		Plan:        plan,
		MaxSessions: getPlanLimit(plan, "sessions"),
		MaxUsers:    getPlanLimit(plan, "users"),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(1, 0, 0),
	}

	_, err := m.db.ExecContext(ctx,
		`INSERT INTO tenants (id, name, slug, enabled, plan, max_sessions, max_users, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Enabled, tenant.Plan,
		tenant.MaxSessions, tenant.MaxUsers, tenant.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	m.tenants[tenant.Slug] = tenant
	m.logger.Info("Tenant created", zap.String("slug", slug), zap.String("plan", plan))

	return tenant, nil
}

// GetTenantBySlug получает tenant по slug
func (m *Manager) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	m.mu.RLock()
	if tenant, ok := m.tenants[slug]; ok {
		m.mu.RUnlock()
		return tenant, nil
	}
	m.mu.RUnlock()

	tenant := &Tenant{}
	err := m.db.QueryRowContext(ctx,
		`SELECT id, name, slug, enabled, plan, max_sessions, max_users, created_at, expires_at, metadata
		 FROM tenants WHERE slug = ? AND enabled = true`, slug).
		Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Enabled, &tenant.Plan,
			&tenant.MaxSessions, &tenant.MaxUsers, &tenant.CreatedAt, &tenant.ExpiresAt, &tenant.Metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", slug)
	}
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.tenants[slug] = tenant
	m.mu.Unlock()

	return tenant, nil
}

// CreateUser создает пользователя
func (m *Manager) CreateUser(ctx context.Context, tenantID, email, password, role string) (*User, error) {
	id := uuid.New().String()
	user := &User{
		ID:        id,
		TenantID:  tenantID,
		Email:     email,
		Password:  password, // TODO: Hash password
		Role:      role,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	_, err := m.db.ExecContext(ctx,
		`INSERT INTO users (id, tenant_id, email, password, role, enabled)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, user.TenantID, user.Email, user.Password, user.Role, user.Enabled)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	m.mu.Lock()
	m.users[email] = user
	m.mu.Unlock()

	m.logger.Info("User created", zap.String("email", email), zap.String("role", role))
	return user, nil
}

// GetUserByEmail получает пользователя по email
func (m *Manager) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	if user, ok := m.users[email]; ok {
		m.mu.RUnlock()
		return user, nil
	}
	m.mu.RUnlock()

	user := &User{}
	err := m.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, email, password, role, enabled, last_login, created_at
		 FROM users WHERE email = ? AND enabled = true`, email).
		Scan(&user.ID, &user.TenantID, &user.Email, &user.Password, &user.Role,
			&user.Enabled, &user.LastLogin, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", email)
	}
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.users[email] = user
	m.mu.Unlock()

	return user, nil
}

// CheckQuota проверяет квоты
func (m *Manager) CheckQuota(ctx context.Context, tenantID, resource string) (bool, error) {
	tenant := &Tenant{}
	err := m.db.QueryRowContext(ctx,
		`SELECT id, max_sessions, max_users FROM tenants WHERE id = ?`, tenantID).
		Scan(&tenant.ID, &tenant.MaxSessions, &tenant.MaxUsers)

	if err != nil {
		return false, err
	}

	var current int
	var max int

	switch resource {
	case "sessions":
		err = m.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM sessions WHERE tenant_id = ?`, tenantID).Scan(&current)
		max = tenant.MaxSessions
	case "users":
		err = m.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM users WHERE tenant_id = ?`, tenantID).Scan(&current)
		max = tenant.MaxUsers
	default:
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return current < max, nil
}

// ListTenants возвращает список tenant'ов
func (m *Manager) ListTenants(ctx context.Context) ([]*Tenant, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, name, slug, enabled, plan, max_sessions, max_users, created_at, expires_at, metadata
		 FROM tenants ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		t := &Tenant{}
		err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Enabled, &t.Plan,
			&t.MaxSessions, &t.MaxUsers, &t.CreatedAt, &t.ExpiresAt, &t.Metadata)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}

	return tenants, nil
}

// GetTenantStats возвращает статистику tenant
func (m *Manager) GetTenantStats(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var sessions, creds, users, phishlets int

	m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE tenant_id = ?`, tenantID).Scan(&sessions)
	m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM credentials WHERE tenant_id = ?`, tenantID).Scan(&creds)
	m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE tenant_id = ?`, tenantID).Scan(&users)
	m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM phishlets WHERE tenant_id = ?`, tenantID).Scan(&phishlets)

	stats["sessions"] = sessions
	stats["credentials"] = creds
	stats["users"] = users
	stats["phishlets"] = phishlets

	return stats, nil
}

// DeleteTenant удаляет tenant
func (m *Manager) DeleteTenant(ctx context.Context, slug string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.db.ExecContext(ctx, `DELETE FROM tenants WHERE slug = ?`, slug)
	if err != nil {
		return err
	}

	delete(m.tenants, slug)
	m.logger.Info("Tenant deleted", zap.String("slug", slug))
	return nil
}

func getPlanLimit(plan, resource string) int {
	limits := map[string]map[string]int{
		"free":       {"sessions": 100, "users": 5},
		"pro":        {"sessions": 1000, "users": 25},
		"enterprise": {"sessions": 10000, "users": 100},
	}

	if p, ok := limits[plan]; ok {
		if v, ok := p[resource]; ok {
			return v
		}
	}
	return 100
}
