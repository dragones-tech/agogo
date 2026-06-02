package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// cookieDeRespuesta extrae la cookie de sesión escrita en w y arma un request
// que la lleva, simulando el viaje de ida y vuelta navegador↔servidor.
func cookieDeRespuesta(t *testing.T, w *httptest.ResponseRecorder) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	return req
}

func TestSaveGetRoundTrip(t *testing.T) {
	m := NewManager([]byte("0123456789abcdef0123456789abcdef"))

	s := &Session{Values: map[string]string{}}
	s.Set("user_id", "42")
	w := httptest.NewRecorder()
	m.Save(w, s)

	got := m.Get(cookieDeRespuesta(t, w))
	if got.Get("user_id") != "42" {
		t.Errorf("user_id = %q, quería 42", got.Get("user_id"))
	}
	// El sello interno de expiración no debe filtrarse a los handlers.
	if _, ok := got.Values[expKey]; ok {
		t.Errorf("la sesión leída no debería exponer %q", expKey)
	}
}

func TestFirmaInvalida(t *testing.T) {
	m := NewManager([]byte("0123456789abcdef0123456789abcdef"))
	s := &Session{Values: map[string]string{"user_id": "42"}}
	w := httptest.NewRecorder()
	m.Save(w, s)

	// Otra clave → la firma no valida → sesión vacía (no se confía en ella).
	otro := NewManager([]byte("ffffffffffffffffffffffffffffffff"))
	got := otro.Get(cookieDeRespuesta(t, w))
	if got.Get("user_id") != "" {
		t.Error("una cookie firmada con otra clave no debería validar")
	}
}

func TestManipulacionDelPayload(t *testing.T) {
	m := NewManager([]byte("0123456789abcdef0123456789abcdef"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "eyJ1c2VyX2lkIjoiOTk5In0.firma-falsa"})
	if got := m.Get(req).Get("user_id"); got != "" {
		t.Errorf("payload manipulado debería rechazarse, got %q", got)
	}
}

func TestSesionExpirada(t *testing.T) {
	m := NewManager([]byte("0123456789abcdef0123456789abcdef"))

	// Guardamos "en el pasado".
	now = func() time.Time { return time.Unix(1_000_000, 0) }
	w := httptest.NewRecorder()
	m.Save(w, &Session{Values: map[string]string{"user_id": "42"}})

	// Leemos pasado el maxAge: debe tratarse como vacía.
	now = func() time.Time { return time.Unix(1_000_000, 0).Add(maxAge + time.Second) }
	defer func() { now = time.Now }()
	if got := m.Get(cookieDeRespuesta(t, w)).Get("user_id"); got != "" {
		t.Errorf("sesión expirada debería leerse vacía, got %q", got)
	}
}
