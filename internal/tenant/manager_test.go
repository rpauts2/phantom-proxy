package tenant

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	manager := NewManager(db, logger)
	if manager == nil {
		t.Fatal("Expected manager to be created")
	}
}

func TestCreateTenant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)

	// Init schema
	if err := manager.Init(ctx); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Create tenant
	tenant, err := manager.CreateTenant(ctx, "Test Corp", "test-corp", "pro")
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	if tenant.ID == "" {
		t.Error("Expected tenant ID to be set")
	}
	if tenant.Name != "Test Corp" {
		t.Errorf("Expected name 'Test Corp', got %s", tenant.Name)
	}
	if tenant.Slug != "test-corp" {
		t.Errorf("Expected slug 'test-corp', got %s", tenant.Slug)
	}
	if tenant.Plan != "pro" {
		t.Errorf("Expected plan 'pro', got %s", tenant.Plan)
	}
}

func TestGetTenantBySlug(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)
	manager.Init(ctx)

	// Create tenant
	_, _ = manager.CreateTenant(ctx, "Test Corp", "test-corp", "pro")

	// Get tenant
	tenant, err := manager.GetTenantBySlug(ctx, "test-corp")
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if tenant.Name != "Test Corp" {
		t.Errorf("Expected name 'Test Corp', got %s", tenant.Name)
	}
}

func TestCreateUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)
	manager.Init(ctx)

	// Create tenant
	tenant, _ := manager.CreateTenant(ctx, "Test Corp", "test-corp", "pro")

	// Create user
	user, err := manager.CreateUser(ctx, tenant.ID, "admin@test.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == "" {
		t.Error("Expected user ID to be set")
	}
	if user.Email != "admin@test.com" {
		t.Errorf("Expected email 'admin@test.com', got %s", user.Email)
	}
	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got %s", user.Role)
	}
}

func TestGetUserByEmail(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)
	manager.Init(ctx)

	// Create tenant and user
	tenant, _ := manager.CreateTenant(ctx, "Test Corp", "test-corp", "pro")
	_, _ = manager.CreateUser(ctx, tenant.ID, "admin@test.com", "password123", "admin")

	// Get user
	user, err := manager.GetUserByEmail(ctx, "admin@test.com")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.Email != "admin@test.com" {
		t.Errorf("Expected email 'admin@test.com', got %s", user.Email)
	}
}

func TestCheckQuota(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)
	manager.Init(ctx)

	// Create tenant
	tenant, _ := manager.CreateTenant(ctx, "Test Corp", "test-corp", "free")

	// Check quota (should pass - under limit)
	ok, err := manager.CheckQuota(ctx, tenant.ID, "sessions")
	if err != nil {
		t.Fatalf("Failed to check quota: %v", err)
	}
	if !ok {
		t.Error("Expected quota check to pass")
	}
}

func TestListTenants(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)
	manager.Init(ctx)

	// Create multiple tenants
	_, _ = manager.CreateTenant(ctx, "Corp 1", "corp1", "free")
	_, _ = manager.CreateTenant(ctx, "Corp 2", "corp2", "pro")

	// List tenants
	tenants, err := manager.ListTenants(ctx)
	if err != nil {
		t.Fatalf("Failed to list tenants: %v", err)
	}

	if len(tenants) != 2 {
		t.Errorf("Expected 2 tenants, got %d", len(tenants))
	}
}

func TestGetTenantStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	ctx := context.Background()
	manager := NewManager(db, logger)
	manager.Init(ctx)

	// Create tenant
	tenant, _ := manager.CreateTenant(ctx, "Test Corp", "test-corp", "pro")

	// Get stats
	stats, err := manager.GetTenantStats(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats == nil {
		t.Error("Expected stats to be returned")
	}
}

// Helper function to create test database stub
func createTestDB(t *testing.T) DB {
	return newTestDB()
}

// testDB is a minimal in-memory database emulation used by unit tests.
// It tracks tenants and users so that manager methods can observe state.
type testDB struct {
	tenants map[string]*Tenant
	users   map[string]*User
}

func newTestDB() *testDB {
	return &testDB{
		tenants: make(map[string]*Tenant),
		users:   make(map[string]*User),
	}
}

func (db *testDB) Close() error {
	return nil
}

