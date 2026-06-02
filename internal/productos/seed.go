package productos

import (
	"context"

	"agogo/internal/productos/db"
)

// Seed inserta datos de ejemplo solo si la tabla está vacía. NO es parte del
// sitio: es para DESARROLLO. Se ejecuta a mano con `go run ./cmd/seed`.
func Seed(ctx context.Context, q *db.Queries) error {
	n, err := q.CountProductos(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	samples := []db.CreateProductoParams{
		{Slug: "cafe-organico", Titulo: "Café Orgánico", Descripcion: "Café de altura, tostado artesanal en lotes pequeños.", Precio: 180},
		{Slug: "miel-pura", Titulo: "Miel Pura", Descripcion: "Miel de abeja sin procesar, frasco de 500g.", Precio: 120},
		{Slug: "cacao-puro", Titulo: "Cacao Puro", Descripcion: "Cacao en polvo 100%, sin azúcar añadida.", Precio: 95},
	}
	for _, s := range samples {
		if err := q.CreateProducto(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
