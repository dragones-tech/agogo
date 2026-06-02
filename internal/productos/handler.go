package productos

import (
	"embed"

	"jehosogo/internal/jsonld"
	"jehosogo/internal/productos/db"
	"jehosogo/internal/view"
	"jehosogo/internal/web"
)

//go:embed templates/*.html
var tplFS embed.FS

var (
	tplList = view.Layout(tplFS, "templates/list.html")
	tplItem = view.Layout(tplFS, "templates/detail.html")
)

// New configura el COMPORTAMIENTO del recurso "productos". No conoce su URL:
// sus handlers se cablean en routes.go.
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
				Title:       "Jehosogo — Catálogo de productos",
				Description: "Catálogo compacto de productos Jehosogo, servido con Go puro y sin dependencias de más.",
				Canonical:   url,
				OGType:      "website",
			}
		},
		ItemMeta: func(p db.Producto, url string) view.Meta {
			return view.Meta{
				Title:       p.Titulo + " — Jehosogo",
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
