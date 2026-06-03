// Command migrate applies the schema of every DB-backed module to the database.
// It's a bootstrap step (used in dev and on deploy), separate from the server.
//
//	go run ./cmd/migrate
//
// It wires up an App and runs app.Migrate: each wired-in module registered its
// migration via a.AddMigration, so adding a DB-backed domain to the list below
// is enough for its schema to be applied (nothing else). The starter ships only
// `auth` as a DB example; add your own modules here as you create them.
package main

import (
	"context"
	"database/sql"
	"log"

	"agogo/internal/app"
	"agogo/internal/auth"
	"agogo/internal/config"

	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	sqldb, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("abrir db: %v", err)
	}
	defer sqldb.Close()

	application := app.New(cfg, sqldb)
	if err := application.Use(
		auth.Module(), // breadcrumb: add your DB-backed modules here too
	); err != nil {
		log.Fatalf("módulos: %v", err)
	}
	if err := application.Migrate(context.Background()); err != nil {
		log.Fatalf("migrar: %v", err)
	}
	log.Printf("esquema aplicado en %s", cfg.DB)
}
