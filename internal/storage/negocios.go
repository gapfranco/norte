package storage

import (
	"database/sql"

	"norte/internal/models"
)

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ─── Negócios ───────────────────────────────────────────────────────────────

func (t *TursoDB) ListNegociosFilter(codigo, nome, cnpj string, limit, offset int) ([]models.Negocio, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM negocios WHERE codigo LIKE ? AND nome LIKE ? AND COALESCE(cnpj, '') LIKE ?",
		like(codigo), like(nome), like(cnpj),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		`SELECT codigo, nome, cnpj FROM negocios
		 WHERE codigo LIKE ? AND nome LIKE ? AND COALESCE(cnpj, '') LIKE ?
		 ORDER BY codigo LIMIT ? OFFSET ?`,
		like(codigo), like(nome), like(cnpj), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var negocios []models.Negocio
	for rows.Next() {
		var n models.Negocio
		var cnpj sql.NullString
		if err := rows.Scan(&n.Codigo, &n.Nome, &cnpj); err != nil {
			return nil, 0, err
		}
		n.CNPJ = cnpj.String
		negocios = append(negocios, n)
	}
	return negocios, total, rows.Err()
}

func (t *TursoDB) GetNegocio(codigo string) (*models.Negocio, error) {
	var n models.Negocio
	var cnpj sql.NullString
	err := t.db.QueryRow(
		"SELECT codigo, nome, cnpj FROM negocios WHERE codigo = ?",
		codigo,
	).Scan(&n.Codigo, &n.Nome, &cnpj)
	if err != nil {
		return nil, err
	}
	n.CNPJ = cnpj.String
	return &n, nil
}

func (t *TursoDB) CreateNegocio(n models.Negocio) error {
	_, err := t.db.Exec(
		"INSERT INTO negocios (codigo, nome, cnpj) VALUES (?, ?, ?)",
		n.Codigo, n.Nome, nullIfEmpty(n.CNPJ),
	)
	return err
}

func (t *TursoDB) UpdateNegocio(n models.Negocio) error {
	_, err := t.db.Exec(
		"UPDATE negocios SET nome = ?, cnpj = ?, atualizado_em = datetime('now') WHERE codigo = ?",
		n.Nome, nullIfEmpty(n.CNPJ), n.Codigo,
	)
	return err
}

func (t *TursoDB) DeleteNegocio(codigo string) error {
	_, err := t.db.Exec("DELETE FROM negocios WHERE codigo = ?", codigo)
	return err
}

// ─── Unidades ─────────────────────────────────────────────────────────────────

func (t *TursoDB) ListUnidades(negocioCodigo string, limit, offset int) ([]models.Unidade, int, error) {
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM unidades WHERE negocio_codigo = ?",
		negocioCodigo,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		"SELECT negocio_codigo, codigo, nome, cnpj FROM unidades WHERE negocio_codigo = ? ORDER BY codigo LIMIT ? OFFSET ?",
		negocioCodigo, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var unidades []models.Unidade
	for rows.Next() {
		var u models.Unidade
		var cnpj sql.NullString
		if err := rows.Scan(&u.NegocioCodigo, &u.Codigo, &u.Nome, &cnpj); err != nil {
			return nil, 0, err
		}
		u.CNPJ = cnpj.String
		unidades = append(unidades, u)
	}
	return unidades, total, rows.Err()
}

func (t *TursoDB) GetUnidade(negocioCodigo, codigo string) (*models.Unidade, error) {
	var u models.Unidade
	var cnpj sql.NullString
	err := t.db.QueryRow(
		"SELECT negocio_codigo, codigo, nome, cnpj FROM unidades WHERE negocio_codigo = ? AND codigo = ?",
		negocioCodigo, codigo,
	).Scan(&u.NegocioCodigo, &u.Codigo, &u.Nome, &cnpj)
	if err != nil {
		return nil, err
	}
	u.CNPJ = cnpj.String
	return &u, nil
}

func (t *TursoDB) CreateUnidade(u models.Unidade) error {
	_, err := t.db.Exec(
		"INSERT INTO unidades (negocio_codigo, codigo, nome, cnpj) VALUES (?, ?, ?, ?)",
		u.NegocioCodigo, u.Codigo, u.Nome, nullIfEmpty(u.CNPJ),
	)
	return err
}

func (t *TursoDB) UpdateUnidade(u models.Unidade) error {
	_, err := t.db.Exec(
		"UPDATE unidades SET nome = ?, cnpj = ?, atualizado_em = datetime('now') WHERE negocio_codigo = ? AND codigo = ?",
		u.Nome, nullIfEmpty(u.CNPJ), u.NegocioCodigo, u.Codigo,
	)
	return err
}

func (t *TursoDB) DeleteUnidade(negocioCodigo, codigo string) error {
	_, err := t.db.Exec(
		"DELETE FROM unidades WHERE negocio_codigo = ? AND codigo = ?",
		negocioCodigo, codigo,
	)
	return err
}
