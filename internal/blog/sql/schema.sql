CREATE TABLE IF NOT EXISTS posts (
    slug      TEXT PRIMARY KEY,
    titulo    TEXT NOT NULL,
    resumen   TEXT NOT NULL,
    contenido TEXT NOT NULL,
    publicado TEXT NOT NULL   -- fecha ISO 8601 (YYYY-MM-DD)
);
