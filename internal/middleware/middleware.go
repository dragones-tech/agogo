// Package middleware reúne el baseline de seguridad del NÚCLEO (siempre activo):
// recuperación de panics y cabeceras de seguridad. (El logging de acceso vive en
// el módulo opt-in internal/logs.)
package middleware

import (
	"log"
	"net/http"
)

// Recover atrapa cualquier panic en un handler y responde 500 sin tumbar
// el servidor completo.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recuperado en %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "error interno", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders añade cabeceras de endurecimiento a todas las respuestas.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Todo es del mismo origen, así que default-src 'self' no rompe nada.
		h.Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}
