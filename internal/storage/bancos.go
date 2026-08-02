package storage

import "norte/internal/models"

func (t *TursoDB) ListBancosFilter(codigo, nome string, limit, offset int) ([]models.Banco, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM bancos WHERE codigo LIKE ? AND nome LIKE ?",
		like(codigo), like(nome),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		"SELECT codigo, nome FROM bancos WHERE codigo LIKE ? AND nome LIKE ? ORDER BY codigo LIMIT ? OFFSET ?",
		like(codigo), like(nome), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var bancos []models.Banco
	for rows.Next() {
		var b models.Banco
		if err := rows.Scan(&b.Codigo, &b.Nome); err != nil {
			return nil, 0, err
		}
		bancos = append(bancos, b)
	}
	return bancos, total, rows.Err()
}

func (t *TursoDB) GetBanco(codigo string) (*models.Banco, error) {
	var b models.Banco
	err := t.db.QueryRow("SELECT codigo, nome FROM bancos WHERE codigo = ?", codigo).
		Scan(&b.Codigo, &b.Nome)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (t *TursoDB) CreateBanco(b models.Banco) error {
	_, err := t.db.Exec("INSERT INTO bancos (codigo, nome) VALUES (?, ?)", b.Codigo, b.Nome)
	return err
}

func (t *TursoDB) UpdateBanco(b models.Banco) error {
	_, err := t.db.Exec("UPDATE bancos SET nome = ? WHERE codigo = ?", b.Nome, b.Codigo)
	return err
}

func (t *TursoDB) DeleteBanco(codigo string) error {
	_, err := t.db.Exec("DELETE FROM bancos WHERE codigo = ?", codigo)
	return err
}
