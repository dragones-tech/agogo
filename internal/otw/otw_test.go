package otw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newHandler(apiURL, token string) *handler {
	return &handler{p: provider{apiURL: apiURL, token: token, client: &http.Client{Timeout: 5 * time.Second}}}
}

// The server fetches the token-gated external API and returns rendered HTML; the
// token must NEVER appear in the response sent to the client.
func TestPanelRenderizaDatosExternosSinFiltrarToken(t *testing.T) {
	ext := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Same shape as the GitHub repo API.
		_, _ = w.Write([]byte(`{"full_name":"acme/widget","description":"Un widget","stargazers_count":42,"forks_count":7,"open_issues_count":1,"language":"Go","license":{"spdx_id":"MIT"}}`))
	}))
	defer ext.Close()

	rec := httptest.NewRecorder()
	newHandler(ext.URL, "secret-token").panel(rec, httptest.NewRequest(http.MethodGet, "/otw/panel", nil))

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, quería 200", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Errorf("debería responder HTML, got %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(body, "acme/widget") || !strings.Contains(body, "42") {
		t.Errorf("debería renderizar los datos del repo (nombre y estrellas), got %q", body)
	}
	if strings.Contains(body, "secret-token") {
		t.Error("el token NO debe aparecer en el HTML enviado al cliente")
	}
}

// The built-in simulated API is token-gated, just like a real one would be.
func TestDemoAPIRequiresToken(t *testing.T) {
	h := newHandler("", "")

	noTok := httptest.NewRecorder()
	h.demoAPI(noTok, httptest.NewRequest(http.MethodGet, "/otw/demo-api", nil))
	if noTok.Code != http.StatusUnauthorized {
		t.Errorf("sin token debería dar 401, got %d", noTok.Code)
	}

	withTok := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/otw/demo-api", nil)
	req.Header.Set("Authorization", "Bearer "+demoToken)
	h.demoAPI(withTok, req)
	if withTok.Code != http.StatusOK || !strings.Contains(withTok.Body.String(), "dragones-tech/agogo") {
		t.Errorf("con token debería dar 200 con JSON (shape GitHub), got %d %q", withTok.Code, withTok.Body.String())
	}
}

// The BFF absorbs an upstream failure: friendly fragment, no token leak.
func TestPanelErrorExterno(t *testing.T) {
	ext := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ext.Close()

	rec := httptest.NewRecorder()
	newHandler(ext.URL, "secret-token").panel(rec, httptest.NewRequest(http.MethodGet, "/otw/panel", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, quería 200 (el BFF absorbe el fallo)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "no disponible") {
		t.Errorf("debería mostrar el fallback, got %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret-token") {
		t.Error("el token NO debe filtrarse ni en el camino de error")
	}
}
