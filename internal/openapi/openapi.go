// Package openapi sirve la especificación OpenAPI (escrita a mano, sin deps) de
// la cara JSON, y una UI de Swagger en /docs. Swagger UI está VENDORIZADO
// (embebido en internal/openapi/static), así /docs es self-contained: no carga
// nada de un CDN y la CSP se mantiene estricta ('self').
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

// Spec sirve openapi.json. Importable en Postman, Bruno, editor.swagger.io, etc.
func Spec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(spec)
}

// Assets sirve los archivos vendorizados de Swagger UI (css/js). Móntalo en
// routes.go: r.Handle("GET /docs-assets/", openapi.Assets()).
func Assets() http.Handler {
	sub, _ := fs.Sub(staticFS, "static")
	return http.StripPrefix("/docs-assets/", http.FileServerFS(sub))
}

// docsHTML carga Swagger UI desde NUESTRAS rutas locales (no un CDN).
const docsHTML = `<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>API — Jehosogo</title>
<link rel="stylesheet" href="/docs-assets/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="/docs-assets/swagger-ui-bundle.js"></script>
<script src="/docs-assets/init.js"></script>
</body>
</html>`

// Docs sirve la UI de Swagger. Todo es local (self-contained), así la CSP queda
// en 'self'; Swagger UI inyecta estilos en runtime, por eso style-src permite
// 'unsafe-inline' (solo en esta ruta).
func Docs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(docsHTML))
}
