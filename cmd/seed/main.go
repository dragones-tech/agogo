// Command seed populates the database with sample data for DEVELOPMENT.
// It's not part of the server. It ensures the schema (via app.Migrate,
// idempotent) and then seeds each DB-backed module.
//
//	go run ./cmd/seed
//
// The starter ships only `auth` (a demo user). Add your own modules' Seed calls
// here as you create them.
package main

import (
	"context"
	"database/sql"
	"log"

	"agogo/internal/app"
	"agogo/internal/auth"
	authdb "agogo/internal/auth/db"
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

	ctx := context.Background()

	// Ensure the schema before seeding (idempotent), reusing the host's
	// migration hook.
	application := app.New(cfg, sqldb)
	if err := application.Use(
		auth.Module(),
	); err != nil {
		log.Fatalf("módulos: %v", err)
	}
	must(application.Migrate(ctx), "migrar")

	must(auth.Seed(ctx, authdb.New(sqldb)), "sembrar auth")
	log.Printf("datos de ejemplo sembrados en %s (usuario demo: admin@agogo.com / demo1234)", cfg.DB)
}

func must(err error, what string) {
	if err != nil {
		log.Fatalf("%s: %v", what, err)
	}
}
