package blog

import (
	"embed"

	"jehosogo/internal/blog/db"
	"jehosogo/internal/jsonld"
	"jehosogo/internal/view"
	"jehosogo/internal/web"
)

//go:embed templates/*.html
var tplFS embed.FS

var (
	tplList = view.Layout(tplFS, "templates/list.html")
	tplItem = view.Layout(tplFS, "templates/detail.html")
)

// New configura el COMPORTAMIENTO del recurso "blog". No conoce su URL.
func New(q *db.Queries, baseURL string) web.Resource[db.Post] {
	return web.Resource[db.Post]{
		BaseURL: baseURL,

		List: q.ListPosts,
		Get:  q.GetPost,
		Slug: func(p db.Post) string { return p.Slug },

		TplList: tplList,
		TplItem: tplItem,

		SitemapFreq: "monthly",
		SitemapPrio: "0.6",

		ListMeta: func(url string) view.Meta {
			return view.Meta{
				Title:       "Blog — Jehosogo",
				Description: "Artículos y notas de Jehosogo.",
				Canonical:   url,
				OGType:      "website",
			}
		},
		ItemMeta: func(p db.Post, url string) view.Meta {
			return view.Meta{
				Title:       p.Titulo + " — Blog Jehosogo",
				Description: p.Resumen,
				Canonical:   url,
				OGType:      "article",
				JSONLD: jsonld.BlogPosting{
					Headline:      p.Titulo,
					Description:   p.Resumen,
					URL:           url,
					DatePublished: p.Publicado,
				}.Script(),
			}
		},
	}
}
