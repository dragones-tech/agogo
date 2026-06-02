package paginas

import (
	"jehosogo/internal/app"
	"jehosogo/internal/sitemap"
)

// Module acopla las páginas estáticas (sin BD).
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "paginas" }

func (mod) Register(a *app.App) error {
	a.Router.Get("/quienes-somos", QuienesSomos(a.Config.BaseURL))
	a.AddSitemap(sitemap.StaticURLs("/quienes-somos"))
	return nil
}
