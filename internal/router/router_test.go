package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Verifica que se pueden pasar VARIOS middlewares por ruta y que corren en el
// orden listado (el primero es la capa más externa), antes del handler.
func TestVariosMiddlewaresEnOrden(t *testing.T) {
	var orden []string
	marca := func(nombre string) Middleware {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				orden = append(orden, nombre)
				next(w, r)
			}
		}
	}

	r := New()
	r.Get("/x", func(w http.ResponseWriter, r *http.Request) {
		orden = append(orden, "handler")
		w.WriteHeader(http.StatusOK)
	}, marca("A"), marca("B"), marca("C")) // ← tres middlewares

	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if got := strings.Join(orden, ","); got != "A,B,C,handler" {
		t.Fatalf("orden = %q, quería \"A,B,C,handler\"", got)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

// Un middleware que corta (no llama a next) detiene la cadena: los siguientes y
// el handler no corren.
func TestMiddlewareQueCorta(t *testing.T) {
	var corrioHandler bool
	corta := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden) // no llama a next
		}
	}

	r := New()
	r.Get("/x", func(w http.ResponseWriter, r *http.Request) {
		corrioHandler = true
	}, corta)

	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if corrioHandler {
		t.Fatal("el handler corrió pese a que el middleware cortó")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, quería 403", rec.Code)
	}
}
