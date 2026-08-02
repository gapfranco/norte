-- +goose Up
ALTER TABLE unidades ADD COLUMN cnpj TEXT;

CREATE UNIQUE INDEX idx_unidades_cnpj ON unidades(cnpj)
    WHERE cnpj IS NOT NULL AND cnpj != '';

-- +goose Down
DROP INDEX IF EXISTS idx_unidades_cnpj;

ALTER TABLE unidades DROP COLUMN cnpj;