func (db *testDB) ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	q := strings.ToUpper(strings.TrimSpace(query))
	if strings.HasPrefix(q, "INSERT INTO TENANTS") {
		id := args[0].(string)
		name := args[1].(string)
		slug := args[2].(string)
		plan := args[4].(string)
		// derive limits similarly to manager.getPlanLimit
		maxSessions := 100
		maxUsers := 10
		switch plan {
		case "pro":
			maxSessions = 1000
			maxUsers = 25
		case "enterprise":
			maxSessions = 10000
			maxUsers = 100
		}
		db.tenants[slug] = &Tenant{ID: id, Name: name, Slug: slug, Plan: plan, MaxSessions: maxSessions, MaxUsers: maxUsers}
	}
	if strings.HasPrefix(q, "INSERT INTO USERS") {
		id := args[0].(string)
		tid := args[1].(string)
		email := args[2].(string)
		role := args[4].(string)
		db.users[email] = &User{ID: id, TenantID: tid, Email: email, Role: role}
	}
	return nil, nil
}

// QueryContext returns a row iterator for LISTTENANTS or other multi-row queries
func (db *testDB) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	q := strings.ToUpper(query)
	if strings.Contains(q, "FROM TENANTS") {
		// build rows from stored tenants
		data := make([][]interface{}, 0, len(db.tenants))
		for _, t := range db.tenants {
			data = append(data, []interface{}{t.ID, t.Name, t.Slug, t.Enabled, t.Plan, t.MaxSessions, t.MaxUsers, t.CreatedAt, t.ExpiresAt, t.Metadata})
		}
		return &testRows{data: data}, nil
	}
	// default to empty
	return &testRows{}, nil
}

// QueryRowContext handles single-row selects used by various manager methods
func (db *testDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	q := strings.ToUpper(query)
	if strings.Contains(q, "FROM TENANTS WHERE SLUG") {
		slug := args[0].(string)
		if t, ok := db.tenants[slug]; ok {
			return &genericRow{vals: []interface{}{t.ID, t.Name, t.Slug, t.Enabled, t.Plan, t.MaxSessions, t.MaxUsers, t.CreatedAt, t.ExpiresAt, t.Metadata}}
		}
		return &genericRow{err: sql.ErrNoRows}
	}
	if strings.Contains(q, "FROM USERS WHERE EMAIL") {
		email := args[0].(string)
		if u, ok := db.users[email]; ok {
			return &genericRow{vals: []interface{}{u.ID, u.TenantID, u.Email, u.Password, u.Role, u.Enabled, u.LastLogin, u.CreatedAt}}
		}
		return &genericRow{err: sql.ErrNoRows}
	}
	if strings.Contains(q, "SELECT ID, MAX_SESSIONS") {
		// used by CheckQuota first query
		tid := args[0].(string)
		for _, t := range db.tenants {
			if t.ID == tid {
				return &genericRow{vals: []interface{}{t.ID, t.MaxSessions, t.MaxUsers}}
			}
		}
		return &genericRow{err: sql.ErrNoRows}
	}
	// count queries: return zero
	if strings.Contains(q, "COUNT(*)") {
		return &genericRow{vals: []interface{}{0}}
	}
	return &genericRow{err: sql.ErrNoRows}
}

// Rows implementation storing row data

type testRows struct {
	idx  int
	data [][]interface{}
}

func (r *testRows) Close() error                   { return nil }
func (r *testRows) Next() bool {
	if r.idx < len(r.data) {
		r.idx++
		return true
	}
	return false
}
func (r *testRows) Scan(dest ...interface{}) error {
	if r.idx == 0 || r.idx > len(r.data) {
		return sql.ErrNoRows
	}
	row := r.data[r.idx-1]
	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			*d = row[i].(string)
		case *bool:
			*d = row[i].(bool)
		case *int:
			*d = row[i].(int)
		case *time.Time:
			if v, ok := row[i].(time.Time); ok {
				*d = v
			}
		default:
			// ignore other types
		}
	}
	return nil
}
func (r *testRows) Columns() ([]string, error)     { return nil, nil }

// Row implementation with arbitrary values or error

type genericRow struct {
	vals []interface{}
	err error
}

func (r *genericRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			*d = r.vals[i].(string)
		case *bool:
			*d = r.vals[i].(bool)
		case *int:
			*d = r.vals[i].(int)
		case *time.Time:
			if v, ok := r.vals[i].(time.Time); ok {
				*d = v
			}
		default:
			// ignore
		}
	}
	return nil
}
