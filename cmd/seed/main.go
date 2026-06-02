// Command seed puebla la base de datos con datos de ejemplo para DESARROLLO.
// No es parte del servidor. Asegura el esquema (vía app.Migrate, idempotente) y
// luego siembra cada dominio.
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

	// Aseguramos el esquema antes de sembrar (idempotente), reusando el hook de
	// migraciones del host.
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
