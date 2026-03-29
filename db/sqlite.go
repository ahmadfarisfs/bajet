package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

func ConnectSQLite(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := database.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return database, nil
}

func InitSQLiteSchema(database *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		currency_code TEXT NOT NULL DEFAULT 'USD',
		language TEXT NOT NULL DEFAULT 'en',
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS cycles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		total_budget INTEGER NOT NULL,
		period_count INTEGER NOT NULL DEFAULT 4,
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS periods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cycle_id INTEGER NOT NULL,
		period_number INTEGER NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		budget INTEGER NOT NULL,
		status TEXT NOT NULL,
		input_type TEXT,
		input_amount INTEGER NOT NULL DEFAULT 0,
		savings INTEGER NOT NULL DEFAULT 0,
		overspending INTEGER NOT NULL DEFAULT 0,
		completed_at TEXT,
		created_at TEXT NOT NULL,
		UNIQUE(cycle_id, period_number),
		FOREIGN KEY(cycle_id) REFERENCES cycles(id) ON DELETE CASCADE
	);
	`

	if _, err := database.Exec(query); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}

	// Best-effort migration from the previous schema version.
	_, _ = database.Exec(`ALTER TABLE users ADD COLUMN currency_code TEXT NOT NULL DEFAULT 'USD'`)
	_, _ = database.Exec(`ALTER TABLE users ADD COLUMN language TEXT NOT NULL DEFAULT 'en'`)
	_, _ = database.Exec(`ALTER TABLE cycles ADD COLUMN user_id INTEGER`)
	_, _ = database.Exec(`ALTER TABLE cycles ADD COLUMN period_count INTEGER NOT NULL DEFAULT 4`)
	// Drop the named unique window index if it exists.
	_, _ = database.Exec(`DROP INDEX IF EXISTS idx_cycles_user_window`)
	// Recreate cycles table without any UNIQUE constraint on (start_date, end_date).
	// This handles old DB files that had UNIQUE(start_date, end_date) as a table constraint.
	if err := migrateCyclesTable(database); err != nil {
		return fmt.Errorf("migrate cycles table: %w", err)
	}

	return nil
}

// migrateCyclesTable rewrites the cycles table if it still carries a UNIQUE
// constraint on (start_date, end_date) from an earlier schema version.
func migrateCyclesTable(db *sql.DB) error {
	var ddl string
	row := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='cycles'`)
	if err := row.Scan(&ddl); err != nil {
		return nil // table doesn't exist yet, nothing to do
	}

	// Only rewrite if the table still has a UNIQUE constraint on start_date/end_date.
	if !strings.Contains(ddl, "UNIQUE(start_date") && !strings.Contains(ddl, "UNIQUE (start_date") {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS cycles_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			total_budget INTEGER NOT NULL,
			period_count INTEGER NOT NULL DEFAULT 4,
			created_at TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`INSERT INTO cycles_new SELECT id, user_id, start_date, end_date, total_budget, period_count, created_at FROM cycles`,
		`DROP TABLE cycles`,
		`ALTER TABLE cycles_new RENAME TO cycles`,
	}
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			return fmt.Errorf("migrate cycles: %w", err)
		}
	}
	return tx.Commit()
}
