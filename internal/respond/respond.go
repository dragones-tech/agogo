// Package respond escribe respuestas HTTP comunes, compartidas por los dominios.
// Son funciones planas: lo que ves es lo que pasa, sin estado ni reflexión.
package respond

import (
	"encoding/json"
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
