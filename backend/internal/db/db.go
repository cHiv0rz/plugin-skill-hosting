// Package db owns the connection pool and the embedded migration scripts.
package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

//go:embed migrations/0001_init.sql
var migration0001 string

//go:embed migrations/0002_oidc.sql
var migration0002 string

//go:embed migrations/0003_api_token.sql
var migration0003 string

//go:embed migrations/0004_skill_audit.sql
var migration0004 string

//go:embed migrations/0005_plugin_soft_delete.sql
var migration0005 string

//go:embed migrations/0006_skill_files.sql
var migration0006 string

//go:embed migrations/0007_skill_file_versions.sql
var migration0007 string

//go:embed migrations/0008_user_approval.sql
var migration0008 string

//go:embed migrations/0009_skill_extra_frontmatter.sql
var migration0009 string

//go:embed migrations/0010_user_admin.sql
var migration0010 string

func Open(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	// Tune the pool so connections don't sit idle long enough for a fronting
	// PgBouncer / managed-PG proxy to silently drop them. Without these,
	// database/sql happily reuses a "good" connection whose server-side peer
	// has already been reaped, producing intermittent "bad connection" /
	// "EOF" / "prepared statement does not exist" errors on the next query.
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(30 * time.Second)
	db.SetConnMaxLifetime(5 * time.Minute)
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("ping db: %w", err)
}

func Migrate(db *sql.DB) error {
	if _, err := db.Exec(migration0001); err != nil {
		return fmt.Errorf("0001_init: %w", err)
	}
	if _, err := db.Exec(migration0002); err != nil {
		return fmt.Errorf("0002_oidc: %w", err)
	}
	if _, err := db.Exec(migration0003); err != nil {
		return fmt.Errorf("0003_api_token: %w", err)
	}
	if _, err := db.Exec(migration0004); err != nil {
		return fmt.Errorf("0004_skill_audit: %w", err)
	}
	if _, err := db.Exec(migration0005); err != nil {
		return fmt.Errorf("0005_plugin_soft_delete: %w", err)
	}
	if _, err := db.Exec(migration0006); err != nil {
		return fmt.Errorf("0006_skill_files: %w", err)
	}
	if _, err := db.Exec(migration0007); err != nil {
		return fmt.Errorf("0007_skill_file_versions: %w", err)
	}
	if _, err := db.Exec(migration0008); err != nil {
		return fmt.Errorf("0008_user_approval: %w", err)
	}
	if _, err := db.Exec(migration0009); err != nil {
		return fmt.Errorf("0009_skill_extra_frontmatter: %w", err)
	}
	if _, err := db.Exec(migration0010); err != nil {
		return fmt.Errorf("0010_user_admin: %w", err)
	}
	return nil
}

// Exec is the subset of *sql.DB / *sql.Tx that the application reaches for.
// Anything that takes Exec can run inside or outside a transaction.
type Exec interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
