-- +goose Up
CREATE TABLE IF NOT EXISTS contas_bancarias (
    banco           TEXT NOT NULL REFERENCES bancos(codigo),
    conta           TEXT NOT NULL,
    agencia         TEXT NOT NULL DEFAULT '',
    tipo            TEXT NOT NULL DEFAULT '',
    descricao       TEXT NOT NULL DEFAULT '',
    negocio_codigo  TEXT,
    unidade_codigo  TEXT,
    PRIMARY KEY (banco, conta),
    FOREIGN KEY (negocio_codigo, unidade_codigo)
        REFERENCES unidades(negocio_codigo, codigo)
);

-- +goose Down
DROP TABLE IF EXISTS contas_bancarias;
