// Command migrate aplica el esquema de todos los dominios a la base de datos.
// Es un paso de bootstrap (sirve en dev y en despliegue), separado del servidor.
//
//	go run ./cmd/migrate
//
// Arma el mismo App que el servidor y corre app.Migrate: cada módulo acoplado
// registró su migración con a.AddMigration, así que añadir un dominio con BD a
// la lista de abajo basta para que su esquema se aplique (sin tocar nada más).
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
	sqldb, err := sql.Open("sqlite", cfg.DB)
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
