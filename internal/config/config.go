// Package config lee la configuración del entorno una sola vez, tipada y
// validada. Lo usan tanto el servidor como los comandos (cmd/*).
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DB        string // ruta del archivo SQLite
	BaseURL   string // URL base (canonical, sitemap)
	Addr      string // dirección de escucha del servidor
	SecretKey []byte // clave para firmar sesiones (HMAC)
	DevSecret bool   // true si se cayó a la clave de desarrollo (insegura)
	Secure    bool   // marca las cookies como Secure (solo viajan por HTTPS)
}

// devSecret se usa si no se define AGOGO_SECRET_KEY. Es estable entre
// reinicios (las sesiones de dev sobreviven), pero INSEGURA para producción.
const devSecret = "agogo-dev-secret-insegura-cambiala-en-produccion"

// Load arma la configuración desde el entorno. Devuelve error en validaciones
// duras (p. ej. clave demasiado corta).
func Load() (Config, error) {
	c := Config{
		DB:      env("AGOGO_DB", "agogo.db"),
		BaseURL: env("AGOGO_BASE_URL", "http://localhost:8080"),
		Addr:    env("AGOGO_ADDR", ":8080"),
	}

	// Cookies Secure: por defecto se deduce del esquema de la URL base (https →
	// true); AGOGO_SECURE_COOKIES lo fuerza explícitamente (útil tras un proxy
	// TLS que termina en http://).
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

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
