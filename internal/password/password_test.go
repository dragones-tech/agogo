package password

import "testing"

func TestHashMatch(t *testing.T) {
	h, err := Hash("correcto-caballo")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !Match("correcto-caballo", h) {
		t.Error("Match debería aceptar la contraseña correcta")
	}
	if Match("incorrecto", h) {
		t.Error("Match no debería aceptar una contraseña distinta")
	}
}

func TestHashEsAleatorio(t *testing.T) {
	// Random salt: the same text produces different hashes, both valid.
	a, _ := Hash("misma")
	b, _ := Hash("misma")
	if a == b {
		t.Error("dos hashes del mismo texto no deberían coincidir (salt aleatorio)")
	}
	if !Match("misma", a) || !Match("misma", b) {
		t.Error("ambos hashes deberían validar")
	}
}

func TestMatchFormatoInvalido(t *testing.T) {
	casos := []string{
		"",
		"texto-plano",
		"bcrypt$10$x$y",                   // unknown algorithm
		"pbkdf2-sha256$abc$c2FsdA$aGFzaA", // non-numeric iterations
		"pbkdf2-sha256$1000$@@@$aGFzaA",   // invalid base64 salt
	}
	for _, c := range casos {
		if Match("lo-que-sea", c) {
			t.Errorf("Match(%q) debería ser false", c)
		}
	}
}
