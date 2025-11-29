package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

type Database struct {
	conn *sql.DB
}

// NewDatabase establishes a connection to PostgreSQL
func NewDatabase(dbURL string) (*Database, error) {
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")
	return &Database{conn: conn}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.conn.Close()
}

// Conn returns the underlying SQL connection for query execution
func (d *Database) Conn() *sql.DB {
	return d.conn
}

// RunMigrations executes SQL migration files from the migrations directory
func (d *Database) RunMigrations() error {
	// Try to find migrations directory
	var migrationsDir string

	// First, try relative path from current directory
	if info, err := os.Stat("migrations"); err == nil && info.IsDir() {
		migrationsDir = "migrations"
	} else if info, err := os.Stat("../migrations"); err == nil && info.IsDir() {
		migrationsDir = "../migrations"
	} else {
		// Fall back to looking in the expected location from binary
		wd, _ := os.Getwd()
		log.Printf("Working directory: %s", wd)
		migrationsDir = "migrations"
		log.Printf("Note: Could not find migrations directory, attempting with: %s", migrationsDir)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			filePath := filepath.Join(migrationsDir, entry.Name())
			sqlBytes, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
			}

			_, err = d.conn.Exec(string(sqlBytes))
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", filePath, err)
			}

			log.Printf("✓ Executed migration: %s", entry.Name())
		}
	}

	return nil
}
