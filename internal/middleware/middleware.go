// Package middleware gathers the CORE security baseline (always active): panic
// recovery and security headers. (Access logging lives in the opt-in module
// internal/logs.)
package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// Recover catches any panic in a handler and responds with a 500 without
// bringing down the whole server.
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

// maxBodyBytes is the body cap accepted on a request (1 MiB). The site's forms
// are small; a larger body is an error or abuse.
const maxBodyBytes = 1 << 20

// LimitBody caps the body of each request before any handler reads it
// (ParseForm, json.Decode...), avoiding exhausting memory with a huge POST.
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// gzipPool reuses the *gzip.Writer across requests (avoids reallocating the
// internal buffer on each response).
var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

// compressible reports whether a content-type is worth compressing (text).
// Images, fonts and zips already come compressed: gzipping them only wastes CPU.
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

// Gzip compresses the response with gzip when the client accepts it and the
// content is text (HTML/CSS/JS/JSON/XML). It decides when headers are set, based
// on the Content-Type the handler put, so it doesn't recompress images. Our own
// code on top of the stdlib (compress/gzip). In production this is usually
// terminated at the edge (Caddy: encode gzip zstd); this middleware serves the
// bare binary.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Without client support, or a Range request (gzip would break the
		// ranges): served uncompressed.
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
	plain       bool // true → write uncompressed (non-compressible type)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	// If it's already encoded or not text, pass it through as is.
	if w.Header().Get("Content-Encoding") != "" || !compressible(w.Header().Get("Content-Type")) {
		w.plain = true
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.Header().Del("Content-Length") // will no longer be valid once compressed
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(w.ResponseWriter)
	w.gz = gz
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// Most handlers set the Content-Type before the first Write; if not,
		// we deduce it the way net/http does.
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

// Flush preserves http.Flusher (flushes gzip and then the wrapped writer).
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

// SecurityHeaders adds hardening headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Self-hosted by default; scc (CSS) and lumen (JS) load from jsDelivr, so
		// style-src/script-src also allow that CDN. To go back to a strict 'self',
		// self-host those assets and drop the CDN from here.
		h.Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' https://cdn.jsdelivr.net; script-src 'self' https://cdn.jsdelivr.net")
		next.ServeHTTP(w, r)
	})
}
