// Package otw is an opt-in module that demonstrates "HTML over the wire" with a
// server-side BFF (backend-for-frontend): the browser asks for a section, the
// server calls a token-gated EXTERNAL API (the token lives only on the server,
// never reaches the client), renders the JSON as an HTML fragment, and returns
// that HTML. The client just does innerHTML — no JSON, no token, no framework
// on the page.
//
// It's the right shape for "the data is behind a secret I can't expose": keep
// the secret and the rendering on the server; ship HTML.
//
// The concrete example fetches a GitHub repo's public stats (the GitHub API,
// which you call with a token —a PAT— for a higher rate limit or private repos).
// Configured by its own environment (a plugin brings its own config):
//
//	OTW_API_URL    e.g. https://api.github.com/repos/golang/go
//	OTW_API_TOKEN  the bearer token (a server-side secret; a GitHub PAT here)
//
// Without configuration it still works out of the box: the BFF points at a
// built-in SIMULATED endpoint (/otw/demo-api, also token-gated) that returns the
// SAME shape GitHub does, so you see the flow without credentials. In production
// you set OTW_API_URL to the real service and the simulated endpoint isn't used.
package otw

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"agogo/internal/app"
)

//go:embed templates/panel.html
var tplFS embed.FS

var tplPanel = template.Must(template.ParseFS(tplFS, "templates/panel.html"))

// demoToken gates the built-in simulated API. It's a fake secret that lives only
// on the server (the browser never sees it), exactly like a real one would.
const demoToken = "demo-token-solo-del-servidor"

type provider struct {
	apiURL string
	token  string
	demo   bool // true when pointing at the built-in simulated API
	client *http.Client
}

type handler struct{ p provider }

func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "otw" }

func (mod) Register(a *app.App) error {
	p := provider{
		apiURL: os.Getenv("OTW_API_URL"),
		token:  os.Getenv("OTW_API_TOKEN"),
		client: &http.Client{Timeout: 5 * time.Second},
	}
	// No real API configured → point the BFF at the built-in simulated one so the
	// demo works without credentials. (In production this loopback isn't used.)
	if p.apiURL == "" || p.token == "" {
		p.apiURL = a.Config.BaseURL + "/otw/demo-api"
		p.token = demoToken
		p.demo = true
	}

	h := &handler{p: p}
	a.Router.Get("/otw/panel", h.panel)       // BFF: returns rendered HTML
	a.Router.Get("/otw/demo-api", h.demoAPI)   // simulated token-gated external API
	return nil
}

// panelData is what the fragment template renders.
type panelData struct {
	Title string
	Rows  []row
	Note  string
}

type row struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// panel renders the section as an HTML fragment (no page layout): the client
// drops it straight into a container with innerHTML.
func (h *handler) panel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, err := h.fetch(r.Context())
	if err != nil {
		// The BFF absorbs the upstream failure: log the real error, show a
		// friendly fragment (the client just injects whatever HTML we return).
		log.Printf("otw panel: %v", err)
		render(w, panelData{
			Title: "Panel externo",
			Rows:  []row{{Label: "Estado", Value: "no disponible"}},
			Note:  "No pudimos contactar el servicio externo.",
		})
		return
	}
	render(w, data)
}

func render(w http.ResponseWriter, d panelData) {
	if err := tplPanel.ExecuteTemplate(w, "panel.html", d); err != nil {
		log.Printf("otw render: %v", err)
	}
}

// fetch calls the (real or simulated) GitHub API server-side with the bearer
// token, decodes the repo JSON and maps it to the fragment's rows. The token
// stays here; the client never sees it. The decode shape is a subset of GitHub's
// real response, so the same code works against the live API or the simulation.
func (h *handler) fetch(ctx context.Context) (panelData, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, h.p.apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+h.p.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := h.p.client.Do(req)
	if err != nil {
		return panelData{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return panelData{}, fmt.Errorf("github API: status %d", resp.StatusCode)
	}

	var repo struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Stars       int    `json:"stargazers_count"`
		Forks       int    `json:"forks_count"`
		OpenIssues  int    `json:"open_issues_count"`
		Language    string `json:"language"`
		License     struct {
			SPDXID string `json:"spdx_id"`
		} `json:"license"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return panelData{}, err
	}

	note := "Datos en vivo de la API de GitHub; el servidor la llamó con un token (PAT) que nunca llega al navegador."
	if h.p.demo {
		note = "API de GitHub SIMULADA (demo): el servidor la llamaría con un token —para más rate-limit o repos privados— que nunca llega al navegador. Pon OTW_API_URL=https://api.github.com/repos/<owner>/<repo> y un PAT para datos reales."
	}
	return panelData{
		Title: repo.FullName,
		Rows: []row{
			{Label: "Descripción", Value: repo.Description},
			{Label: "Estrellas", Value: strconv.Itoa(repo.Stars)},
			{Label: "Forks", Value: strconv.Itoa(repo.Forks)},
			{Label: "Issues abiertas", Value: strconv.Itoa(repo.OpenIssues)},
			{Label: "Lenguaje", Value: repo.Language},
			{Label: "Licencia", Value: repo.License.SPDXID},
		},
		Note: note,
	}, nil
}

// demoAPI is a stand-in for the real GitHub API: it requires the bearer token
// and returns the SAME JSON shape GitHub does. It exists only so the demo runs
// without external setup or a real token.
func (h *handler) demoAPI(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer "+demoToken {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{` +
		`"full_name":"dragones-tech/agogo",` +
		`"description":"Servidor web compacto en Go puro: HTML para SEO + API JSON.",` +
		`"stargazers_count":128,"forks_count":9,"open_issues_count":3,` +
		`"language":"Go","license":{"spdx_id":"MIT"}}`))
}
