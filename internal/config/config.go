// Package config lee la configuración del entorno una sola vez, tipada y
// validada. Lo usan tanto el servidor como los comandos (cmd/*).
package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB        string // ruta del archivo SQLite
	BaseURL   string // URL base (canonical, sitemap)
	Addr      string // dirección de escucha del servidor
	SecretKey []byte // clave para firmar sesiones (HMAC)
	DevSecret bool   // true si se cayó a la clave de desarrollo (insegura)
}

// devSecret se usa si no se define JEHOSOGO_SECRET_KEY. Es estable entre
// reinicios (las sesiones de dev sobreviven), pero INSEGURA para producción.
const devSecret = "jehosogo-dev-secret-insegura-cambiala-en-produccion"

// Load arma la configuración desde el entorno. Devuelve error en validaciones
// duras (p. ej. clave demasiado corta).
func Load() (Config, error) {
	c := Config{
		DB:      env("JEHOSOGO_DB", "jehosogo.db"),
		BaseURL: env("JEHOSOGO_BASE_URL", "http://localhost:8080"),
		Addr:    env("JEHOSOGO_ADDR", ":8080"),
	}

	switch secret := os.Getenv("JEHOSOGO_SECRET_KEY"); {
	case secret == "":
		c.SecretKey = []byte(devSecret)
		c.DevSecret = true
	case len(secret) < 32:
		return c, fmt.Errorf("JEHOSOGO_SECRET_KEY debe tener al menos 32 caracteres (tiene %d)", len(secret))
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
