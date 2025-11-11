package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunDbmateUp runs dbmate up against TEST_DATABASE_URL. It returns an error if dbmate
// is not available or migrations fail. This is optional when using the dbmate_test service,
// but can be useful when running tests locally without compose.
func RunDbmateUp() error {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://stori:stori@db_test:5432/stori_test?sslmode=disable"
	}
	// Resolve repo root by walking up to find go.mod, then join db/migrations
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	migrationsDir := filepath.Join(repoRoot, "db", "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		return fmt.Errorf("migrations dir not found at %s: %w", migrationsDir, err)
	}
	cmd := exec.Command("dbmate", "--url", dsn, "--no-dump-schema", "up")
	// Ensure dbmate sees the absolute migrations path
	cmd.Env = append(os.Environ(), "DBMATE_MIGRATIONS_DIR="+migrationsDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run at repo root for clarity
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dbmate up failed: %w", err)
	}
	return nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("unable to locate repo root (go.mod not found)")
		}
		dir = parent
	}
}
