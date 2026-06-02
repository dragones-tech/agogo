package productos

import (
	"database/sql"
	"embed"
	"net/http"
	"strings"

	"agogo/internal/jsonld"
	"agogo/internal/productos/db"
	"agogo/internal/respond"
	"agogo/internal/view"
	"agogo/internal/web"
)

//go:embed templates/*.html
var tplFS embed.FS

var (
	tplList = view.Layout(tplFS, "templates/list.html")
	tplItem = view.Layout(tplFS, "templates/detail.html")
)

// SearchJSON serves /api/productos: returns everything, or filters by ?q= (LIKE
// on title and description). It's the endpoint the catalog filter queries on
// every keystroke. The search is parameterized SQL (sqlc), not in memory.
func SearchJSON(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		var (
			items []db.Producto
			err   error
		)
		if query == "" {
			items, err = q.ListProductos(r.Context())
		} else {
			items, err = q.SearchProductos(r.Context(), sql.NullString{String: query, Valid: true})
		}
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "error interno")
			return
		}
		respond.JSON(w, http.StatusOK, items)
	}
}

// New configures the BEHAVIOR of the "productos" resource. It doesn't know its
// URL: its handlers are wired up in its module.go.
func New(q *db.Queries, baseURL string) web.Resource[db.Producto] {
	return web.Resource[db.Producto]{
		BaseURL: baseURL,

		List: q.ListProductos,
		Get:  q.GetProducto,
		Slug: func(p db.Producto) string { return p.Slug },

		TplList: tplList,
		TplItem: tplItem,

		SitemapFreq: "weekly",
		SitemapPrio: "0.8",

		ListMeta: func(url string) view.Meta {
			return view.Meta{
				Title:       "Agogo — Catálogo de productos",
				Description: "Catálogo compacto de productos Agogo, servido con Go puro y sin dependencias de más.",
				Canonical:   url,
				OGType:      "website",
			}
		},
		ItemMeta: func(p db.Producto, url string) view.Meta {
			return view.Meta{
				Title:       p.Titulo + " — Agogo",
				Description: p.Descripcion,
				Canonical:   url,
				OGType:      "website",
				JSONLD: jsonld.Product{
					Name:        p.Titulo,
					Description: p.Descripcion,
					URL:         url,
					Price:       p.Precio,
					Currency:    "MXN",
				}.Script(),
			}
		},
	}
}
