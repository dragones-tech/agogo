package contacto

import (
	"jehosogo/internal/app"
	"jehosogo/internal/contacto/db"
	"jehosogo/internal/sitemap"
)

// Module acopla el formulario de contacto: rutas, migración y sitemap. Usa el
// servicio de sesión compartido del host (a.Session) para el flash.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "contacto" }

func (mod) Register(a *app.App) error {
	h := New(db.New(a.DB), a.Config.BaseURL, a.Session)

	r := a.Router
	r.Get("/contacto", h.Mostrar)
	r.Post("/contacto", h.Recibir)

	a.AddMigration(Migrate)
	a.AddSitemap(sitemap.StaticURLs("/contacto"))
	return nil
}
