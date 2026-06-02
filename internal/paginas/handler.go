// Package paginas defines static page handlers (no DB). They don't know their
// URL: the canonical is computed from the route you mount them on (its module.go).
package paginas

import (
	"embed"
	"net/http"

	"agogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

// QuienesSomos returns the handler for the "about us" page.
func QuienesSomos(baseURL string) http.HandlerFunc {
	tpl := view.Layout(tplFS, "templates/quienes-somos.html")
	return func(w http.ResponseWriter, r *http.Request) {
		view.Render(w, r, tpl, struct{ Meta view.Meta }{view.Meta{
			Title:       "Quiénes somos — Agogo",
			Description: "Qué es Agogo y por qué lo construimos con Go puro, sin dependencias de más.",
			Canonical:   baseURL + r.URL.Path,
			OGType:      "website",
		}})
	}
}
