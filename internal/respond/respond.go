// Package respond escribe respuestas HTTP comunes, compartidas por los dominios.
// Son funciones planas: lo que ves es lo que pasa, sin estado ni reflexión.
package respond

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON escribe v como JSON con el status dado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error escribe un cuerpo {"error": msg} con el status dado.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// ServerError registra el error REAL (con método y ruta, para depurar) y
// responde un 500 genérico al cliente, sin filtrar detalles internos. Úsalo en
// handlers JSON cuando una operación falla por culpa del servidor.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error en %s %s: %v", r.Method, r.URL.Path, err)
	Error(w, http.StatusInternalServerError, "error interno")
}
