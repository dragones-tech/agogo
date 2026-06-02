// Package config reads the configuration from the environment once, typed and
// validated. Both the server and the commands (cmd/*) use it.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DB        string // path to the SQLite file
	BaseURL   string // base URL (canonical, sitemap)
	Addr      string // server listen address
	SecretKey []byte // key for signing sessions (HMAC)
	DevSecret bool   // true if it fell back to the (insecure) development key
	Secure    bool   // marks cookies as Secure (only sent over HTTPS)
}

// devSecret is used if AGOGO_SECRET_KEY isn't defined. It's stable across
// restarts (dev sessions survive), but INSECURE for production.
const devSecret = "agogo-dev-secret-insegura-cambiala-en-produccion"

// Load builds the configuration from the environment. It returns an error on
// hard validations (e.g. a key that's too short).
func Load() (Config, error) {
	c := Config{
		DB:      env("AGOGO_DB", "agogo.db"),
		BaseURL: env("AGOGO_BASE_URL", "http://localhost:46060"),
		Addr:    env("AGOGO_ADDR", ":46060"),
	}

	// Secure cookies: by default deduced from the base URL's scheme (https →
	// true); AGOGO_SECURE_COOKIES forces it explicitly (useful behind a TLS
	// proxy that terminates to http://).
	c.Secure = strings.HasPrefix(c.BaseURL, "https://")
	switch strings.ToLower(os.Getenv("AGOGO_SECURE_COOKIES")) {
	case "1", "true", "yes":
		c.Secure = true
	case "0", "false", "no":
		c.Secure = false
	}

	switch secret := os.Getenv("AGOGO_SECRET_KEY"); {
	case secret == "":
		c.SecretKey = []byte(devSecret)
		c.DevSecret = true
	case len(secret) < 32:
		return c, fmt.Errorf("AGOGO_SECRET_KEY debe tener al menos 32 caracteres (tiene %d)", len(secret))
	default:
		c.SecretKey = []byte(secret)
	}

	return c, nil
}

// DSN is the SQLite connection string: the path + PRAGMAs that the modernc
// driver applies when opening each connection. WAL allows concurrent reads with
// one write; busy_timeout makes a write wait (up to 5s) instead of failing
// instantly with "database is locked"; foreign_keys enables FKs.
func (c Config) DSN() string {
	return c.DB + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)"
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
