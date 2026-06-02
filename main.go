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
	"agogo/internal/otw"
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

	// The server's "Gemfile": what this app is = which modules it wires in.
	// Comment out a line to drop that feature (and strip it from the binary).
	if err := application.Use(
		logs.Module(),      // observability
		productos.Module(), // content
		blog.Module(),
		paginas.Module(),
		contacto.Module(), // form
		auth.Module(),     // username/password authentication (login/account)
		oauth.Module(),    // OAuth 2.0 authentication (reuses identity)
		otw.Module(),      // BFF: HTML over the wire from a token-gated external API
		openapi.Module(),  // API docs
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

	// Graceful shutdown: on Ctrl+C or SIGTERM we stop accepting, give in-flight
	// requests up to 10s to finish, and ONLY then return (so the defers —
	// including sqldb.Close() — run; with log.Fatal/os.Exit they wouldn't).
	// Pure stdlib.
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
