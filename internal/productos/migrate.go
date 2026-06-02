package productos

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed sql/schema.sql
var schemaSQL string

// Migrate creates the schema if it doesn't exist (idempotent). The server does
// NOT call it: run it with `go run ./cmd/migrate` (schema) or `go run ./cmd/seed` (dev).
func Migrate(ctx context.Context, sqldb *sql.DB) error {
	_, err := sqldb.ExecContext(ctx, schemaSQL)
	return err
}
