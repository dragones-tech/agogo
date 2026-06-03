package auth

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed sql/schema.sql
var schemaSQL string

// Migrate creates the users table if it doesn't exist (idempotent).
func Migrate(ctx context.Context, sqldb *sql.DB) error {
	_, err := sqldb.ExecContext(ctx, schemaSQL)
	return err
}
