package tests

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection.
func SetupTestDB() (*sql.DB, error) {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "avito_user"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "avito_password"
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "avito_db"
	}

	// Connect
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Clean up
	if err := CleanupTestDB(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to cleanup database: %w", err)
	}

	return db, nil
}

// CleanupTestDB truncates all tables to clean up test data.
func CleanupTestDB(db *sql.DB) error {
	// Truncate tables in reverse order of dependencies
	tables := []string{
		"pr_reviewers",
		"pull_requests",
		"users",
		"teams",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}
