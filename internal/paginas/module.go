package paginas

import (
	"agogo/internal/app"
	"agogo/internal/sitemap"
)

// Module wires up the static pages (no DB).
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "paginas" }

func (mod) Register(a *app.App) error {
	a.Router.Get("/quienes-somos", QuienesSomos(a.Config.BaseURL))
	a.AddSitemap(sitemap.StaticURLs("/quienes-somos"))
	return nil
}
