package storage

import (
	"database/sql"

	"norte/internal/models"
)

func (t *TursoDB) ListProdutosFilter(codigo, descricao string, limit, offset int) ([]models.Produto, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM produtos WHERE codigo LIKE ? AND descricao LIKE ?",
		like(codigo), like(descricao),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		`SELECT codigo, descricao, unidade, preco_venda, preco_custo,
			controla_estoque, estoque_minimo, observacoes
		 FROM produtos WHERE codigo LIKE ? AND descricao LIKE ?
		 ORDER BY codigo LIMIT ? OFFSET ?`,
		like(codigo), like(descricao), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var produtos []models.Produto
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return nil, 0, err
		}
		produtos = append(produtos, p)
	}
	return produtos, total, rows.Err()
}

func scanProduto(scanner interface {
	Scan(dest ...any) error
}) (models.Produto, error) {
	var p models.Produto
	var controla int
	var obs sql.NullString
	err := scanner.Scan(
		&p.Codigo, &p.Descricao, &p.Unidade,
		&p.PrecoVenda, &p.PrecoCusto, &controla, &p.EstoqueMinimo, &obs,
	)
	if err != nil {
		return p, err
	}
	p.ControlaEstoque = controla != 0
	p.Observacoes = obs.String
	return p, nil
}

func (t *TursoDB) GetProduto(codigo string) (*models.Produto, error) {
	row := t.db.QueryRow(
		`SELECT codigo, descricao, unidade, preco_venda, preco_custo,
			controla_estoque, estoque_minimo, observacoes
		 FROM produtos WHERE codigo = ?`, codigo,
	)
	p, err := scanProduto(row)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (t *TursoDB) CreateProduto(p models.Produto) error {
	controla := 0
	if p.ControlaEstoque {
		controla = 1
	}
	_, err := t.db.Exec(
		`INSERT INTO produtos (
			codigo, descricao, unidade, preco_venda, preco_custo,
			controla_estoque, estoque_minimo, observacoes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Codigo, p.Descricao, p.Unidade, p.PrecoVenda, p.PrecoCusto,
		controla, p.EstoqueMinimo, nullIfEmpty(p.Observacoes),
	)
	return err
}

func (t *TursoDB) UpdateProduto(p models.Produto) error {
	controla := 0
	if p.ControlaEstoque {
		controla = 1
	}
	_, err := t.db.Exec(
		`UPDATE produtos SET
			descricao = ?, unidade = ?, preco_venda = ?, preco_custo = ?,
			controla_estoque = ?, estoque_minimo = ?, observacoes = ?,
			atualizado_em = datetime('now')
		 WHERE codigo = ?`,
		p.Descricao, p.Unidade, p.PrecoVenda, p.PrecoCusto,
		controla, p.EstoqueMinimo, nullIfEmpty(p.Observacoes),
		p.Codigo,
	)
	return err
}

func (t *TursoDB) DeleteProduto(codigo string) error {
	_, err := t.db.Exec("DELETE FROM produtos WHERE codigo = ?", codigo)
	return err
}
