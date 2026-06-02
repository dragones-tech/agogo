package productos

import (
	"agogo/internal/app"
	"agogo/internal/productos/db"
)

// Module wires the "productos" domain into the server: registers its routes,
// its sitemap source and its migration. It's the domain's "plugin" face.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "productos" }

func (mod) Register(a *app.App) error {
	q := db.New(a.DB)
	res := New(q, a.Config.BaseURL)

	r := a.Router
	r.Get("/{$}", res.ListHTML) // home (exact match of "/")
	r.Get("/productos/{slug}", res.DetailHTML)
	r.Get("/api/productos", SearchJSON(q)) // list + search by ?q=
	r.Get("/api/productos/{slug}", res.DetailJSON)

	a.AddSitemap(res.SitemapSource("/", "/productos"))
	a.AddMigration(Migrate)
	return nil
}
