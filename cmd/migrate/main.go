// Command migrate aplica el esquema de todos los dominios a la base de datos.
// Es un paso de bootstrap (sirve en dev y en despliegue), separado del servidor.
//
//	go run ./cmd/migrate
package main

import (
	"context"
	"database/sql"
	"log"

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
	sqldb, err := sql.Open("sqlite", cfg.DB)
	if err != nil {
		log.Fatalf("abrir db: %v", err)
	}
	defer sqldb.Close()

	ctx := context.Background()
	must(productos.Migrate(ctx, sqldb), "migrar productos")
	must(blog.Migrate(ctx, sqldb), "migrar blog")
	must(contacto.Migrate(ctx, sqldb), "migrar contacto")
	must(auth.Migrate(ctx, sqldb), "migrar auth")
	log.Printf("esquema aplicado en %s", cfg.DB)
}

func must(err error, what string) {
	if err != nil {
		log.Fatalf("%s: %v", what, err)
	}
}
