package validate

import "testing"

func TestRequired(t *testing.T) {
	e := New()
	e.Required("nombre", "  ", "obligatorio") // solo espacios → falla
	e.Required("ciudad", "Oaxaca", "obligatorio")
	if e["nombre"] != "obligatorio" {
		t.Errorf("nombre vacío debería fallar, got %q", e["nombre"])
	}
	if _, ok := e["ciudad"]; ok {
		t.Error("ciudad con valor no debería fallar")
	}
}

func TestMinLen(t *testing.T) {
	e := New()
	e.MinLen("msg", "corto", 10, "muy corto")
	e.MinLen("ok", "suficientemente largo", 10, "muy corto")
	// café = 4 runas aunque 5 bytes: cuenta runas
	e.MinLen("acentos", "café", 4, "muy corto")
	if e["msg"] == "" {
		t.Error("'corto' (<10) debería fallar")
	}
	if e["ok"] != "" {
		t.Error("una cadena larga no debería fallar")
	}
	if e["acentos"] != "" {
		t.Error("'café' tiene 4 runas, no debería fallar con min=4")
	}
}

func TestEmail(t *testing.T) {
	e := New()
	e.Email("bad", "no-es-email", "inválido")
	e.Email("good", "user@example.com", "inválido")
	if e["bad"] == "" {
		t.Error("un email inválido debería fallar")
	}
	if e["good"] != "" {
		t.Error("un email válido no debería fallar")
	}
}

func TestAddKeepsFirst(t *testing.T) {
	e := New()
	e.Required("email", "", "obligatorio")
	e.Email("email", "", "inválido") // también falla, pero Add conserva el primero
	if e["email"] != "obligatorio" {
		t.Errorf("Add debería conservar el primer mensaje, got %q", e["email"])
	}
	if e.OK() {
		t.Error("OK() debería ser false con errores")
	}
}
