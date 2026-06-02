// Package paginas is an EXAMPLE of the simplest domain shape: a static page (no
// DB). It's wired (commented out) in main.go's app.Use(...) — uncomment it to
// serve /ejemplo. Copy this package to add static sections of your own: a route
// + a template, nothing else.
package paginas

import (
	"embed"
	"net/http"

	"agogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

// Example returns the handler for the demo static page. A page handler doesn't
// know its URL: the canonical is computed from the request path, so the same
// handler works on any route you mount it on (see module.go).
func Example(baseURL string) http.HandlerFunc {
	tpl := view.Layout(tplFS, "templates/example.html")
	return func(w http.ResponseWriter, r *http.Request) {
		view.Render(w, r, tpl, struct{ Meta view.Meta }{view.Meta{
			Title:       "Sección de ejemplo — agogo",
			Description: "Una sección estática de ejemplo.",
			Canonical:   baseURL + r.URL.Path,
			OGType:      "website",
		}})
	}
}
