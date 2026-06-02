// Package password hashes and verifies passwords with PBKDF2-HMAC-SHA256, all
// with the stdlib (crypto/pbkdf2, available since Go 1.24). No dependencies.
//
// Honest note: bcrypt/argon2 (golang.org/x/crypto) are preferred for their
// memory hardness; PBKDF2 with many iterations is approved by NIST/OWASP and
// keeps us in the stdlib. The stored format carries the parameters, so the
// algorithm can be migrated without breaking old hashes.
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
	iterations = 600_000 // OWASP 2023 for PBKDF2-HMAC-SHA256
	keyLen     = 32
	saltLen    = 16
)

// Hash returns "pbkdf2-sha256$iter$salt$hash" (salt and hash in base64).
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

// Match verifies a password against an encoded hash, in constant time.
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
