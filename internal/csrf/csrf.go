// Package csrf implements CSRF protection via the "double submit cookie"
// pattern: the server issues a token, stores it in a cookie and embeds it in a
// hidden form field; on receiving the POST it compares the two. Our own code,
// stdlib, no dependencies.
package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const cookieName = "csrf_token"

// FieldName is the name of the hidden field the form must carry.
const FieldName = "csrf_token"

// Issue generates a token, stores it in a cookie and returns it for the <input
// hidden>. secure marks the cookie as Secure (only sent over HTTPS); pass it
// from the configuration (Config.Secure).
func Issue(w http.ResponseWriter, secure bool) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand shouldn't fail; if it does, we don't issue a usable token.
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

// Valid compares the cookie's token with the form's (constant-time comparison).
// It fails if either is missing or they don't match.
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
