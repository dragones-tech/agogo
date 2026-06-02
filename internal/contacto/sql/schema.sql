CREATE TABLE IF NOT EXISTS contactos (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre  TEXT NOT NULL,
    email   TEXT NOT NULL,
    mensaje TEXT NOT NULL,
    creado  TEXT NOT NULL DEFAULT (datetime('now'))
);
