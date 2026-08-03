-- +goose Up
CREATE TABLE IF NOT EXISTS produtos (
    codigo           TEXT PRIMARY KEY,
    descricao        TEXT NOT NULL,
    unidade          TEXT NOT NULL DEFAULT 'UN',
    preco_venda      REAL NOT NULL DEFAULT 0,
    preco_custo      REAL NOT NULL DEFAULT 0,
    controla_estoque INTEGER NOT NULL DEFAULT 1,
    estoque_minimo   REAL NOT NULL DEFAULT 0,
    observacoes      TEXT,
    criado_em        TEXT NOT NULL DEFAULT (datetime('now')),
    atualizado_em    TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS produtos;
