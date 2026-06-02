package productos

import (
	"context"

	"agogo/internal/productos/db"
)

// Seed inserts sample data only if the table is empty. It's NOT part of the
// site: it's for DEVELOPMENT. Run it by hand with `go run ./cmd/seed`.
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
