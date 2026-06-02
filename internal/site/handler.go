// Package site es un MÓDULO: sirve robots.txt, sitemap.xml y los archivos
// estáticos embebidos. El sitemap se arma con TODAS las fuentes que los demás
// módulos registraron en el App; se leen en CADA petición, así no importa el
// orden en que se acoplaron.
package site

import (
	"context"
	"embed"
	"encoding/xml"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"agogo/internal/app"
	"agogo/internal/sitemap"
	"agogo/internal/view"
)

func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "site" }

func (mod) Register(a *app.App) error {
	base := a.Config.BaseURL
	r := a.Router

	r.Get("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", base)
	})
	r.Get("/sitemap.xml", func(w http.ResponseWriter, req *http.Request) {
		writeSitemap(req.Context(), w, base, a.SitemapSources())
	})
	r.Handle("GET /static/", Static())

	r.Get("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		b, _ := staticFS.ReadFile("static/favicon.svg")
		_, _ = w.Write(b)
	})

	// Catch-all: cualquier ruta no registrada cae aquí → 404 con el layout del
	// sitio. "/{$}" (home) y las rutas específicas son más concretas y ganan;
	// "/" (subárbol) solo atrapa lo no emparejado.
	r.Handle("/", http.HandlerFunc(view.NotFound))
	return nil
}

//go:embed static
var staticFS embed.FS

// Static devuelve el handler de archivos estáticos embebidos (con caché).
func Static() http.Handler {
	static, _ := fs.Sub(staticFS, "static")
	return http.StripPrefix("/static/", cacheControl(http.FileServerFS(static)))
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		next.ServeHTTP(w, r)
	})
}

type urlset struct {
	XMLName xml.Name   `xml:"urlset"`
	Xmlns   string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

type urlEntry struct {
	Loc        string `xml:"loc"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func writeSitemap(ctx context.Context, w http.ResponseWriter, base string, sources []sitemap.Source) {
	doc := urlset{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, s := range sources {
		urls, err := s.SitemapURLs(ctx)
		if err != nil {
			log.Printf("sitemap: %v", err)
			http.Error(w, "error interno", http.StatusInternalServerError)
			return
		}
		for _, u := range urls {
			doc.URLs = append(doc.URLs, urlEntry{Loc: base + u.Path, ChangeFreq: u.ChangeFreq, Priority: u.Priority})
		}
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		log.Printf("sitemap encode: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
	}
}
