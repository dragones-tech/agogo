// Package sitemap es el CONTRATO (hoja, sin dependencias internas) para que
// cualquier módulo aporte URLs al sitemap. El módulo site las sirve; el host app
// las recolecta. Vive aparte para que app y site no formen un ciclo de imports.
package sitemap

import "context"

// URL es una entrada de sitemap (Path relativo; el baseURL lo antepone site).
type URL struct {
	Path       string
	ChangeFreq string
	Priority   string
}

// Source lo implementa cualquier cosa que quiera aparecer en el sitemap.
type Source interface {
	SitemapURLs(ctx context.Context) ([]URL, error)
}

// Entries construye entradas por ítem bajo un prefijo (genérico, una sola vez).
func Entries[T any](items []T, prefix, changeFreq, priority string, slug func(T) string) []URL {
	urls := make([]URL, 0, len(items))
	for _, it := range items {
		urls = append(urls, URL{Path: prefix + slug(it), ChangeFreq: changeFreq, Priority: priority})
	}
	return urls
}

// StaticURLs crea una fuente para rutas fijas (páginas, formularios).
func StaticURLs(paths ...string) Source { return staticURLs(paths) }

type staticURLs []string

func (s staticURLs) SitemapURLs(ctx context.Context) ([]URL, error) {
	urls := make([]URL, 0, len(s))
	for _, p := range s {
		urls = append(urls, URL{Path: p, ChangeFreq: "yearly", Priority: "0.5"})
	}
	return urls, nil
}
