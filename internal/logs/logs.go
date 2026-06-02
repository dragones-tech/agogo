// Package logs is an opt-in MODULE: it registers a global middleware that logs
// each request (method, path, status, duration). If you don't plug it in with
// app.Use(logs.Module()), there are no access logs and the code doesn't even
// enter the binary.
package logs

import (
	"log"
	"net/http"
	"time"

	"agogo/internal/app"
)

func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "logs" }

func (mod) Register(a *app.App) error {
	a.UseMiddleware(logging)
	return nil
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// statusRecorder remembers the status code so it can be logged.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Flush preserves the wrapped ResponseWriter's http.Flusher interface
// (streaming, SSE); without this, wrapping it would hide the interface.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
