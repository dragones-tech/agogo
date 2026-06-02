package productos

import (
	"jehosogo/internal/app"
	"jehosogo/internal/productos/db"
)

// Module acopla el dominio "productos" al servidor: registra sus rutas, su
// fuente de sitemap y su migración. Es la cara de "plugin" del dominio.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "productos" }

func (mod) Register(a *app.App) error {
	res := New(db.New(a.DB), a.Config.BaseURL)

	r := a.Router
	r.Get("/{$}", res.ListHTML) // home (match exacto de "/")
	r.Get("/productos/{slug}", res.DetailHTML)
	r.Get("/api/productos", res.ListJSON)
	r.Get("/api/productos/{slug}", res.DetailJSON)

	a.AddSitemap(res.SitemapSource("/", "/productos"))
	a.AddMigration(Migrate)
	return nil
}
