package blog

import (
	"agogo/internal/app"
	"agogo/internal/blog/db"
)

// Module acopla el dominio "blog": rutas, sitemap y migración.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "blog" }

func (mod) Register(a *app.App) error {
	res := New(db.New(a.DB), a.Config.BaseURL)

	r := a.Router
	r.Get("/blog", res.ListHTML)
	r.Get("/blog/{slug}", res.DetailHTML)
	r.Get("/api/posts", res.ListJSON)
	r.Get("/api/posts/{slug}", res.DetailJSON)

	a.AddSitemap(res.SitemapSource("/blog", "/blog"))
	a.AddMigration(Migrate)
	return nil
}
