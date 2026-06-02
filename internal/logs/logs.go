// Package logs es un MÓDULO opt-in: registra un middleware global que loguea
// cada petición (método, ruta, status, duración). Si no lo acoplas con
// app.Use(logs.Module()), no hay logs de acceso y el código ni entra al binario.
package logs

import (
	"log"
	"net/http"
	"time"

	"jehosogo/internal/app"
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

// statusRecorder recuerda el código de estado para poder registrarlo.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
