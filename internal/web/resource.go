// Package web contiene Resource[T], el recurso de BD genérico (lista/detalle ×
// HTML/JSON), escrito una sola vez. NO conoce su URL: sus handlers se cablean en
// el module.go del dominio (r.Get(path, recurso.ListHTML), ...). El canonical y
// el JSON-LD se calculan desde la URL de la petición, así el mismo handler sirve
// en cualquier ruta donde lo montes.
package web

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net/http"

	"agogo/internal/respond"
	"agogo/internal/sitemap"
	"agogo/internal/view"
)

// Page e ItemPage son los datos que viajan a las plantillas.
type Page[T any] struct {
	Meta  view.Meta
	Items []T
}
type ItemPage[T any] struct {
	Meta view.Meta
	Item T
}

// Resource describe el COMPORTAMIENTO de un recurso (queries, plantillas, SEO).
type Resource[T any] struct {
	BaseURL string

	List func(context.Context) ([]T, error)
	Get  func(context.Context, string) (T, error)
	Slug func(T) string

	TplList, TplItem *template.Template
	ListMeta         func(url string) view.Meta
	ItemMeta         func(it T, url string) view.Meta

	SitemapFreq string
	SitemapPrio string
}

func (r Resource[T]) url(req *http.Request) string { return r.BaseURL + req.URL.Path }

func (r Resource[T]) ListHTML(w http.ResponseWriter, req *http.Request) {
	items, err := r.List(req.Context())
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	view.Render(w, req, r.TplList, Page[T]{Meta: r.ListMeta(r.url(req)), Items: items})
}

func (r Resource[T]) DetailHTML(w http.ResponseWriter, req *http.Request) {
	it, err := r.Get(req.Context(), req.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, req)
		return
	} else if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	view.Render(w, req, r.TplItem, ItemPage[T]{Meta: r.ItemMeta(it, r.url(req)), Item: it})
}

func (r Resource[T]) ListJSON(w http.ResponseWriter, req *http.Request) {
	items, err := r.List(req.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "error interno")
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

func (r Resource[T]) DetailJSON(w http.ResponseWriter, req *http.Request) {
	it, err := r.Get(req.Context(), req.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "no encontrado")
		return
	} else if err != nil {
		respond.Error(w, http.StatusInternalServerError, "error interno")
		return
	}
	respond.JSON(w, http.StatusOK, it)
}

// SitemapSource crea la fuente de sitemap del recurso: la entrada de la lista
// (listPath) más una por ítem bajo itemBase. Las rutas se las pasas en module.go.
func (r Resource[T]) SitemapSource(listPath, itemBase string) sitemap.Source {
	return resourceSitemap[T]{res: r, listPath: listPath, itemBase: itemBase}
}

type resourceSitemap[T any] struct {
	res                Resource[T]
	listPath, itemBase string
}

func (s resourceSitemap[T]) SitemapURLs(ctx context.Context) ([]sitemap.URL, error) {
	items, err := s.res.List(ctx)
	if err != nil {
		return nil, err
	}
	urls := []sitemap.URL{{Path: s.listPath, ChangeFreq: "daily", Priority: "0.9"}}
	urls = append(urls, sitemap.Entries(items, s.itemBase+"/", s.res.SitemapFreq, s.res.SitemapPrio, s.res.Slug)...)
	return urls, nil
}
