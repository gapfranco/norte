-- +goose Up
PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS unidades;
DROP TABLE IF EXISTS negocios;

CREATE TABLE negocios (
    codigo        TEXT PRIMARY KEY,
    nome          TEXT NOT NULL,
    cnpj          TEXT,
    criado_em     TEXT NOT NULL DEFAULT (datetime('now')),
    atualizado_em TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE unidades (
    negocio_codigo TEXT NOT NULL REFERENCES negocios(codigo) ON DELETE CASCADE,
    codigo         TEXT NOT NULL,
    nome           TEXT NOT NULL,
    criado_em      TEXT NOT NULL DEFAULT (datetime('now')),
    atualizado_em  TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (negocio_codigo, codigo)
);

CREATE INDEX idx_unidades_negocio ON unidades(negocio_codigo);

CREATE UNIQUE INDEX idx_negocios_cnpj ON negocios(cnpj)
    WHERE cnpj IS NOT NULL AND cnpj != '';

PRAGMA foreign_keys = ON;

-- +goose Down
PRAGMA foreign_keys = OFF;

DROP INDEX IF EXISTS idx_negocios_cnpj;
DROP INDEX IF EXISTS idx_unidades_negocio;
DROP TABLE IF EXISTS unidades;
DROP TABLE IF EXISTS negocios;

CREATE TABLE negocios (
    id            INTEGER PRIMARY KEY,
    nome          TEXT    NOT NULL,
    nome_fantasia TEXT    NOT NULL DEFAULT '',
    documento     TEXT    NOT NULL DEFAULT '',
    ativo         INTEGER NOT NULL DEFAULT 1,
    criado_em     TEXT    NOT NULL DEFAULT (datetime('now')),
    atualizado_em TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE unidades (
    id            INTEGER PRIMARY KEY,
    negocio_id    INTEGER NOT NULL REFERENCES negocios(id) ON DELETE CASCADE,
    nome          TEXT    NOT NULL,
    codigo        TEXT    NOT NULL DEFAULT '',
    ativo         INTEGER NOT NULL DEFAULT 1,
    criado_em     TEXT    NOT NULL DEFAULT (datetime('now')),
    atualizado_em TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_unidades_negocio ON unidades(negocio_id);

PRAGMA foreign_keys = ON;
