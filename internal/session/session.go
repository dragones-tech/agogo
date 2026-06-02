// Package session implementa sesiones por COOKIE FIRMADA (HMAC-SHA256), sin
// store en servidor y sin dependencias. Código nuestro sobre la stdlib, en la
// misma línea que internal/csrf.
//
// La cookie está FIRMADA (a prueba de manipulación) pero NO cifrada: los datos
// son legibles por el cliente. No guardes secretos en la sesión; sí cosas como
// el id de usuario o un mensaje flash. Límite práctico: ~4 KB por cookie.
package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

const cookieName = "jehosogo_session"

// Manager firma y verifica sesiones con una clave secreta.
type Manager struct {
	key    []byte
	Secure bool // poner en true al servir tras TLS
}

func NewManager(key []byte) *Manager {
	return &Manager{key: key}
}

// Session son los valores de la sesión actual.
type Session struct {
	Values map[string]string
}

func (s *Session) Get(key string) string  { return s.Values[key] }
func (s *Session) Set(key, val string)     { s.Values[key] = val }
func (s *Session) Delete(key string)       { delete(s.Values, key) }

// Get lee la sesión de la cookie (vacía si no hay o la firma no valida).
func (m *Manager) Get(r *http.Request) *Session {
	s := &Session{Values: map[string]string{}}
	if c, err := r.Cookie(cookieName); err == nil {
		if values, ok := m.decode(c.Value); ok {
			s.Values = values
		}
	}
	return s
}

// Save escribe la sesión firmada en una cookie.
func (m *Manager) Save(w http.ResponseWriter, s *Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    m.encode(s.Values),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Manager) encode(values map[string]string) string {
	payload, _ := json.Marshal(values)
	body := base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + m.sign(body)
}

func (m *Manager) decode(value string) (map[string]string, bool) {
	i := strings.LastIndex(value, ".")
	if i < 0 {
		return nil, false
	}
	body, sig := value[:i], value[i+1:]
	if !hmac.Equal([]byte(sig), []byte(m.sign(body))) {
		return nil, false // firma inválida → manipulada o clave distinta
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil, false
	}
	var values map[string]string
	if err := json.Unmarshal(payload, &values); err != nil {
		return nil, false
	}
	return values, true
}

func (m *Manager) sign(body string) string {
	mac := hmac.New(sha256.New, m.key)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
