package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverDevuelve500(t *testing.T) {
	panicky := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	rec := httptest.NewRecorder()
	// No debe propagar el panic (si lo hiciera, el test entra en pánico).
	Recover(panicky).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, quería %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestSecurityHeaders(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	SecurityHeaders(ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, quería \"nosniff\"", got)
	}
}
