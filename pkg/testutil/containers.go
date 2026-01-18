// Package testutil provides testing utilities including database containers
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a testcontainer postgres instance
type PostgresContainer struct {
	Container testcontainers.Container
	DB        *sql.DB
	ConnStr   string
}

// NewPostgresContainer creates a new PostgreSQL container for testing
func NewPostgresContainer(t *testing.T) *PostgresContainer {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("vtp_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	pc := &PostgresContainer{
		Container: container,
		DB:        db,
		ConnStr:   connStr,
	}

	// Run migrations
	if err := pc.RunMigrations(t); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Cleanup on test completion
	t.Cleanup(func() {
		db.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Warning: failed to terminate container: %v", err)
		}
	})

	return pc
}

// RunMigrations applies all SQL migrations from the migrations directory
func (pc *PostgresContainer) RunMigrations(t *testing.T) error {
	t.Helper()

	// Find migrations directory (relative to project root)
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		return fmt.Errorf("migrations directory not found")
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply each migration
	for _, filename := range migrationFiles {
		path := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		_, err = pc.DB.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}

		t.Logf("Applied migration: %s", filename)
	}

	return nil
}

// findMigrationsDir searches for the migrations directory
func findMigrationsDir() string {
	// Try common relative paths
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"../../../migrations",
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}

// SeedTestData inserts common test data
func (pc *PostgresContainer) SeedTestData(t *testing.T) {
	t.Helper()

	// Create a test user
	_, err := pc.DB.Exec(`
		INSERT INTO users (id, email, password_hash, role, full_name, created_at, updated_at)
		VALUES 
			('11111111-1111-1111-1111-111111111111', 'teacher@test.com', '$2a$10$test', 'teacher', 'Test Teacher', NOW(), NOW()),
			('22222222-2222-2222-2222-222222222222', 'student@test.com', '$2a$10$test', 'student', 'Test Student', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		t.Logf("Warning: failed to seed test users: %v", err)
	}
}

// TruncateTables removes all data from specified tables
func (pc *PostgresContainer) TruncateTables(t *testing.T, tables ...string) {
	t.Helper()

	for _, table := range tables {
		_, err := pc.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// SkipIfNoDocker skips the test if Docker is not available
func SkipIfNoDocker(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		t.Skip("Skipping test: SKIP_DOCKER_TESTS=true")
	}

	// Quick check if Docker is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		t.Skipf("Skipping test: Docker not available: %v", err)
	}
	defer provider.Close()

	if _, err := provider.Client().Ping(ctx); err != nil {
		t.Skipf("Skipping test: Docker daemon not responding: %v", err)
	}
}
