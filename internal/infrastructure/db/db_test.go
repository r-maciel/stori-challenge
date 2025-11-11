package db

import (
	"testing"
)

func TestOpen_WithoutEnv_ReturnsError(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	db, err := Open()
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatalf("expected error when DATABASE_URL is empty")
	}
}

func TestOpen_WithEnv_SetsMaxOpenConns(t *testing.T) {
	// Use a parseable DSN; sql.Open does not dial until first use.
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/dbname?sslmode=disable")

	db, err := Open()
	if err != nil {
		t.Fatalf("unexpected error from Open: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatalf("expected non-nil *sql.DB")
	}
	// Verify defaults are untouched (database/sql default is 0 = unlimited)
	if got := db.Stats().MaxOpenConnections; got != 0 {
		t.Fatalf("expected default MaxOpenConnections=0, got %d", got)
	}
}
