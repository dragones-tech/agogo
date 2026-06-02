package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"agogo/internal/app"
	"agogo/internal/auth"
	"agogo/internal/blog"
	"agogo/internal/config"
	"agogo/internal/contacto"
	"agogo/internal/logs"
	"agogo/internal/oauth"
	"agogo/internal/openapi"
	"agogo/internal/paginas"
	"agogo/internal/productos"
	"agogo/internal/site"

	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.DevSecret {
		log.Printf("AVISO: usando clave de sesión de desarrollo (define AGOGO_SECRET_KEY en producción)")
	}

	sqldb, err := sql.Open("sqlite", cfg.DSN())
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

	// Apagado ordenado: al recibir Ctrl+C o SIGTERM dejamos de aceptar, damos
	// hasta 10s a que terminen las peticiones en curso y SOLO entonces volvemos
	// (los defer —incluido sqldb.Close()— corren; con log.Fatal/os.Exit no lo
	// harían). Stdlib pura.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("agogo escuchando en %s (db=%s)", cfg.Addr, cfg.DB)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("escuchar: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("apagando…")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("apagado forzado: %v", err)
	}
}
