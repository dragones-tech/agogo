// Package view provides the HTML layout shared by all domains. Here lives, once,
// the <head> with SEO, the common navigation and footer; each domain contributes
// only its "content" block.
package view

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

//go:embed base.html error.html
var baseFS embed.FS

// Meta is the header (SEO) data that each page fills in.
type Meta struct {
	Title       string
	Description string
	Canonical   string
	OGType      string // "website", "article", ...
	JSONLD      template.JS
}

// funcs are the helpers available in all templates.
var funcs = template.FuncMap{
	// tojson serializes a value to JSON to embed it in the page (e.g. a data
	// island the frontend reads). json.Marshal escapes <, > and & by default, so
	// it's safe inside <script> (the tag can't be closed). Same pattern as the
	// <head>'s JSON-LD.
	"tojson": func(v any) (template.JS, error) {
		b, err := json.Marshal(v)
		return template.JS(b), err
	},
}

// Layout composes the base layout with the domain's content templates.
// contentFS is the domain's embed.FS; files are paths within it
// (each must define the "content" block).
func Layout(contentFS fs.FS, files ...string) *template.Template {
	t := template.Must(template.New("layout").Funcs(funcs).ParseFS(baseFS, "base.html"))
	return template.Must(t.ParseFS(contentFS, files...))
}

// FragmentHeader is the header the client sends to request ONLY the content
// block (partial navigation), instead of the full page.
const FragmentHeader = "X-Fragment"

// Render executes the template with the page data. If the request carries
// X-Fragment (partial site navigation), it renders only the "fragment" block
// (title + content + scripts); otherwise the full page ("base"). This way a soft
// navigation reuses header/footer/CSS and only swaps <main>, but a direct hit, a
// crawler or no-JS receives the whole document (SEO intact).
//
// It renders to a buffer first: if the template fails midway, we respond a clean
// 500 instead of a 200 with truncated HTML already sent to the client.
func Render(w http.ResponseWriter, r *http.Request, t *template.Template, data any) {
	block := "base"
	if r.Header.Get(FragmentHeader) != "" {
		block = "fragment"
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, block, data); err != nil {
		http.Error(w, "error de plantilla", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Vary", FragmentHeader) // same URL, two responses depending on the header
	_, _ = buf.WriteTo(w)
}

// tplError is the error page with the site layout (header/footer/branding).
var tplError = Layout(baseFS, "error.html")

type errorPage struct {
	Meta    Meta
	Code    int
	Heading string
	Message string
}

// renderError writes an error page with the given status and the site layout.
// Unlike http.Error (plain text), it keeps header/footer and branding.
func renderError(w http.ResponseWriter, code int, heading, message string) {
	data := errorPage{Meta: Meta{Title: fmt.Sprintf("%d — Agogo", code)}, Code: code, Heading: heading, Message: message}
	var buf bytes.Buffer
	if err := tplError.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, message, code) // last resort if even the layout fails
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = buf.WriteTo(w)
}

// NotFound responds with a 404 using the site layout. For HTML handlers.
func NotFound(w http.ResponseWriter, _ *http.Request) {
	renderError(w, http.StatusNotFound, "No encontramos esa página",
		"El enlace puede estar roto o la página ya no existe.")
}

// ServerError logs the REAL error (with method and path, for debugging) and
// responds with a 500 using the site layout, without leaking internal details
// to the client.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error en %s %s: %v", r.Method, r.URL.Path, err)
	renderError(w, http.StatusInternalServerError, "Algo salió mal",
		"Tuvimos un problema de nuestro lado. Inténtalo de nuevo en un momento.")
}
