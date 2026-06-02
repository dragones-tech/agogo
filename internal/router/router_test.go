package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Verifies that you can pass MULTIPLE middlewares per route and that they run in
// the listed order (the first being the outermost layer), before the handler.
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
	}, marca("A"), marca("B"), marca("C")) // ← three middlewares

	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if got := strings.Join(orden, ","); got != "A,B,C,handler" {
		t.Fatalf("orden = %q, quería \"A,B,C,handler\"", got)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

// A middleware that short-circuits (doesn't call next) stops the chain: the rest
// and the handler don't run.
func TestMiddlewareQueCorta(t *testing.T) {
	var corrioHandler bool
	corta := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden) // doesn't call next
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
