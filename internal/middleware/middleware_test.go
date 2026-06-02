package middleware

import (
	"compress/gzip"
	"io"
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

func htmlHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
	})
}

func TestGzipComprimeTexto(t *testing.T) {
	body := "<h1>hola</h1>" + string(make([]byte, 1000)) // compresible
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	Gzip(htmlHandler(body)).ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, quería gzip", got)
	}
	if rec.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("falta Vary: Accept-Encoding")
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("cuerpo no es gzip válido: %v", err)
	}
	got, _ := io.ReadAll(gr)
	if string(got) != body {
		t.Errorf("el cuerpo descomprimido no coincide con el original")
	}
}

func TestGzipSinAcceptEncoding(t *testing.T) {
	rec := httptest.NewRecorder()
	// Sin cabecera Accept-Encoding: no debe comprimir.
	Gzip(htmlHandler("hola")).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("no debería comprimir sin Accept-Encoding")
	}
	if rec.Body.String() != "hola" {
		t.Fatalf("cuerpo = %q, quería \"hola\"", rec.Body.String())
	}
}

func TestGzipNoRecomprimeBinario(t *testing.T) {
	img := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47})
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x.png", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	Gzip(img).ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("no debería comprimir un image/png")
	}
}
