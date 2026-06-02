// Package csrf implementa protección CSRF por "doble envío de cookie": el
// servidor emite un token, lo guarda en una cookie y lo incrusta en un campo
// oculto del formulario; al recibir el POST compara ambos. Código nuestro,
// stdlib, sin dependencias.
package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const cookieName = "csrf_token"

// FieldName es el nombre del campo oculto que debe llevar el formulario.
const FieldName = "csrf_token"

// Issue genera un token, lo guarda en cookie y lo devuelve para el <input hidden>.
// secure marca la cookie como Secure (solo viaja por HTTPS); pásalo desde la
// configuración (Config.Secure).
func Issue(w http.ResponseWriter, secure bool) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand no debería fallar; si lo hace, no emitimos token usable.
		http.Error(w, "error interno", http.StatusInternalServerError)
		return ""
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// Valid compara el token de la cookie con el del formulario (comparación
// en tiempo constante). Falla si falta cualquiera o no coinciden.
func Valid(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return false
	}
	field := r.PostFormValue(FieldName)
	if field == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(field)) == 1
}
