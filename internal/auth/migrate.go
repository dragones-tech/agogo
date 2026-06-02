package auth

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed sql/schema.sql
var schemaSQL string

// Migrate crea la tabla de usuarios si no existe (idempotente).
func Migrate(ctx context.Context, sqldb *sql.DB) error {
	_, err := sqldb.ExecContext(ctx, schemaSQL)
	return err
}
