// Package contacto serves a form: Mostrar (GET) and Recibir (POST). Uses the
// session for the success flash message (Post-Redirect-Get).
package contacto

import (
	"embed"
	"net/http"
	"strings"

	"agogo/internal/contacto/db"
	"agogo/internal/csrf"
	"agogo/internal/session"
	"agogo/internal/validate"
	"agogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

var tplForm = view.Layout(tplFS, "templates/form.html")

type Handler struct {
	q       *db.Queries
	baseURL string
	sess    *session.Manager
	secure  bool // marks the CSRF cookie as Secure
}

func New(q *db.Queries, baseURL string, sess *session.Manager, secure bool) *Handler {
	return &Handler{q: q, baseURL: baseURL, sess: sess, secure: secure}
}

type formPage struct {
	Meta   view.Meta
	Token  string
	Flash  string
	Errors map[string]string
	Values map[string]string
}

func (h *Handler) meta(r *http.Request) view.Meta {
	return view.Meta{
		Title:       "Contacto — Agogo",
		Description: "Escríbenos un mensaje.",
		Canonical:   h.baseURL + r.URL.Path,
		OGType:      "website",
	}
}

// Mostrar (GET): shows the form. If there's a flash in the session, it reads
// and consumes it (get-and-clear).
func (h *Handler) Mostrar(w http.ResponseWriter, r *http.Request) {
	s := h.sess.Get(r)
	flash := s.Get("flash")
	if flash != "" {
		s.Delete("flash")
		h.sess.Save(w, s)
	}
	view.Render(w, r, tplForm, formPage{
		Meta:   h.meta(r),
		Token:  csrf.Issue(w, h.secure),
		Flash:  flash,
		Errors: map[string]string{},
		Values: map[string]string{},
	})
}

// Recibir (POST): validates CSRF, validates fields, saves, leaves a flash in
// the session and redirects (PRG).
func (h *Handler) Recibir(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario inválido", http.StatusBadRequest)
		return
	}
	if !csrf.Valid(r) {
		http.Error(w, "token CSRF inválido", http.StatusForbidden)
		return
	}

	nombre := strings.TrimSpace(r.PostFormValue("nombre"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	mensaje := strings.TrimSpace(r.PostFormValue("mensaje"))

	// Same rules mirrored by the client in static/js/contact/index.js (UX only;
	// this is the authority).
	errs := validate.New()
	errs.Required("nombre", nombre, "El nombre es obligatorio.")
	errs.Required("email", email, "El email es obligatorio.")
	errs.Email("email", email, "Email inválido.")
	errs.MinLen("mensaje", mensaje, 10, "El mensaje es muy corto (mínimo 10 caracteres).")

	if !errs.OK() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		view.Render(w, r, tplForm, formPage{
			Meta:   h.meta(r),
			Token:  csrf.Issue(w, h.secure),
			Errors: errs,
			Values: map[string]string{"nombre": nombre, "email": email, "mensaje": mensaje},
		})
		return
	}

	if err := h.q.CreateContacto(r.Context(), db.CreateContactoParams{
		Nombre:  nombre,
		Email:   email,
		Mensaje: mensaje,
	}); err != nil {
		view.ServerError(w, r, err)
		return
	}

	s := h.sess.Get(r)
	s.Set("flash", "¡Gracias! Tu mensaje fue enviado.")
	h.sess.Save(w, s)
	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}
