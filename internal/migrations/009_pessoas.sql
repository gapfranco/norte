-- +goose Up
CREATE TABLE IF NOT EXISTS pessoas (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    tipo_pessoa   TEXT NOT NULL CHECK (tipo_pessoa IN ('F', 'J')),
    nome          TEXT NOT NULL,
    cpf           TEXT,
    cnpj          TEXT,
    email         TEXT,
    telefone      TEXT,
    celular       TEXT,
    cep           TEXT,
    logradouro    TEXT,
    numero        TEXT,
    complemento   TEXT,
    bairro        TEXT,
    cidade        TEXT,
    uf            TEXT,
    observacoes   TEXT,
    criado_em     TEXT NOT NULL DEFAULT (datetime('now')),
    atualizado_em TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX idx_pessoas_cpf ON pessoas(cpf)
    WHERE cpf IS NOT NULL AND cpf != '';

CREATE UNIQUE INDEX idx_pessoas_cnpj ON pessoas(cnpj)
    WHERE cnpj IS NOT NULL AND cnpj != '';

CREATE TABLE IF NOT EXISTS pessoa_papeis (
    pessoa_id    INTEGER NOT NULL REFERENCES pessoas(id) ON DELETE CASCADE,
    papel_codigo TEXT NOT NULL REFERENCES papeis(codigo),
    PRIMARY KEY (pessoa_id, papel_codigo)
);

-- +goose Down
DROP TABLE IF EXISTS pessoa_papeis;
DROP INDEX IF EXISTS idx_pessoas_cnpj;
DROP INDEX IF EXISTS idx_pessoas_cpf;
DROP TABLE IF EXISTS pessoas;
