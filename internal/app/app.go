// Package app is the framework CORE: the host that composes modules. A server
// is assembled by creating an App and plugging modules into it with Use().
// Whatever you don't plug in isn't imported, so it doesn't exist in the binary.
//
// The host offers cheap shared services (Session, Auth) that any module can
// use; the visible features (login, content) are modules.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"agogo/internal/config"
	"agogo/internal/identity"
	"agogo/internal/middleware"
	"agogo/internal/router"
	"agogo/internal/session"
	"agogo/internal/sitemap"
)

// Middleware is global: it wraps the whole server (unlike the per-route kind).
type Middleware = func(http.Handler) http.Handler

// Module is what a domain or plugin implements to plug into the server.
type Module interface {
	Name() string
	Register(*App) error
}

// App is the host: shared services + hook points for modules.
type App struct {
	Config   config.Config
	DB       *sql.DB
	Router   *router.Router
	Session  *session.Manager
	Identity *identity.Service

	middleware []Middleware
	migrations []func(context.Context, *sql.DB) error
	sitemaps   []sitemap.Source
}

func New(cfg config.Config, db *sql.DB) *App {
	sess := session.NewManager(cfg.SecretKey)
	sess.Secure = cfg.Secure
	return &App{
		Config:   cfg,
		DB:       db,
		Router:   router.New(),
		Session:  sess,
		Identity: identity.New(sess),
	}
}

// Use plugs in modules in order. If one fails to register, it stops with its name.
func (a *App) Use(mods ...Module) error {
	for _, m := range mods {
		if err := m.Register(a); err != nil {
			return fmt.Errorf("módulo %q: %w", m.Name(), err)
		}
	}
	return nil
}

// --- hook points for modules ---

func (a *App) UseMiddleware(mw Middleware) { a.middleware = append(a.middleware, mw) }
func (a *App) AddMigration(fn func(context.Context, *sql.DB) error) {
	a.migrations = append(a.migrations, fn)
}
func (a *App) AddSitemap(src sitemap.Source)    { a.sitemaps = append(a.sitemaps, src) }
func (a *App) SitemapSources() []sitemap.Source { return a.sitemaps }

// Migrate runs the migrations of all plugged-in modules.
func (a *App) Migrate(ctx context.Context) error {
	for _, fn := range a.migrations {
		if err := fn(ctx, a.DB); err != nil {
			return err
		}
	}
	return nil
}

// Handler assembles the final http.Handler: routes + security baseline (core,
// always) + the modules' global middleware (e.g. logs) as the outer layer.
func (a *App) Handler() http.Handler {
	var h http.Handler = a.Router.Handler()
	h = middleware.Gzip(h)
	h = middleware.SecurityHeaders(h)
	h = middleware.LimitBody(h)
	h = middleware.Recover(h)
	for i := len(a.middleware) - 1; i >= 0; i-- {
		h = a.middleware[i](h)
	}
	return h
}
