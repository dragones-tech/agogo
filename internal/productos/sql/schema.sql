CREATE TABLE IF NOT EXISTS productos (
    slug        TEXT    PRIMARY KEY,
    titulo      TEXT    NOT NULL,
    descripcion TEXT    NOT NULL,
    precio      INTEGER NOT NULL
);
