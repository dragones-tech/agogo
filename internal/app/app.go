// Package app es el NÚCLEO del framework: el host que compone módulos. Un
// servidor se arma creando un App y acoplándole módulos con Use(). Lo que no
// acoplas no se importa, así que no existe en el binario.
//
// El host ofrece servicios compartidos baratos (Session, Auth) que cualquier
// módulo puede usar; las funcionalidades visibles (login, contenido) son módulos.
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

// Middleware global: envuelve todo el servidor (a diferencia del por-ruta).
type Middleware = func(http.Handler) http.Handler

// Module es lo que un dominio o plugin implementa para acoplarse al servidor.
type Module interface {
	Name() string
	Register(*App) error
}

// App es el host: servicios compartidos + puntos de enganche para módulos.
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

// Use acopla módulos en orden. Si uno falla al registrarse, corta con su nombre.
func (a *App) Use(mods ...Module) error {
	for _, m := range mods {
		if err := m.Register(a); err != nil {
			return fmt.Errorf("módulo %q: %w", m.Name(), err)
		}
	}
	return nil
}

// --- puntos de enganche para módulos ---

func (a *App) UseMiddleware(mw Middleware) { a.middleware = append(a.middleware, mw) }
func (a *App) AddMigration(fn func(context.Context, *sql.DB) error) {
	a.migrations = append(a.migrations, fn)
}
func (a *App) AddSitemap(src sitemap.Source)    { a.sitemaps = append(a.sitemaps, src) }
func (a *App) SitemapSources() []sitemap.Source { return a.sitemaps }

// Migrate corre las migraciones de todos los módulos acoplados.
func (a *App) Migrate(ctx context.Context) error {
	for _, fn := range a.migrations {
		if err := fn(ctx, a.DB); err != nil {
			return err
		}
	}
	return nil
}

// Handler arma el http.Handler final: rutas + baseline de seguridad (núcleo,
// siempre) + middleware global de los módulos (p. ej. logs) como capa externa.
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
