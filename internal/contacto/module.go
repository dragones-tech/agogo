package contacto

import (
	"agogo/internal/app"
	"agogo/internal/contacto/db"
	"agogo/internal/sitemap"
)

// Module wires up the contact form: routes, migration and sitemap. Uses the
// host's shared session service (a.Session) for the flash.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "contacto" }

func (mod) Register(a *app.App) error {
	h := New(db.New(a.DB), a.Config.BaseURL, a.Session, a.Config.Secure)

	r := a.Router
	r.Get("/contacto", h.Mostrar)
	r.Post("/contacto", h.Recibir)

	a.AddMigration(Migrate)
	a.AddSitemap(sitemap.StaticURLs("/contacto"))
	return nil
}
