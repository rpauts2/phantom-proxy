package tenant

import (
	"context"
	"testing"

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

// Helper function to create test database
func createTestDB(t *testing.T) *testDB {
	// Simplified test DB helper
	return &testDB{}
}

type testDB struct {
	closed bool
}

func (db *testDB) Close() error {
	db.closed = true
	return nil
}

func (db *testDB) ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

func (db *testDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*testRows, error) {
	return &testRows{}, nil
}

func (db *testDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *testRow {
	return &testRow{}
}

type testRows struct{}
func (r *testRows) Close() error                     { return nil }
func (r *testRows) Next() bool                       { return false }
func (r *testRows) Scan(dest ...interface{}) error   { return nil }
func (r *testRows) Columns() ([]string, error)       { return nil, nil }

type testRow struct{}
func (r *testRow) Scan(dest ...interface{}) error { return nil }
