package blog

import (
	"context"

	"jehosogo/internal/blog/db"
)

// Seed inserta artículos de ejemplo solo si la tabla está vacía. NO es parte
// del sitio: es para DESARROLLO. Se ejecuta a mano con `go run ./cmd/seed`.
func Seed(ctx context.Context, q *db.Queries) error {
	n, err := q.CountPosts(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	samples := []db.CreatePostParams{
		{
			Slug:      "por-que-go-puro",
			Titulo:    "Por qué construimos en Go puro",
			Resumen:   "La historia de elegir la stdlib sobre un framework.",
			Contenido: "Empezamos preguntándonos qué tan complicado sería un servidor seguro sin dependencias...",
			Publicado: "2026-05-20",
		},
		{
			Slug:      "sqlc-sin-orm",
			Titulo:    "sqlc: queries a medida sin ORM",
			Resumen:   "Cómo tener tipos seguros escribiendo SQL real, sin magia.",
			Contenido: "Un ORM genérico resuelve todo y por eso pesa. Nosotros queríamos solo lo que el proyecto usa...",
			Publicado: "2026-05-27",
		},
	}
	for _, s := range samples {
		if err := q.CreatePost(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
