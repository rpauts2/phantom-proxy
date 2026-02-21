package database

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &DatabaseConfig{
		Type:       DatabaseSQLite,
		SQLitePath: ":memory:",
	}

	db, err := NewDatabase(config, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	if db == nil {
		t.Fatal("Expected database to be created")
	}
	db.Close()
}

func TestDatabaseConnection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	db, err := NewDatabase(&DatabaseConfig{
		Type:       DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test database is usable
	err = db.db.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestGetStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	db, _ := NewDatabase(&DatabaseConfig{
		Type:       DatabaseSQLite,
		SQLitePath: ":memory:",
	}, logger)
	defer db.Close()

	// Get stats (should work even with empty database)
	stats, err := db.GetStats()
	if err != nil {
		t.Errorf("GetStats returned error: %v", err)
	}
	if stats == nil {
		t.Error("Expected stats to be returned")
	}
}
