-- name: ListProductos :many
SELECT slug, titulo, descripcion, precio
FROM productos
ORDER BY titulo;

-- name: GetProducto :one
SELECT slug, titulo, descripcion, precio
FROM productos
WHERE slug = ?;

-- name: SearchProductos :many
SELECT slug, titulo, descripcion, precio
FROM productos
WHERE titulo LIKE '%' || sqlc.arg(q) || '%'
   OR descripcion LIKE '%' || sqlc.arg(q) || '%'
ORDER BY titulo;

-- name: CountProductos :one
SELECT count(*) FROM productos;

-- name: CreateProducto :exec
INSERT INTO productos (slug, titulo, descripcion, precio)
VALUES (?, ?, ?, ?);
