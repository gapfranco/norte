package storage

import (
	"database/sql"

	"norte/internal/models"
)

func (t *TursoDB) ListCategoriasFilter(codigo, tipo string, limit, offset int) ([]models.Categoria, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM categorias WHERE codigo LIKE ? AND tipo LIKE ?",
		like(codigo), like(tipo),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		`SELECT codigo, tipo, classe FROM categorias
		 WHERE codigo LIKE ? AND tipo LIKE ?
		 ORDER BY codigo LIMIT ? OFFSET ?`,
		like(codigo), like(tipo), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.Categoria
	for rows.Next() {
		var c models.Categoria
		var classe sql.NullString
		if err := rows.Scan(&c.Codigo, &c.Tipo, &classe); err != nil {
			return nil, 0, err
		}
		c.Classe = classe.String
		list = append(list, c)
	}
	return list, total, rows.Err()
}

func (t *TursoDB) GetCategoria(codigo string) (*models.Categoria, error) {
	var c models.Categoria
	var classe sql.NullString
	err := t.db.QueryRow(
		"SELECT codigo, tipo, classe FROM categorias WHERE codigo = ?", codigo,
	).Scan(&c.Codigo, &c.Tipo, &classe)
	if err != nil {
		return nil, err
	}
	c.Classe = classe.String
	return &c, nil
}

func (t *TursoDB) CreateCategoria(c models.Categoria) error {
	_, err := t.db.Exec(
		"INSERT INTO categorias (codigo, tipo, classe) VALUES (?, ?, ?)",
		c.Codigo, c.Tipo, nullIfEmpty(c.Classe),
	)
	return err
}

func (t *TursoDB) UpdateCategoria(c models.Categoria) error {
	_, err := t.db.Exec(
		"UPDATE categorias SET tipo=?, classe=? WHERE codigo=?",
		c.Tipo, nullIfEmpty(c.Classe), c.Codigo,
	)
	return err
}

func (t *TursoDB) DeleteCategoria(codigo string) error {
	_, err := t.db.Exec("DELETE FROM categorias WHERE codigo = ?", codigo)
	return err
}
