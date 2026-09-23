package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// InitSchema creates the database tables defined in the schema file
// and applies migrations required by existing databases.
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

	if err := runMigrations(db); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}

	return nil
}

// runMigrations applies database schema changes required by existing databases.
func runMigrations(db *sql.DB) error {
	migrations := []struct {
		table      string
		column     string
		definition string
	}{
		{
			table:      "reports",
			column:     "response",
			definition: "TEXT NOT NULL DEFAULT ''",
		},
		{
			table:      "notifications",
			column:     "report_id",
			definition: "INTEGER",
		},
	}

	for _, migration := range migrations {
		exists, err := columnExists(
			db,
			migration.table,
			migration.column,
		)
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		query := fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s %s",
			migration.table,
			migration.column,
			migration.definition,
		)

		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf(
				"add %s.%s column: %w",
				migration.table,
				migration.column,
				err,
			)
		}
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_report_id
		ON notifications(report_id)
	`); err != nil {
		return fmt.Errorf(
			"create notifications report index: %w",
			err,
		)
	}

	return nil
}

// columnExists reports whether a column exists in the specified database table.
func columnExists(
	db *sql.DB,
	table string,
	column string,
) (bool, error) {
	if strings.TrimSpace(table) == "" ||
		strings.TrimSpace(column) == "" {
		return false, errors.New(
			"table and column names cannot be empty",
		)
	}

	rows, err := db.Query(
		fmt.Sprintf("PRAGMA table_info(%s)", table),
	)
	if err != nil {
		return false, fmt.Errorf(
			"get table info for %s: %w",
			table,
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid          int
			name         string
			columnType   string
			notNull      int
			defaultValue sql.NullString
			primaryKey   int
		)

		if err := rows.Scan(
			&cid,
			&name,
			&columnType,
			&notNull,
			&defaultValue,
			&primaryKey,
		); err != nil {
			return false, fmt.Errorf(
				"scan table info for %s: %w",
				table,
				err,
			)
		}

		if name == column {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf(
			"iterate table info for %s: %w",
			table,
			err,
		)
	}

	return false, nil
}