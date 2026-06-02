-- name: ListPosts :many
SELECT slug, titulo, resumen, contenido, publicado
FROM posts
ORDER BY publicado DESC;

-- name: GetPost :one
SELECT slug, titulo, resumen, contenido, publicado
FROM posts
WHERE slug = ?;

-- name: CountPosts :one
SELECT count(*) FROM posts;

-- name: CreatePost :exec
INSERT INTO posts (slug, titulo, resumen, contenido, publicado)
VALUES (?, ?, ?, ?, ?);
