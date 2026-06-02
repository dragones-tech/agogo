// Package contacto sirve un formulario: Mostrar (GET) y Recibir (POST). Usa
// sesión para el mensaje flash de éxito (Post-Redirect-Get).
package contacto

import (
	"embed"
	"net/http"
	"strings"

	"agogo/internal/contacto/db"
	"agogo/internal/csrf"
	"agogo/internal/session"
	"agogo/internal/view"
)

//go:embed templates/*.html
var tplFS embed.FS

var tplForm = view.Layout(tplFS, "templates/form.html")

type Handler struct {
	q       *db.Queries
	baseURL string
	sess    *session.Manager
	secure  bool // marca la cookie CSRF como Secure
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

// Mostrar (GET): muestra el formulario. Si hay un flash en la sesión, lo lee y
// lo consume (get-and-clear).
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

// Recibir (POST): valida CSRF, valida campos, guarda, deja un flash en la sesión
// y redirige (PRG).
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

	errs := map[string]string{}
	if nombre == "" {
		errs["nombre"] = "El nombre es obligatorio."
	}
	if !strings.Contains(email, "@") {
		errs["email"] = "Email inválido."
	}
	if len(mensaje) < 10 {
		errs["mensaje"] = "El mensaje es muy corto (mínimo 10 caracteres)."
	}

	if len(errs) > 0 {
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
