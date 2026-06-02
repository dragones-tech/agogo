// Package auth handles authentication: login, logout and a protected page.
// Uses internal/auth (identity over session) and internal/password (PBKDF2).
package auth

import (
	"database/sql"
	"embed"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"agogo/internal/auth/db"
	"agogo/internal/csrf"
	"agogo/internal/identity"
	"agogo/internal/password"
	"agogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

var (
	tplLogin  = view.Layout(tplFS, "templates/login.html")
	tplCuenta = view.Layout(tplFS, "templates/cuenta.html")
)

type Handler struct {
	q        *db.Queries
	baseURL  string
	identity *identity.Service
	secure   bool // marks the (CSRF) cookies as Secure
}

func New(q *db.Queries, baseURL string, a *identity.Service, secure bool) *Handler {
	return &Handler{q: q, baseURL: baseURL, identity: a, secure: secure}
}

func (h *Handler) meta(r *http.Request, title string) view.Meta {
	return view.Meta{Title: title, Canonical: h.baseURL + r.URL.Path, OGType: "website"}
}

type loginPage struct {
	Meta  view.Meta
	Token string
	Error string
	Email string
}

// GET /login
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	view.Render(w, r, tplLogin, loginPage{Meta: h.meta(r, "Acceder — Agogo"), Token: csrf.Issue(w, h.secure)})
}

// POST /login: validates CSRF, verifies credentials, starts the session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario inválido", http.StatusBadRequest)
		return
	}
	if !csrf.Valid(r) {
		http.Error(w, "token CSRF inválido", http.StatusForbidden)
		return
	}
	email := strings.TrimSpace(r.PostFormValue("email"))
	pass := r.PostFormValue("password")

	u, err := h.q.GetUsuarioByEmail(r.Context(), email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		view.ServerError(w, r, err)
		return
	}
	// Don't reveal whether the email exists: same message for "not found" and "bad password".
	if errors.Is(err, sql.ErrNoRows) || !password.Match(pass, u.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		view.Render(w, r, tplLogin, loginPage{
			Meta:  h.meta(r, "Acceder — Agogo"),
			Token: csrf.Issue(w, h.secure),
			Error: "Correo o contraseña incorrectos.",
			Email: email,
		})
		return
	}

	h.identity.Login(w, r, strconv.FormatInt(u.ID, 10))
	http.Redirect(w, r, "/cuenta", http.StatusSeeOther)
}

// POST /logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario inválido", http.StatusBadRequest)
		return
	}
	if !csrf.Valid(r) {
		http.Error(w, "token CSRF inválido", http.StatusForbidden)
		return
	}
	h.identity.Logout(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type cuentaPage struct {
	Meta  view.Meta
	Token string
	Email string
}

// GET /cuenta (protected by auth.Require): shows the current user.
func (h *Handler) Cuenta(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(h.identity.UserID(r), 10, 64)
	u, err := h.q.GetUsuario(r.Context(), id)
	if err != nil {
		view.ServerError(w, r, err)
		return
	}
	// Authenticated page: don't let any proxy or the browser cache it.
	w.Header().Set("Cache-Control", "no-store")
	view.Render(w, r, tplCuenta, cuentaPage{Meta: h.meta(r, "Mi cuenta — Agogo"), Token: csrf.Issue(w, h.secure), Email: u.Email})
}
