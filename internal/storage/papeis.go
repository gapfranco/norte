package storage

import "norte/internal/models"

func (t *TursoDB) ListPapeisFilter(codigo, nome string, limit, offset int) ([]models.Papel, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM papeis WHERE codigo LIKE ? AND nome LIKE ?",
		like(codigo), like(nome),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		"SELECT codigo, nome FROM papeis WHERE codigo LIKE ? AND nome LIKE ? ORDER BY codigo LIMIT ? OFFSET ?",
		like(codigo), like(nome), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var papeis []models.Papel
	for rows.Next() {
		var p models.Papel
		if err := rows.Scan(&p.Codigo, &p.Nome); err != nil {
			return nil, 0, err
		}
		papeis = append(papeis, p)
	}
	return papeis, total, rows.Err()
}

func (t *TursoDB) ListPapeisAll() ([]models.Papel, error) {
	rows, err := t.db.Query("SELECT codigo, nome FROM papeis ORDER BY nome")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var papeis []models.Papel
	for rows.Next() {
		var p models.Papel
		if err := rows.Scan(&p.Codigo, &p.Nome); err != nil {
			return nil, err
		}
		papeis = append(papeis, p)
	}
	return papeis, rows.Err()
}

func (t *TursoDB) GetPapel(codigo string) (*models.Papel, error) {
	var p models.Papel
	err := t.db.QueryRow("SELECT codigo, nome FROM papeis WHERE codigo = ?", codigo).
		Scan(&p.Codigo, &p.Nome)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (t *TursoDB) CreatePapel(p models.Papel) error {
	_, err := t.db.Exec("INSERT INTO papeis (codigo, nome) VALUES (?, ?)", p.Codigo, p.Nome)
	return err
}

func (t *TursoDB) UpdatePapel(p models.Papel) error {
	_, err := t.db.Exec("UPDATE papeis SET nome = ? WHERE codigo = ?", p.Nome, p.Codigo)
	return err
}

func (t *TursoDB) DeletePapel(codigo string) error {
	_, err := t.db.Exec("DELETE FROM papeis WHERE codigo = ?", codigo)
	return err
}
