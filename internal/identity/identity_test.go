package identity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"agogo/internal/session"
)

func nuevo() *Service {
	return New(session.NewManager([]byte("0123456789abcdef0123456789abcdef")))
}

func conCookies(w *httptest.ResponseRecorder) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/cuenta", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	return req
}

func TestLoginYUserID(t *testing.T) {
	s := nuevo()
	w := httptest.NewRecorder()
	s.Login(w, httptest.NewRequest(http.MethodGet, "/", nil), "7")
	if got := s.UserID(conCookies(w)); got != "7" {
		t.Errorf("UserID = %q, quería 7", got)
	}
}

func TestLogout(t *testing.T) {
	s := nuevo()
	w := httptest.NewRecorder()
	s.Login(w, httptest.NewRequest(http.MethodGet, "/", nil), "7")

	w2 := httptest.NewRecorder()
	s.Logout(w2, conCookies(w))
	if got := s.UserID(conCookies(w2)); got != "" {
		t.Errorf("tras logout UserID debería ser vacío, got %q", got)
	}
}

func TestRequireRedirige(t *testing.T) {
	s := nuevo()
	llamado := false
	h := s.Require(func(http.ResponseWriter, *http.Request) { llamado = true })

	// Sin sesión: redirige a /login y NO ejecuta el handler.
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest(http.MethodGet, "/cuenta", nil))
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Errorf("sin sesión debería redirigir a /login, got %d %q", w.Code, w.Header().Get("Location"))
	}
	if llamado {
		t.Error("el handler protegido no debería ejecutarse sin sesión")
	}

	// Con sesión: ejecuta el handler.
	login := httptest.NewRecorder()
	s.Login(login, httptest.NewRequest(http.MethodGet, "/", nil), "7")
	h(httptest.NewRecorder(), conCookies(login))
	if !llamado {
		t.Error("con sesión el handler protegido debería ejecutarse")
	}
}
