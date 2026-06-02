// Package oauth is an optional MODULE: login via OAuth 2.0 (authorization code
// flow), in pure stdlib. It shows how a login mechanism other than
// username/password REUSES the core identity service: it verifies with an
// external provider and then calls identity.Login to mark the session.
//
// It's configured by ITS OWN environment (a plugin brings its own config):
//
//	OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_AUTH_URL, OAUTH_TOKEN_URL,
//	OAUTH_USERINFO_URL, OAUTH_SCOPE
//
// If not configured, the routes exist but /oauth/login responds 503.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"agogo/internal/app"
	"agogo/internal/identity"
	"agogo/internal/session"
)

type provider struct {
	clientID, clientSecret         string
	authURL, tokenURL, userinfoURL string
	redirectURL, scope             string
}

func (p provider) configured() bool {
	return p.clientID != "" && p.authURL != "" && p.tokenURL != "" && p.userinfoURL != ""
}

type handler struct {
	p        provider
	sess     *session.Manager
	identity *identity.Service
	client   *http.Client
}

func newHandler(p provider, sess *session.Manager, id *identity.Service) *handler {
	return &handler{p: p, sess: sess, identity: id, client: &http.Client{Timeout: 10 * time.Second}}
}

func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "oauth" }

func (mod) Register(a *app.App) error {
	p := provider{
		clientID:     os.Getenv("OAUTH_CLIENT_ID"),
		clientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		authURL:      os.Getenv("OAUTH_AUTH_URL"),
		tokenURL:     os.Getenv("OAUTH_TOKEN_URL"),
		userinfoURL:  os.Getenv("OAUTH_USERINFO_URL"),
		redirectURL:  a.Config.BaseURL + "/oauth/callback",
		scope:        envOr("OAUTH_SCOPE", "openid email"),
	}
	h := newHandler(p, a.Session, a.Identity)
	r := a.Router
	r.Get("/oauth/login", h.login)
	r.Get("/oauth/callback", h.callback)
	r.Get("/oauth/me", h.me, a.Identity.Require) // protected: reuses identity.Require
	return nil
}

const stateKey = "oauth_state"

// login: generates an anti-CSRF state, stores it in the session and redirects to the provider.
func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	if !h.p.configured() {
		http.Error(w, "OAuth no configurado (define OAUTH_CLIENT_ID, OAUTH_AUTH_URL, ...)", http.StatusServiceUnavailable)
		return
	}
	state := randomState()
	s := h.sess.Get(r)
	s.Set(stateKey, state)
	h.sess.Save(w, s)

	q := url.Values{
		"response_type": {"code"},
		"client_id":     {h.p.clientID},
		"redirect_uri":  {h.p.redirectURL},
		"scope":         {h.p.scope},
		"state":         {state},
	}
	http.Redirect(w, r, h.p.authURL+"?"+q.Encode(), http.StatusSeeOther)
}

// callback: verifies the state, exchanges the code for a token, fetches the
// user and REUSES identity.Login to start the session.
func (h *handler) callback(w http.ResponseWriter, r *http.Request) {
	s := h.sess.Get(r)
	want, got := s.Get(stateKey), r.URL.Query().Get("state")
	if want == "" || subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
		http.Error(w, "state inválido", http.StatusForbidden)
		return
	}
	// The state is single-use: consume it so it can't be reused. It's
	// persisted in the same Save as the login (LoginInto), not separately.
	s.Delete(stateKey)
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "falta code", http.StatusBadRequest)
		return
	}

	token, err := h.exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "error intercambiando code", http.StatusBadGateway)
		return
	}
	user, err := h.userinfo(r.Context(), token)
	if err != nil || user == "" {
		http.Error(w, "error obteniendo usuario", http.StatusBadGateway)
		return
	}

	h.identity.LoginInto(w, s, user) // ← REUSES the core identity service
	http.Redirect(w, r, "/oauth/me", http.StatusSeeOther)
}

func (h *handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store") // authenticated response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>Conectado vía OAuth</h1><p>Eres <strong>%s</strong>.</p>", html.EscapeString(h.identity.UserID(r)))
}

func (h *handler) exchange(ctx context.Context, code string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {h.p.redirectURL},
		"client_id":     {h.p.clientID},
		"client_secret": {h.p.clientSecret},
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, h.p.tokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token: status %d", resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("respuesta sin access_token")
	}
	return out.AccessToken, nil
}

func (h *handler) userinfo(ctx context.Context, token string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, h.p.userinfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("userinfo: status %d", resp.StatusCode)
	}
	var out struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Email != "" {
		return out.Email, nil
	}
	return out.Sub, nil
}

func randomState() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
