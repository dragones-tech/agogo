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
	"agogo/internal/config"
	"agogo/internal/home"
	"agogo/internal/logs"
	"agogo/internal/site"
	// "agogo/internal/paginas" // ← breadcrumb: uncomment with its app.Use line below

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

	// The "Gemfile": what this app is = which modules it wires in, one line each.
	//
	// Breadcrumb — to ADD a section: copy internal/home (or internal/paginas),
	// adjust its Register (routes + template), and plug it in here with one line.
	// What you don't wire in isn't imported, so it stays out of the binary.
	if err := application.Use(
		logs.Module(), // observability (access log)
		home.Module(), // home: "hola mundo" + link to the docs
		site.Module(), // robots.txt, sitemap.xml, /static, favicon, styled 404
		// paginas.Module(), // ← example static section (no DB) at /ejemplo.
		//   Uncomment THIS line and its import above to activate it. It's the
		//   simplest domain shape; copy it as the template for your own sections.
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

	// Graceful shutdown: stop accepting on SIGINT/SIGTERM, give in-flight requests
	// up to 10s, and only then return so the defers run (incl. closing the DB).
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
