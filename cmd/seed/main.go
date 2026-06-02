// Command seed puebla la base de datos con datos de ejemplo para DESARROLLO.
// No es parte del servidor. Asegura el esquema (idempotente) y luego siembra.
//
//	go run ./cmd/seed
package main

import (
	"context"
	"database/sql"
	"log"

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
	sqldb, err := sql.Open("sqlite", cfg.DB)
	if err != nil {
		log.Fatalf("abrir db: %v", err)
	}
	defer sqldb.Close()

	ctx := context.Background()
	// Aseguramos el esquema antes de sembrar (idempotente).
	must(productos.Migrate(ctx, sqldb), "migrar productos")
	must(blog.Migrate(ctx, sqldb), "migrar blog")
	must(contacto.Migrate(ctx, sqldb), "migrar contacto")
	must(auth.Migrate(ctx, sqldb), "migrar auth")
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
