package storage

import "norte/internal/models"

func (t *TursoDB) ListContasBancariasFilter(banco, descricao string, limit, offset int) ([]models.ContaBancaria, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM contas_bancarias WHERE banco LIKE ? AND descricao LIKE ?",
		like(banco), like(descricao),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		`SELECT banco, conta, agencia, tipo, descricao,
		        COALESCE(negocio_codigo,''), COALESCE(unidade_codigo,'')
		 FROM contas_bancarias WHERE banco LIKE ? AND descricao LIKE ?
		 ORDER BY banco, conta LIMIT ? OFFSET ?`,
		like(banco), like(descricao), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.ContaBancaria
	for rows.Next() {
		var c models.ContaBancaria
		if err := rows.Scan(&c.Banco, &c.Conta, &c.Agencia, &c.Tipo, &c.Descricao, &c.NegocioCodigo, &c.UnidadeCodigo); err != nil {
			return nil, 0, err
		}
		list = append(list, c)
	}
	return list, total, rows.Err()
}

func (t *TursoDB) GetContaBancaria(banco, conta string) (*models.ContaBancaria, error) {
	var c models.ContaBancaria
	err := t.db.QueryRow(
		`SELECT banco, conta, agencia, tipo, descricao,
		        COALESCE(negocio_codigo,''), COALESCE(unidade_codigo,'')
		 FROM contas_bancarias WHERE banco = ? AND conta = ?`,
		banco, conta,
	).Scan(&c.Banco, &c.Conta, &c.Agencia, &c.Tipo, &c.Descricao, &c.NegocioCodigo, &c.UnidadeCodigo)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (t *TursoDB) CreateContaBancaria(c models.ContaBancaria) error {
	var negocioCodigo, unidadeCodigo any
	if c.NegocioCodigo != "" {
		negocioCodigo = c.NegocioCodigo
	}
	if c.UnidadeCodigo != "" {
		unidadeCodigo = c.UnidadeCodigo
	}
	_, err := t.db.Exec(
		`INSERT INTO contas_bancarias (banco, conta, agencia, tipo, descricao, negocio_codigo, unidade_codigo)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.Banco, c.Conta, c.Agencia, c.Tipo, c.Descricao, negocioCodigo, unidadeCodigo,
	)
	return err
}

func (t *TursoDB) UpdateContaBancaria(c models.ContaBancaria) error {
	var negocioCodigo, unidadeCodigo any
	if c.NegocioCodigo != "" {
		negocioCodigo = c.NegocioCodigo
	}
	if c.UnidadeCodigo != "" {
		unidadeCodigo = c.UnidadeCodigo
	}
	_, err := t.db.Exec(
		`UPDATE contas_bancarias SET agencia=?, tipo=?, descricao=?, negocio_codigo=?, unidade_codigo=?
		 WHERE banco=? AND conta=?`,
		c.Agencia, c.Tipo, c.Descricao, negocioCodigo, unidadeCodigo, c.Banco, c.Conta,
	)
	return err
}

func (t *TursoDB) DeleteContaBancaria(banco, conta string) error {
	_, err := t.db.Exec("DELETE FROM contas_bancarias WHERE banco = ? AND conta = ?", banco, conta)
	return err
}
