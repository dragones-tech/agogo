// Package router wraps http.ServeMux with an expressive Express-style API:
//
//	r.Get("/productos/{slug}", fn)
//	r.Post("/contacto", fn)
//	r.Get("/cuenta", fn, authsvc.Require)   // with per-route middleware
//
// It's our own code on top of the stdlib: each method prepends the HTTP verb to
// the ServeMux pattern and applies the given middleware (the first being the
// outermost). The handler follows Go's idiomatic signature: func(w, r).
package router

import "net/http"

// Middleware wraps a handler to add behavior (auth, etc.).
type Middleware = func(http.HandlerFunc) http.HandlerFunc

type Router struct {
	mux *http.ServeMux
}

func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Handler returns the http.Handler to pass to the http.Server.
func (r *Router) Handler() http.Handler { return r.mux }

func (r *Router) Get(path string, fn http.HandlerFunc, mw ...Middleware) {
	r.mux.HandleFunc("GET "+path, chain(fn, mw))
}
func (r *Router) Post(path string, fn http.HandlerFunc, mw ...Middleware) {
	r.mux.HandleFunc("POST "+path, chain(fn, mw))
}
func (r *Router) Put(path string, fn http.HandlerFunc, mw ...Middleware) {
	r.mux.HandleFunc("PUT "+path, chain(fn, mw))
}
func (r *Router) Delete(path string, fn http.HandlerFunc, mw ...Middleware) {
	r.mux.HandleFunc("DELETE "+path, chain(fn, mw))
}

// Handle mounts an http.Handler on a full pattern (verb included),
// e.g. a file server at "GET /static/".
func (r *Router) Handle(pattern string, h http.Handler) { r.mux.Handle(pattern, h) }

// chain applies the middleware in order: the first in the list becomes the
// outermost layer (runs first).
func chain(fn http.HandlerFunc, mw []Middleware) http.HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		fn = mw[i](fn)
	}
	return fn
}
