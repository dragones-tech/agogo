// Package router envuelve http.ServeMux con una API expresiva al estilo Express:
//
//	r.Get("/productos/{slug}", fn)
//	r.Post("/contacto", fn)
//	r.Get("/cuenta", fn, authsvc.Require)   // con middleware por ruta
//
// Es código nuestro sobre la stdlib: cada método antepone el verbo HTTP al
// patrón de ServeMux y aplica los middleware dados (el primero queda más
// externo). El handler sigue la firma idiomática de Go: func(w, r).
package router

import "net/http"

// Middleware envuelve un handler para añadir comportamiento (auth, etc.).
type Middleware = func(http.HandlerFunc) http.HandlerFunc

type Router struct {
	mux *http.ServeMux
}

func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Handler devuelve el http.Handler para pasarlo al http.Server.
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

// Handle monta un http.Handler en un patrón completo (incluye el verbo),
// p. ej. un servidor de archivos en "GET /static/".
func (r *Router) Handle(pattern string, h http.Handler) { r.mux.Handle(pattern, h) }

// chain aplica los middleware en orden: el primero de la lista queda como capa
// más externa (se ejecuta primero).
func chain(fn http.HandlerFunc, mw []Middleware) http.HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		fn = mw[i](fn)
	}
	return fn
}
