package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"jehosogo/internal/app"
	"jehosogo/internal/auth"
	"jehosogo/internal/blog"
	"jehosogo/internal/config"
	"jehosogo/internal/contacto"
	"jehosogo/internal/logs"
	"jehosogo/internal/oauth"
	"jehosogo/internal/openapi"
	"jehosogo/internal/paginas"
	"jehosogo/internal/productos"
	"jehosogo/internal/site"

	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.DevSecret {
		log.Printf("AVISO: usando clave de sesión de desarrollo (define JEHOSOGO_SECRET_KEY en producción)")
	}

	sqldb, err := sql.Open("sqlite", cfg.DB)
	if err != nil {
		log.Fatalf("abrir db: %v", err)
	}
	defer sqldb.Close()

	application := app.New(cfg, sqldb)
	application.Router.Get("/salud", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	// El "Gemfile" del servidor: qué es esta app = qué módulos acopla.
	// Comenta una línea para quitar esa funcionalidad (y sacarla del binario).
	if err := application.Use(
		logs.Module(),      // observabilidad
		productos.Module(), // contenido
		blog.Module(),
		paginas.Module(),
		contacto.Module(), // formulario
		auth.Module(),     // autenticación usuario/contraseña (login/cuenta)
		oauth.Module(),    // autenticación vía OAuth 2.0 (reusa identity)
		openapi.Module(),  // docs de la API
		site.Module(),     // robots.txt, sitemap.xml, /static
	); err != nil {
		log.Fatalf("módulos: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           application.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("jehosogo escuchando en %s (db=%s)", cfg.Addr, cfg.DB)
	log.Fatal(srv.ListenAndServe())
}
