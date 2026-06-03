-- name: CreateUsuario :exec
INSERT INTO usuarios (email, password_hash) VALUES (?, ?);

-- name: GetUsuarioByEmail :one
SELECT id, email, password_hash, creado FROM usuarios WHERE email = ?;

-- name: GetUsuario :one
SELECT id, email, password_hash, creado FROM usuarios WHERE id = ?;

-- name: CountUsuarios :one
SELECT count(*) FROM usuarios;
