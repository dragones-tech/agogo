-- name: CreateContacto :exec
INSERT INTO contactos (nombre, email, mensaje)
VALUES (?, ?, ?);
