package auth

import (
	"context"

	"jehosogo/internal/auth/db"
	"jehosogo/internal/password"
)

// Seed crea un usuario demo si la tabla está vacía (solo desarrollo).
//
//	email: admin@jehosogo.com   contraseña: demo1234
func Seed(ctx context.Context, q *db.Queries) error {
	n, err := q.CountUsuarios(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := password.Hash("demo1234")
	if err != nil {
		return err
	}
	return q.CreateUsuario(ctx, db.CreateUsuarioParams{
		Email:        "admin@jehosogo.com",
		PasswordHash: hash,
	})
}
