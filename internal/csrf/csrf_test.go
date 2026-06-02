package csrf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIssueYValid(t *testing.T) {
	w := httptest.NewRecorder()
	token := Issue(w, false)
	if token == "" {
		t.Fatal("Issue devolvió token vacío")
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != cookieName || cookies[0].Value != token {
		t.Fatalf("cookie CSRF mal emitida: %+v", cookies)
	}

	// POST con la cookie y el campo coincidentes → válido.
	form := url.Values{FieldName: {token}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	if !Valid(req) {
		t.Error("Valid debería aceptar cookie y campo coincidentes")
	}
}

func TestIssueSecure(t *testing.T) {
	w := httptest.NewRecorder()
	Issue(w, true)
	if !w.Result().Cookies()[0].Secure {
		t.Error("con secure=true la cookie CSRF debería marcarse Secure")
	}
}

func TestValidRechaza(t *testing.T) {
	mkReq := func(cookieVal, field string, conCookie bool) *http.Request {
		form := url.Values{FieldName: {field}}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if conCookie {
			req.AddCookie(&http.Cookie{Name: cookieName, Value: cookieVal})
		}
		return req
	}

	if Valid(mkReq("", "", false)) {
		t.Error("sin cookie ni campo no debería validar")
	}
	if Valid(mkReq("abc", "xyz", true)) {
		t.Error("cookie y campo distintos no deberían validar")
	}
	if Valid(mkReq("abc", "", true)) {
		t.Error("campo vacío no debería validar")
	}
}
