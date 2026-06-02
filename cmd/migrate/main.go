// Command migrate applies the schema of every domain to the database.
// It's a bootstrap step (used in dev and deploy), separate from the server.
//
//	go run ./cmd/migrate
//
// It wires up the same App as the server and runs app.Migrate: each wired-in
// module registered its migration via a.AddMigration, so adding a domain with a
// DB to the list below is enough for its schema to be applied (nothing else).
package main

import (
	"context"
	"database/sql"
	"log"

	"agogo/internal/app"
	"agogo/internal/auth"
	"agogo/internal/blog"
	"agogo/internal/config"
	"agogo/internal/contacto"
	"agogo/internal/productos"

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
		productos.Module(),
		blog.Module(),
		contacto.Module(),
		auth.Module(),
	); err != nil {
		log.Fatalf("módulos: %v", err)
	}
	if err := application.Migrate(context.Background()); err != nil {
		log.Fatalf("migrar: %v", err)
	}
	log.Printf("esquema aplicado en %s", cfg.DB)
}
