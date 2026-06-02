// Package identity provides identity on top of the session: log in/out, knowing
// who the current user is, and a middleware to protect routes. It doesn't know
// the DB; it only stores the user id in the signed session.
package identity

import (
	"net/http"

	"agogo/internal/router"
	"agogo/internal/session"
)

const userKey = "user_id"

type Service struct {
	sess *session.Manager
}

func New(sess *session.Manager) *Service {
	return &Service{sess: sess}
}

// Login marks the user as authenticated by storing their id in the session.
func (s *Service) Login(w http.ResponseWriter, r *http.Request, userID string) {
	s.LoginInto(w, s.sess.Get(r), userID)
}

// LoginInto marks the user on an ALREADY loaded session and saves it in a single
// Set-Cookie. Use it when the caller made other mutations that must persist in
// the same save (e.g. oauth, which consumes its anti-CSRF state before logging
// in); two separate Saves would emit two cookies with the same name and the
// client would keep only one.
func (s *Service) LoginInto(w http.ResponseWriter, sess *session.Session, userID string) {
	sess.Set(userKey, userID)
	s.sess.Save(w, sess)
}

// Logout clears the identity from the session.
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	sess := s.sess.Get(r)
	sess.Delete(userKey)
	s.sess.Save(w, sess)
}

// UserID returns the current user's id ("" if no session is started).
func (s *Service) UserID(r *http.Request) string {
	return s.sess.Get(r).Get(userKey)
}

// Require protects a route: if there's no session, it redirects to /login.
func (s *Service) Require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.UserID(r) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// Guard registers routes that ALWAYS require a session: it applies Require for
// you, so you don't have to remember to wrap each handler (and you don't
// accidentally leave a route public). Obtain one with Service.Group(r).
type Guard struct {
	r *router.Router
	s *Service
}

// Group returns a Guard over the router: everything you register with it ends up
// protected. Example:
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
