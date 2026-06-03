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
	"agogo/internal/otw"
	"agogo/internal/site"
	// Breadcrumbs — uncomment the import together with its app.Use line below:
	// "agogo/internal/paginas" // example static section (no DB) at /ejemplo
	// "agogo/internal/auth"    // username/password login (needs the DB; see cmd/migrate)
	// "agogo/internal/oauth"   // OAuth 2.0 login (no DB; configure via OAUTH_* env)

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
		home.Module(), // home: "hola mundo" + link to the docs + the otw demo
		otw.Module(),  // BFF "HTML over the wire": fragment from a token-gated API
		site.Module(), // robots.txt, sitemap.xml, /static, favicon, styled 404

		// Breadcrumbs — uncomment a line (and its import above) to plug it in.
		// What you don't wire in isn't imported, so it stays out of the binary.
		//
		// paginas.Module(), // simplest domain shape (no DB); copy it for your sections
		// oauth.Module(),   // OAuth login; routes answer 503 until OAUTH_* is set
		// auth.Module(),    // username/password login. Needs the DB: run
		//                   //   `go run ./cmd/migrate` (and `./cmd/seed` for a demo
		//                   //   user) once before enabling it.
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
