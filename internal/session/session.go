// Package session implements sessions via a SIGNED COOKIE (HMAC-SHA256), with
// no server-side store and no dependencies. Our own code on top of the stdlib,
// in the same vein as internal/csrf.
//
// The cookie is SIGNED (tamper-proof) but NOT encrypted: the data is readable by
// the client. Don't store secrets in the session; do store things like the user
// id or a flash message. Practical limit: ~4 KB per cookie.
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

// expKey is the reserved key where the expiration instant (unix) is stored. The
// server injects it on save and verifies/hides it on read, so a stolen cookie
// stops being valid past maxAge even if the client keeps the file.
const expKey = "_exp"

// maxAge is the lifetime of a session. It's applied in two places: the cookie's
// MaxAge attribute (controlled by the client) and the signed expKey stamp
// (enforced by the server, unforgeable).
const maxAge = 7 * 24 * time.Hour

// now is the clock; a variable so it can be pinned in tests.
var now = time.Now

// Manager signs and verifies sessions with a secret key.
type Manager struct {
	key    []byte
	Secure bool // set to true when serving behind TLS
}

func NewManager(key []byte) *Manager {
	return &Manager{key: key}
}

// Session holds the values of the current session.
type Session struct {
	Values map[string]string
}

func (s *Session) Get(key string) string { return s.Values[key] }
func (s *Session) Set(key, val string)   { s.Values[key] = val }
func (s *Session) Delete(key string)     { delete(s.Values, key) }

// Get reads the session from the cookie (empty if absent, the signature fails to
// validate, or it expired).
func (m *Manager) Get(r *http.Request) *Session {
	s := &Session{Values: map[string]string{}}
	if c, err := r.Cookie(cookieName); err == nil {
		if values, ok := m.decode(c.Value); ok && !expired(values) {
			delete(values, expKey) // internal stamp: not exposed to handlers
			s.Values = values
		}
	}
	return s
}

// Save writes the signed session into a cookie, stamping it with its expiration.
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

// expired reports whether the expKey stamp has already passed. No stamp →
// expired (old cookie predating this scheme, or tampered to remove it).
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
		return nil, false // invalid signature → tampered or different key
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
