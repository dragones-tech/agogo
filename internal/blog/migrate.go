package blog

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed sql/schema.sql
var schemaSQL string

// Migrate crea el esquema si no existe (idempotente). NO lo llama el servidor:
// se ejecuta con `go run ./cmd/migrate` (esquema) o `go run ./cmd/seed` (dev).
func Migrate(ctx context.Context, sqldb *sql.DB) error {
	_, err := sqldb.ExecContext(ctx, schemaSQL)
	return err
}
