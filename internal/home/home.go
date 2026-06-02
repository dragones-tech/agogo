// Package home is the starter's only active section: a "hello world" landing
// with a link to the docs. It's the smallest possible module — one route, one
// template. Copy it to start a section of your own; see main.go for the wiring.
package home

import (
	"embed"
	"net/http"

	"agogo/internal/app"
	"agogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

var tpl = view.Layout(tplFS, "templates/home.html")

// Module wires the home into the host. This is the whole pattern: implement
// Module{ Name() string; Register(*App) error } and register routes in Register.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "home" }

func (mod) Register(a *app.App) error {
	baseURL := a.Config.BaseURL
	a.Router.Get("/{$}", func(w http.ResponseWriter, r *http.Request) { // "/{$}" = exact "/"
		view.Render(w, r, tpl, struct{ Meta view.Meta }{view.Meta{
			Title:       "agogo",
			Description: "Un servidor web compacto en Go puro.",
			Canonical:   baseURL + r.URL.Path,
			OGType:      "website",
		}})
	})
	return nil
}
