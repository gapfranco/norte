package storage

import "norte/internal/models"

// ─── Negócios ───────────────────────────────────────────────────────────────

func (t *TursoDB) ListNegociosFilter(nome string, limit, offset int) ([]models.Negocio, int, error) {
	like := "%" + nome + "%"
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM negocios WHERE (nome LIKE ? OR nome_fantasia LIKE ?)",
		like, like,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(`
		SELECT n.id, n.nome, n.nome_fantasia, n.documento, n.ativo,
		       (SELECT COUNT(*) FROM unidades u WHERE u.negocio_id = n.id) AS unidades
		FROM negocios n
		WHERE (n.nome LIKE ? OR n.nome_fantasia LIKE ?)
		ORDER BY n.nome
		LIMIT ? OFFSET ?`,
		like, like, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var negocios []models.Negocio
	for rows.Next() {
		var n models.Negocio
		if err := rows.Scan(&n.ID, &n.Nome, &n.NomeFantasia, &n.Documento, &n.Ativo, &n.Unidades); err != nil {
			return nil, 0, err
		}
		negocios = append(negocios, n)
	}
	return negocios, total, rows.Err()
}

func (t *TursoDB) GetNegocio(id int) (*models.Negocio, error) {
	var n models.Negocio
	err := t.db.QueryRow(
		"SELECT id, nome, nome_fantasia, documento, ativo FROM negocios WHERE id = ?",
		id,
	).Scan(&n.ID, &n.Nome, &n.NomeFantasia, &n.Documento, &n.Ativo)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (t *TursoDB) CreateNegocio(nome, nomeFantasia, documento string, ativo bool) error {
	_, err := t.db.Exec(
		"INSERT INTO negocios (nome, nome_fantasia, documento, ativo) VALUES (?, ?, ?, ?)",
		nome, nomeFantasia, documento, ativo,
	)
	return err
}

func (t *TursoDB) UpdateNegocio(id int, nome, nomeFantasia, documento string, ativo bool) error {
	_, err := t.db.Exec(
		"UPDATE negocios SET nome = ?, nome_fantasia = ?, documento = ?, ativo = ?, atualizado_em = datetime('now') WHERE id = ?",
		nome, nomeFantasia, documento, ativo, id,
	)
	return err
}

func (t *TursoDB) DeleteNegocio(id int) error {
	_, err := t.db.Exec("DELETE FROM negocios WHERE id = ?", id)
	return err
}

// ─── Unidades ─────────────────────────────────────────────────────────────────

func (t *TursoDB) ListUnidades(negocioID int) ([]models.Unidade, error) {
	rows, err := t.db.Query(
		"SELECT id, negocio_id, nome, codigo, ativo FROM unidades WHERE negocio_id = ? ORDER BY nome",
		negocioID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var unidades []models.Unidade
	for rows.Next() {
		var u models.Unidade
		if err := rows.Scan(&u.ID, &u.NegocioID, &u.Nome, &u.Codigo, &u.Ativo); err != nil {
			return nil, err
		}
		unidades = append(unidades, u)
	}
	return unidades, rows.Err()
}

func (t *TursoDB) GetUnidade(id int) (*models.Unidade, error) {
	var u models.Unidade
	err := t.db.QueryRow(
		"SELECT id, negocio_id, nome, codigo, ativo FROM unidades WHERE id = ?",
		id,
	).Scan(&u.ID, &u.NegocioID, &u.Nome, &u.Codigo, &u.Ativo)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (t *TursoDB) CreateUnidade(negocioID int, nome, codigo string, ativo bool) error {
	_, err := t.db.Exec(
		"INSERT INTO unidades (negocio_id, nome, codigo, ativo) VALUES (?, ?, ?, ?)",
		negocioID, nome, codigo, ativo,
	)
	return err
}

func (t *TursoDB) UpdateUnidade(id int, nome, codigo string, ativo bool) error {
	_, err := t.db.Exec(
		"UPDATE unidades SET nome = ?, codigo = ?, ativo = ?, atualizado_em = datetime('now') WHERE id = ?",
		nome, codigo, ativo, id,
	)
	return err
}

func (t *TursoDB) DeleteUnidade(id int) error {
	_, err := t.db.Exec("DELETE FROM unidades WHERE id = ?", id)
	return err
}
