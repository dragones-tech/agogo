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
	"strconv"
	"strings"
	"time"
)

const cookieName = "agogo_session"

// expKey es la clave reservada donde se guarda el instante de expiración (unix).
// El servidor la inyecta al guardar y la verifica/oculta al leer, así una cookie
// robada deja de valer pasado maxAge aunque el cliente conserve el archivo.
const expKey = "_exp"

// maxAge es la vida de una sesión. Se aplica en dos sitios: el atributo MaxAge de
// la cookie (lo controla el cliente) y el sello firmado expKey (lo impone el
// servidor, infalsificable).
const maxAge = 7 * 24 * time.Hour

// now es el reloj; variable para poder fijarlo en los tests.
var now = time.Now

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

func (s *Session) Get(key string) string { return s.Values[key] }
func (s *Session) Set(key, val string)   { s.Values[key] = val }
func (s *Session) Delete(key string)     { delete(s.Values, key) }

// Get lee la sesión de la cookie (vacía si no hay, la firma no valida o expiró).
func (m *Manager) Get(r *http.Request) *Session {
	s := &Session{Values: map[string]string{}}
	if c, err := r.Cookie(cookieName); err == nil {
		if values, ok := m.decode(c.Value); ok && !expired(values) {
			delete(values, expKey) // sello interno: no se expone a los handlers
			s.Values = values
		}
	}
	return s
}

// Save escribe la sesión firmada en una cookie, sellándola con su expiración.
func (m *Manager) Save(w http.ResponseWriter, s *Session) {
	s.Values[expKey] = strconv.FormatInt(now().Add(maxAge).Unix(), 10)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    m.encode(s.Values),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
	})
}

// expired indica si el sello expKey ya pasó. Sin sello → expirada (cookie vieja
// previa a este esquema, o manipulada para quitarlo).
func expired(values map[string]string) bool {
	exp, err := strconv.ParseInt(values[expKey], 10, 64)
	if err != nil {
		return true
	}
	return now().Unix() >= exp
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
