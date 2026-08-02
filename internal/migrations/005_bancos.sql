-- +goose Up
CREATE TABLE IF NOT EXISTS bancos (
    codigo TEXT PRIMARY KEY,
    nome   TEXT NOT NULL
);

INSERT INTO bancos (codigo, nome) VALUES
    ('001', 'Banco do Brasil'),
    ('033', 'Santander'),
    ('041', 'Banrisul'),
    ('047', 'Banese'),
    ('070', 'BRB - Banco de Brasília'),
    ('077', 'Banco Inter'),
    ('085', 'Cecred / Ailos'),
    ('097', 'Credisis'),
    ('104', 'Caixa Econômica Federal'),
    ('136', 'Unicred'),
    ('212', 'Banco Original'),
    ('237', 'Bradesco'),
    ('260', 'Nubank'),
    ('290', 'PagBank'),
    ('323', 'Mercado Pago'),
    ('336', 'C6 Bank'),
    ('341', 'Itaú Unibanco'),
    ('389', 'Mercantil do Brasil'),
    ('422', 'Safra'),
    ('655', 'Votorantim'),
    ('748', 'Sicredi'),
    ('756', 'Sicoob')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS bancos;
