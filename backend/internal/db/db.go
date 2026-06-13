// Package db opens the SQLite database and applies embedded migrations.
//
// The driver is modernc.org/sqlite — a pure-Go implementation that avoids
// CGO, which is critical for cross-compiling the backend in Phase 10
// (darwin/arm64, linux/amd64, windows/amd64) without a C toolchain.
package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers "sqlite" driver with database/sql

	"gotutor/backend/internal/db/migrations"
)

// Open returns a connected *sql.DB at dbPath, creating the file and parent
// directory if missing, then applying all embedded migrations in order.
//
// DSN flags enable WAL journaling, a 5s busy timeout, and FK enforcement —
// safe defaults for a single-process desktop app doing mostly reads with
// occasional writes from the verifier.
func Open(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	dsn := "file:" + url.PathEscape(dbPath) + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite handles concurrent reads well but writes should be serialized
	// at the connection level. SetMaxOpenConns(1) makes every write transaction
	// block at the connection pool, matching SQLite's single-writer model.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// applyMigrations reads each embedded .sql file in lexical order and
// executes its full text as one statement batch. SQLite's driver splits
// on `;` boundaries internally. Migrations must be idempotent (use
// CREATE TABLE IF NOT EXISTS, etc.) because we don't track applied state
// in v1 — re-applying on every startup is cheap for a 3-table schema.
func applyMigrations(db *sql.DB) error {
	migFS := migrations.FS()
	for _, name := range migrations.SortedNames() {
		stmt, err := fs.ReadFile(migFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.Exec(string(stmt)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}
