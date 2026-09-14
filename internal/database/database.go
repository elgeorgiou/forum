package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Open creates and verifies a connection to the SQLite database.
func Open(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		return nil, errors.New("database path cannot be empty")
	}

	dir := filepath.Dir(dbPath)

	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	dsn := fmt.Sprintf(
		"file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL",
		dbPath,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// InitSchema creates the database tables defined in the schema file.
func InitSchema(db *sql.DB, schemaPath string) error {
	if db == nil {
		return errors.New("database cannot be nil")
	}

	if schemaPath == "" {
		return errors.New("schema path cannot be empty")
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read database schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("initialize database schema: %w", err)
	}

	return nil
}
