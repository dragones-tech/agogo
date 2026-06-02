package openapi

import "agogo/internal/app"

// Module wires up the API documentation: the spec and the Swagger UI.
func Module() app.Module { return mod{} }

type mod struct{}

func (mod) Name() string { return "openapi" }

func (mod) Register(a *app.App) error {
	r := a.Router
	r.Get("/openapi.json", Spec)
	r.Get("/docs", Docs)
	r.Handle("GET /docs-assets/", Assets())
	return nil
}
