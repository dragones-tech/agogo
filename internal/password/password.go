// Package password hashea y verifica contraseñas con PBKDF2-HMAC-SHA256, todo
// con la stdlib (crypto/pbkdf2, disponible desde Go 1.24). Sin dependencias.
//
// Nota honesta: bcrypt/argon2 (golang.org/x/crypto) son preferidos por su dureza
// de memoria; PBKDF2 con muchas iteraciones es aprobado por NIST/OWASP y nos
// mantiene en stdlib. El formato guardado lleva los parámetros, así se puede
// migrar el algoritmo sin romper hashes viejos.
package password

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	iterations = 600_000 // OWASP 2023 para PBKDF2-HMAC-SHA256
	keyLen     = 32
	saltLen    = 16
)

// Hash devuelve "pbkdf2-sha256$iter$salt$hash" (salt y hash en base64).
func Hash(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk, err := pbkdf2.Key(sha256.New, plain, salt, iterations, keyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s",
		iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk),
	), nil
}

// Match verifica una contraseña contra un hash codificado, en tiempo constante.
func Match(plain, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, plain, salt, iter, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}
