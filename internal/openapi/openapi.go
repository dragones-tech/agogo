// Package openapi serves the OpenAPI spec (hand-written, no deps) of the JSON
// face, and a Swagger UI at /docs. Swagger UI is VENDORED (embedded in
// internal/openapi/static), so /docs is self-contained: it loads nothing from a
// CDN and the CSP stays strict ('self').
package openapi

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed openapi.json
var spec []byte

//go:embed static
var staticFS embed.FS

// Spec serves openapi.json. Importable into Postman, Bruno, editor.swagger.io, etc.
func Spec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(spec)
}

// Assets serves the vendored Swagger UI files (css/js). Mount it in its
// module.go: r.Handle("GET /docs-assets/", openapi.Assets()).
func Assets() http.Handler {
	sub, _ := fs.Sub(staticFS, "static")
	return http.StripPrefix("/docs-assets/", http.FileServerFS(sub))
}

// docsHTML loads Swagger UI from OUR local routes (not a CDN).
const docsHTML = `<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>API — Agogo</title>
<link rel="stylesheet" href="/docs-assets/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="/docs-assets/swagger-ui-bundle.js"></script>
<script src="/docs-assets/init.js"></script>
</body>
</html>`

// Docs serves the Swagger UI. Everything is local (self-contained), so the CSP
// stays at 'self'; Swagger UI injects styles at runtime, which is why style-src
// allows 'unsafe-inline' (only on this route).
func Docs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(docsHTML))
}
