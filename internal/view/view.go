// Package view provee el layout HTML compartido por todos los dominios.
// Aquí vive una sola vez el <head> con SEO, la navegación y el pie comunes;
// cada dominio aporta únicamente su bloque "content".
package view

import (
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed base.html
var baseFS embed.FS

// Meta son los datos de cabecera (SEO) que cada página rellena.
type Meta struct {
	Title       string
	Description string
	Canonical   string
	OGType      string // "website", "article", ...
	JSONLD      template.JS
}

// funcs son los helpers disponibles en todas las plantillas.
var funcs = template.FuncMap{
	// tojson serializa un valor a JSON para incrustarlo en la página (p. ej. un
	// data island que lee el frontend). json.Marshal escapa <, > y & por defecto,
	// así que es seguro dentro de <script> (no se puede cerrar la etiqueta). Mismo
	// patrón que el JSON-LD del <head>.
	"tojson": func(v any) (template.JS, error) {
		b, err := json.Marshal(v)
		return template.JS(b), err
	},
}

// Layout compone el layout base con las plantillas de contenido del dominio.
// contentFS es el embed.FS del dominio; files son rutas dentro de él
// (cada una debe definir el bloque "content").
func Layout(contentFS fs.FS, files ...string) *template.Template {
	t := template.Must(template.New("layout").Funcs(funcs).ParseFS(baseFS, "base.html"))
	return template.Must(t.ParseFS(contentFS, files...))
}

// Render ejecuta el layout "base" con los datos de la página. Renderiza primero
// a un buffer: si la plantilla falla a mitad, aún podemos responder 500 limpio
// en vez de un 200 con HTML truncado ya enviado al cliente.
func Render(w http.ResponseWriter, t *template.Template, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "error de plantilla", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
