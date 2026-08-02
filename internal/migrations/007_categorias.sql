-- +goose Up
CREATE TABLE IF NOT EXISTS categorias (
    codigo TEXT PRIMARY KEY,
    tipo   TEXT NOT NULL,
    classe TEXT,
    CHECK ((tipo = 'E' AND classe IS NULL) OR (tipo = 'S' AND classe IS NOT NULL))
);

-- +goose Down
DROP TABLE IF EXISTS categorias;
