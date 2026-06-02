package paginas

import (
	"agogo/internal/app"
	"agogo/internal/sitemap"
)

// Module wires the example static section. Plug it in from main.go with one line.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "paginas" }

func (mod) Register(a *app.App) error {
	a.Router.Get("/ejemplo", Example(a.Config.BaseURL))
	a.AddSitemap(sitemap.StaticURLs("/ejemplo")) // breadcrumb: this is how a section feeds sitemap.xml
	return nil
}
