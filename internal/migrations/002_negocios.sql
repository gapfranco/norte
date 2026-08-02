-- +goose Up
CREATE TABLE IF NOT EXISTS negocios (
    id            INTEGER PRIMARY KEY,
    nome          TEXT    NOT NULL,
    nome_fantasia TEXT    NOT NULL DEFAULT '',
    documento     TEXT    NOT NULL DEFAULT '',
    ativo         INTEGER NOT NULL DEFAULT 1,
    criado_em     TEXT    NOT NULL DEFAULT (datetime('now')),
    atualizado_em TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS unidades (
    id            INTEGER PRIMARY KEY,
    negocio_id    INTEGER NOT NULL REFERENCES negocios(id) ON DELETE CASCADE,
    nome          TEXT    NOT NULL,
    codigo        TEXT    NOT NULL DEFAULT '',
    ativo         INTEGER NOT NULL DEFAULT 1,
    criado_em     TEXT    NOT NULL DEFAULT (datetime('now')),
    atualizado_em TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_unidades_negocio ON unidades(negocio_id);

-- +goose Down
DROP INDEX IF EXISTS idx_unidades_negocio;
DROP TABLE IF EXISTS unidades;
DROP TABLE IF EXISTS negocios;
