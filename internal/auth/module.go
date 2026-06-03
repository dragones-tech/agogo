package auth

import (
	"agogo/internal/app"
	"agogo/internal/auth/db"
)

// Module wires up authentication: login/logout and the protected /cuenta page.
// Uses the host's shared identity service (a.Identity).
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "auth" }

func (mod) Register(a *app.App) error {
	h := New(db.New(a.DB), a.Config.BaseURL, a.Identity, a.Config.Secure)

	r := a.Router
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/cuenta", h.Cuenta, a.Identity.Require) // identity as per-route middleware

	a.AddMigration(Migrate)
	return nil
}
