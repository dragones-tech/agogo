package app_test

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"agogo/internal/app"
	"agogo/internal/config"
	"agogo/internal/home"
	"agogo/internal/paginas"
	"agogo/internal/site"

	_ "modernc.org/sqlite"
)

// newServer wires the starter (home + the example static section + site) over a
// throwaway SQLite and returns a test server.
func newServer(t *testing.T) *httptest.Server {
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
	if err := a.Use(home.Module(), paginas.Module(), site.Module()); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(a.Handler())
	t.Cleanup(srv.Close)
	return srv
}

func TestRutasResponden(t *testing.T) {
	srv := newServer(t)
	casos := []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},                // home
		{"/ejemplo", http.StatusOK},         // example static section
		{"/robots.txt", http.StatusOK},      //
		{"/no-existe", http.StatusNotFound}, // styled catch-all 404
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

func TestHomeDiceHolaMundo(t *testing.T) {
	srv := newServer(t)
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "Hola, mundo") {
		t.Errorf("el home debería decir 'Hola, mundo', got %q", body)
	}
}
