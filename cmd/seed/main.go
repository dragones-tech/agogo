// Command seed populates the database with sample data for DEVELOPMENT.
// It's not part of the server. It ensures the schema (via app.Migrate,
// idempotent) and then seeds each domain.
//
//	go run ./cmd/seed
package main

import (
	"context"
	"database/sql"
	"log"

	"agogo/internal/app"
	"agogo/internal/auth"
	authdb "agogo/internal/auth/db"
	"agogo/internal/blog"
	blogdb "agogo/internal/blog/db"
	"agogo/internal/config"
	"agogo/internal/contacto"
	"agogo/internal/productos"
	productosdb "agogo/internal/productos/db"

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
		productos.Module(),
		blog.Module(),
		contacto.Module(),
		auth.Module(),
	); err != nil {
		log.Fatalf("módulos: %v", err)
	}
	must(application.Migrate(ctx), "migrar")

	must(productos.Seed(ctx, productosdb.New(sqldb)), "sembrar productos")
	must(blog.Seed(ctx, blogdb.New(sqldb)), "sembrar blog")
	must(auth.Seed(ctx, authdb.New(sqldb)), "sembrar auth")
	log.Printf("datos de ejemplo sembrados en %s (usuario demo: admin@agogo.com / demo1234)", cfg.DB)
}

func must(err error, what string) {
	if err != nil {
		log.Fatalf("%s: %v", what, err)
	}
}
