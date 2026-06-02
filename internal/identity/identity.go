// Package identity da identidad sobre la sesión: iniciar/cerrar sesión, saber quién
// es el usuario actual, y un middleware para proteger rutas. No conoce la BD;
// solo guarda el id de usuario en la sesión firmada.
package identity

import (
	"net/http"

	"jehosogo/internal/router"
	"jehosogo/internal/session"
)

const userKey = "user_id"

type Service struct {
	sess *session.Manager
}

func New(sess *session.Manager) *Service {
	return &Service{sess: sess}
}

// Login marca al usuario como autenticado guardando su id en la sesión.
func (s *Service) Login(w http.ResponseWriter, r *http.Request, userID string) {
	sess := s.sess.Get(r)
	sess.Set(userKey, userID)
	s.sess.Save(w, sess)
}

// Logout borra la identidad de la sesión.
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	sess := s.sess.Get(r)
	sess.Delete(userKey)
	s.sess.Save(w, sess)
}

// UserID devuelve el id del usuario actual ("" si no hay sesión iniciada).
func (s *Service) UserID(r *http.Request) string {
	return s.sess.Get(r).Get(userKey)
}

// Require protege una ruta: si no hay sesión, redirige a /login.
func (s *Service) Require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.UserID(r) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// Guard registra rutas que SIEMPRE exigen sesión: aplica Require por ti, así no
// tienes que recordar envolver cada handler (y no se te cuela una ruta pública
// por olvido). Se obtiene con Service.Group(r).
type Guard struct {
	r *router.Router
	s *Service
}

// Group devuelve un Guard sobre el router: todo lo que registres con él queda
// protegido. Ejemplo:
//
//	protected := authsvc.Group(r)
//	protected.Get("/cuenta", h.Cuenta)
func (s *Service) Group(r *router.Router) *Guard {
	return &Guard{r: r, s: s}
}

func (g *Guard) Get(path string, fn http.HandlerFunc)    { g.r.Get(path, g.s.Require(fn)) }
func (g *Guard) Post(path string, fn http.HandlerFunc)   { g.r.Post(path, g.s.Require(fn)) }
func (g *Guard) Put(path string, fn http.HandlerFunc)    { g.r.Put(path, g.s.Require(fn)) }
func (g *Guard) Delete(path string, fn http.HandlerFunc) { g.r.Delete(path, g.s.Require(fn)) }
