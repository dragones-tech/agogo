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

// FragmentHeader es la cabecera que el cliente manda para pedir SOLO el bloque
// de contenido (navegación parcial), en vez de la página completa.
const FragmentHeader = "X-Fragment"

// Render ejecuta la plantilla con los datos de la página. Si la petición trae
// X-Fragment (navegación parcial del sitio), renderiza solo el bloque "fragment"
// (título + content + scripts); si no, la página completa ("base"). Así una
// navegación suave reusa header/footer/CSS y solo cambia <main>, pero un acceso
// directo, un crawler o sin-JS reciben el documento entero (SEO intacto).
//
// Renderiza primero a un buffer: si la plantilla falla a mitad, respondemos 500
// limpio en vez de un 200 con HTML truncado ya enviado al cliente.
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
	w.Header().Add("Vary", FragmentHeader) // misma URL, dos respuestas según la cabecera
	_, _ = buf.WriteTo(w)
}
