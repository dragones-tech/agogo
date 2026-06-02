// Package middleware reúne el baseline de seguridad del NÚCLEO (siempre activo):
// recuperación de panics y cabeceras de seguridad. (El logging de acceso vive en
// el módulo opt-in internal/logs.)
package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
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

// maxBodyBytes es el tope de cuerpo aceptado en una petición (1 MiB). Los
// formularios del sitio son pequeños; un cuerpo mayor es error o abuso.
const maxBodyBytes = 1 << 20

// LimitBody acota el cuerpo de cada petición antes de que cualquier handler lo
// lea (ParseForm, json.Decode...), evitando agotar memoria con un POST enorme.
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// gzipPool reutiliza los *gzip.Writer entre peticiones (evita realocar el buffer
// interno en cada respuesta).
var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

// compressible indica si vale la pena comprimir un content-type (texto). Las
// imágenes, fuentes y zips ya vienen comprimidos: gzip-earlos solo gasta CPU.
func compressible(ct string) bool {
	ct = strings.ToLower(ct)
	switch {
	case strings.HasPrefix(ct, "text/"),
		strings.Contains(ct, "json"),
		strings.Contains(ct, "javascript"),
		strings.Contains(ct, "xml"),
		strings.Contains(ct, "svg"):
		return true
	}
	return false
}

// Gzip comprime la respuesta con gzip cuando el cliente lo acepta y el contenido
// es de texto (HTML/CSS/JS/JSON/XML). Decide al fijarse las cabeceras según el
// Content-Type que puso el handler, así no recomprime imágenes. Código nuestro
// sobre la stdlib (compress/gzip). En producción esto suele terminarse en el
// borde (Caddy: encode gzip zstd); este middleware sirve al binario pelón.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sin soporte del cliente, o petición Range (gzip rompería los rangos):
		// se sirve sin comprimir.
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || r.Header.Get("Range") != "" {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		gw := &gzipResponseWriter{ResponseWriter: w}
		defer gw.close()
		next.ServeHTTP(gw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	wroteHeader bool
	plain       bool // true → escribir sin comprimir (tipo no comprimible)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	// Si ya viene codificado o no es texto, pasa tal cual.
	if w.Header().Get("Content-Encoding") != "" || !compressible(w.Header().Get("Content-Type")) {
		w.plain = true
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.Header().Del("Content-Length") // dejará de ser válido al comprimir
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(w.ResponseWriter)
	w.gz = gz
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// La mayoría de handlers fijan el Content-Type antes del primer Write;
		// si no, lo deducimos como hace net/http.
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.WriteHeader(http.StatusOK)
	}
	if w.plain {
		return w.ResponseWriter.Write(b)
	}
	return w.gz.Write(b)
}

// Flush preserva http.Flusher (vacía gzip y luego el writer envuelto).
func (w *gzipResponseWriter) Flush() {
	if w.gz != nil {
		_ = w.gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *gzipResponseWriter) close() {
	if w.gz != nil {
		_ = w.gz.Close()
		gzipPool.Put(w.gz)
		w.gz = nil
	}
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
