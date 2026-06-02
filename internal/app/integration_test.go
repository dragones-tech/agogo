package app_test

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"agogo/internal/app"
	"agogo/internal/auth"
	"agogo/internal/blog"
	"agogo/internal/config"
	"agogo/internal/contacto"
	"agogo/internal/paginas"
	"agogo/internal/productos"
	"agogo/internal/site"

	_ "modernc.org/sqlite"
)

// newServer wires up the same App as main (the data modules + pages + site),
// migrates a temporary SQLite and returns a test server and its DB.
func newServer(t *testing.T) (*httptest.Server, *sql.DB) {
	t.Helper()
	cfg := config.Config{
		DB:        filepath.Join(t.TempDir(), "test.db"),
		BaseURL:   "http://test",
		Addr:      ":0",
		SecretKey: []byte("0123456789abcdef0123456789abcdef"),
	}
	db, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	a := app.New(cfg, db)
	if err := a.Use(
		productos.Module(), blog.Module(), paginas.Module(),
		contacto.Module(), auth.Module(), site.Module(),
	); err != nil {
		t.Fatal(err)
	}
	if err := a.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(a.Handler())
	t.Cleanup(srv.Close)
	return srv, db
}

func TestRutasResponden(t *testing.T) {
	srv, _ := newServer(t)
	casos := []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/blog", http.StatusOK},
		{"/contacto", http.StatusOK},
		{"/login", http.StatusOK},
		{"/quienes-somos", http.StatusOK},
		{"/api/productos", http.StatusOK},
		{"/robots.txt", http.StatusOK},
		{"/productos/inexistente", http.StatusNotFound}, // slug that doesn't exist
		{"/ruta-que-no-existe", http.StatusNotFound},    // catch-all
	}
	for _, c := range casos {
		res, err := http.Get(srv.URL + c.path)
		if err != nil {
			t.Fatalf("GET %s: %v", c.path, err)
		}
		res.Body.Close()
		if res.StatusCode != c.want {
			t.Errorf("GET %s = %d, quería %d", c.path, res.StatusCode, c.want)
		}
	}
}

func TestFragmentoVsPaginaCompleta(t *testing.T) {
	srv, _ := newServer(t)
	get := func(fragment bool) string {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/quienes-somos", nil)
		if fragment {
			req.Header.Set("X-Fragment", "1")
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return string(b)
	}
	if !strings.Contains(get(false), "<html") {
		t.Error("la página completa debería traer <html>")
	}
	frag := get(true)
	if strings.Contains(frag, "<html") {
		t.Error("el fragmento NO debería traer <html> (solo el contenido)")
	}
	if !strings.Contains(frag, "<title>") {
		t.Error("el fragmento debería traer <title> para actualizar la pestaña")
	}
}

// A server failure responds with a generic 500 WITHOUT leaking the real error to the client.
func TestErrorInternoNoFiltraDetalles(t *testing.T) {
	srv, db := newServer(t)
	db.Close() // now any query fails → forces the error path

	res, err := http.Get(srv.URL + "/api/productos")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, quería 500", res.StatusCode)
	}
	if !strings.Contains(string(body), "error interno") {
		t.Errorf("debería responder un mensaje genérico, got %q", body)
	}
	if strings.Contains(string(body), "closed") || strings.Contains(string(body), "sql") {
		t.Errorf("NO debe filtrar el error interno al cliente: %q", body)
	}
}
