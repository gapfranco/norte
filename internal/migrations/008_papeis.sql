-- +goose Up
CREATE TABLE IF NOT EXISTS papeis (
    codigo TEXT PRIMARY KEY,
    nome   TEXT NOT NULL
);

INSERT INTO papeis (codigo, nome) VALUES
    ('cliente', 'Cliente'),
    ('fornecedor', 'Fornecedor'),
    ('vendedor', 'Vendedor'),
    ('representante', 'Representante')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS papeis;
