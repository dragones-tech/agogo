package oauth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"jehosogo/internal/identity"
	"jehosogo/internal/session"
)

// Prueba el flujo completo con un proveedor OAuth FALSO (httptest) y verifica
// que al final la sesión queda iniciada vía el servicio identity del núcleo.
func TestFlujoReusaIdentity(t *testing.T) {
	// --- proveedor falso ---
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("grant_type") != "authorization_code" || r.FormValue("code") != "abc" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok123"}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"email":"oauthuser@example.com"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sess := session.NewManager([]byte("0123456789abcdef0123456789abcdef"))
	id := identity.New(sess)
	h := newHandler(provider{
		clientID:     "cid",
		clientSecret: "sec",
		authURL:      srv.URL + "/authorize",
		tokenURL:     srv.URL + "/token",
		userinfoURL:  srv.URL + "/userinfo",
		redirectURL:  "http://localhost:8080/oauth/callback",
		scope:        "email",
	}, sess, id)

	// 1) login → redirige al proveedor con state, y guarda el state en sesión.
	rec := httptest.NewRecorder()
	h.login(rec, httptest.NewRequest(http.MethodGet, "/oauth/login", nil))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d, quería 303", rec.Code)
	}
	loc, _ := url.Parse(rec.Result().Header.Get("Location"))
	state := loc.Query().Get("state")
	if state == "" || loc.Query().Get("client_id") != "cid" {
		t.Fatalf("URL de autorización mal formada: %s", loc)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no se guardó la sesión con el state")
	}

	// 2) callback con state + code → intercambia token, obtiene usuario, login.
	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc&state="+url.QueryEscape(state), nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	h.callback(rec2, req)
	if rec2.Code != http.StatusSeeOther {
		t.Fatalf("callback status = %d, quería 303", rec2.Code)
	}
	if loc := rec2.Result().Header.Get("Location"); loc != "/oauth/me" {
		t.Fatalf("redirect = %q, quería /oauth/me", loc)
	}

	// 3) la sesión resultante tiene la identidad del proveedor (vía identity).
	req3 := httptest.NewRequest(http.MethodGet, "/oauth/me", nil)
	for _, c := range rec2.Result().Cookies() {
		req3.AddCookie(c)
	}
	if got := id.UserID(req3); got != "oauthuser@example.com" {
		t.Fatalf("identidad = %q, quería oauthuser@example.com", got)
	}
}

// Sin configurar, /oauth/login responde 503 (módulo acoplado pero inerte).
func TestLoginSinConfigurar(t *testing.T) {
	sess := session.NewManager([]byte("0123456789abcdef0123456789abcdef"))
	h := newHandler(provider{}, sess, identity.New(sess))
	rec := httptest.NewRecorder()
	h.login(rec, httptest.NewRequest(http.MethodGet, "/oauth/login", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, quería 503", rec.Code)
	}
}
