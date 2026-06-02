// Package auth maneja autenticación: login, logout y una página protegida.
// Usa internal/auth (identidad sobre sesión) y internal/password (PBKDF2).
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
}

func New(q *db.Queries, baseURL string, a *identity.Service) *Handler {
	return &Handler{q: q, baseURL: baseURL, identity: a}
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
	view.Render(w, tplLogin, loginPage{Meta: h.meta(r, "Acceder — Agogo"), Token: csrf.Issue(w)})
}

// POST /login: valida CSRF, verifica credenciales, inicia sesión.
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
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	// No revelamos si el correo existe: mismo mensaje para "no existe" y "mala contraseña".
	if errors.Is(err, sql.ErrNoRows) || !password.Match(pass, u.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		view.Render(w, tplLogin, loginPage{
			Meta:  h.meta(r, "Acceder — Agogo"),
			Token: csrf.Issue(w),
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

// GET /cuenta (protegida por auth.Require): muestra el usuario actual.
func (h *Handler) Cuenta(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(h.identity.UserID(r), 10, 64)
	u, err := h.q.GetUsuario(r.Context(), id)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	view.Render(w, tplCuenta, cuentaPage{Meta: h.meta(r, "Mi cuenta — Agogo"), Token: csrf.Issue(w), Email: u.Email})
}
