// Package paginas define handlers de páginas estáticas (sin BD). No conocen su
// URL: el canonical se calcula desde la ruta en que las montes (routes.go).
package paginas

import (
	"embed"
	"net/http"

	"jehosogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

// QuienesSomos devuelve el handler de la página "quiénes somos".
func QuienesSomos(baseURL string) http.HandlerFunc {
	tpl := view.Layout(tplFS, "templates/quienes-somos.html")
	return func(w http.ResponseWriter, r *http.Request) {
		view.Render(w, tpl, struct{ Meta view.Meta }{view.Meta{
			Title:       "Quiénes somos — Jehosogo",
			Description: "Qué es Jehosogo y por qué lo construimos con Go puro, sin dependencias de más.",
			Canonical:   baseURL + r.URL.Path,
			OGType:      "website",
		}})
	}
}
