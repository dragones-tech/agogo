package auth

import (
	"jehosogo/internal/app"
	"jehosogo/internal/auth/db"
)

// Module acopla la autenticación: login/logout y la página protegida /cuenta.
// Usa el servicio de identidad compartido del host (a.Identity).
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "auth" }

func (mod) Register(a *app.App) error {
	h := New(db.New(a.DB), a.Config.BaseURL, a.Identity)

	r := a.Router
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/cuenta", h.Cuenta, a.Identity.Require) // identidad como middleware por ruta

	a.AddMigration(Migrate)
	return nil
}
