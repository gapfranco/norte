package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"norte/internal/models"
)

func scanPessoa(scanner interface {
	Scan(dest ...any) error
}) (models.Pessoa, error) {
	var p models.Pessoa
	var cpf, cnpj, email, telefone, celular sql.NullString
	var cep, logradouro, numero, complemento, bairro, cidade, uf, observacoes sql.NullString
	err := scanner.Scan(
		&p.ID, &p.TipoPessoa, &p.Nome,
		&cpf, &cnpj, &email, &telefone, &celular,
		&cep, &logradouro, &numero, &complemento, &bairro, &cidade, &uf,
		&observacoes,
	)
	if err != nil {
		return p, err
	}
	p.CPF = cpf.String
	p.CNPJ = cnpj.String
	p.Email = email.String
	p.Telefone = telefone.String
	p.Celular = celular.String
	p.CEP = cep.String
	p.Logradouro = logradouro.String
	p.Numero = numero.String
	p.Complemento = complemento.String
	p.Bairro = bairro.String
	p.Cidade = cidade.String
	p.UF = uf.String
	p.Observacoes = observacoes.String
	return p, nil
}

const pessoaSelectCols = `id, tipo_pessoa, nome, cpf, cnpj, email, telefone, celular,
	cep, logradouro, numero, complemento, bairro, cidade, uf, observacoes`

func (t *TursoDB) loadPessoaPapeis(pessoaID int) ([]models.Papel, error) {
	rows, err := t.db.Query(
		`SELECT p.codigo, p.nome FROM pessoa_papeis pp
		 JOIN papeis p ON p.codigo = pp.papel_codigo
		 WHERE pp.pessoa_id = ? ORDER BY p.nome`,
		pessoaID,
	)
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

func (t *TursoDB) syncPessoaPapeis(tx *sql.Tx, pessoaID int, papeis []string) error {
	if _, err := tx.Exec("DELETE FROM pessoa_papeis WHERE pessoa_id = ?", pessoaID); err != nil {
		return err
	}
	for _, codigo := range papeis {
		codigo = strings.TrimSpace(codigo)
		if codigo == "" {
			continue
		}
		if _, err := tx.Exec(
			"INSERT INTO pessoa_papeis (pessoa_id, papel_codigo) VALUES (?, ?)",
			pessoaID, codigo,
		); err != nil {
			return err
		}
	}
	return nil
}

func (t *TursoDB) ListPessoasFilter(nome, documento, papelCodigo string, limit, offset int) ([]models.Pessoa, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	where := `WHERE nome LIKE ? AND (COALESCE(cpf, '') LIKE ? OR COALESCE(cnpj, '') LIKE ?)`
	args := []any{like(nome), like(documento), like(documento)}
	join := ""
	if papelCodigo != "" {
		join = "JOIN pessoa_papeis pp ON pp.pessoa_id = pessoas.id"
		where += " AND pp.papel_codigo = ?"
		args = append(args, papelCodigo)
	}

	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(DISTINCT pessoas.id) FROM pessoas %s %s", join, where)
	if err := t.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	listSQL := fmt.Sprintf(
		`SELECT DISTINCT pessoas.id, pessoas.tipo_pessoa, pessoas.nome, pessoas.cpf, pessoas.cnpj,
			pessoas.email, pessoas.telefone, pessoas.celular,
			pessoas.cep, pessoas.logradouro, pessoas.numero, pessoas.complemento,
			pessoas.bairro, pessoas.cidade, pessoas.uf, pessoas.observacoes
		 FROM pessoas %s %s ORDER BY pessoas.nome LIMIT ? OFFSET ?`,
		join, where,
	)
	rows, err := t.db.Query(listSQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pessoas []models.Pessoa
	for rows.Next() {
		p, err := scanPessoa(rows)
		if err != nil {
			return nil, 0, err
		}
		papeis, err := t.loadPessoaPapeis(p.ID)
		if err != nil {
			return nil, 0, err
		}
		p.Papeis = papeis
		pessoas = append(pessoas, p)
	}
	return pessoas, total, rows.Err()
}

func (t *TursoDB) GetPessoa(id int) (*models.Pessoa, error) {
	row := t.db.QueryRow(
		`SELECT `+pessoaSelectCols+` FROM pessoas WHERE id = ?`, id,
	)
	p, err := scanPessoa(row)
	if err != nil {
		return nil, err
	}
	papeis, err := t.loadPessoaPapeis(p.ID)
	if err != nil {
		return nil, err
	}
	p.Papeis = papeis
	return &p, nil
}

func (t *TursoDB) CreatePessoa(p models.Pessoa, papeis []string) (int64, error) {
	tx, err := t.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO pessoas (
			tipo_pessoa, nome, cpf, cnpj, email, telefone, celular,
			cep, logradouro, numero, complemento, bairro, cidade, uf, observacoes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.TipoPessoa, p.Nome,
		nullIfEmpty(p.CPF), nullIfEmpty(p.CNPJ),
		nullIfEmpty(p.Email), nullIfEmpty(p.Telefone), nullIfEmpty(p.Celular),
		nullIfEmpty(p.CEP), nullIfEmpty(p.Logradouro), nullIfEmpty(p.Numero),
		nullIfEmpty(p.Complemento), nullIfEmpty(p.Bairro), nullIfEmpty(p.Cidade),
		nullIfEmpty(p.UF), nullIfEmpty(p.Observacoes),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := t.syncPessoaPapeis(tx, int(id), papeis); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (t *TursoDB) UpdatePessoa(p models.Pessoa, papeis []string) error {
	tx, err := t.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE pessoas SET
			tipo_pessoa = ?, nome = ?, cpf = ?, cnpj = ?,
			email = ?, telefone = ?, celular = ?,
			cep = ?, logradouro = ?, numero = ?, complemento = ?,
			bairro = ?, cidade = ?, uf = ?, observacoes = ?,
			atualizado_em = datetime('now')
		 WHERE id = ?`,
		p.TipoPessoa, p.Nome,
		nullIfEmpty(p.CPF), nullIfEmpty(p.CNPJ),
		nullIfEmpty(p.Email), nullIfEmpty(p.Telefone), nullIfEmpty(p.Celular),
		nullIfEmpty(p.CEP), nullIfEmpty(p.Logradouro), nullIfEmpty(p.Numero),
		nullIfEmpty(p.Complemento), nullIfEmpty(p.Bairro), nullIfEmpty(p.Cidade),
		nullIfEmpty(p.UF), nullIfEmpty(p.Observacoes),
		p.ID,
	)
	if err != nil {
		return err
	}
	if err := t.syncPessoaPapeis(tx, p.ID, papeis); err != nil {
		return err
	}
	return tx.Commit()
}

func (t *TursoDB) DeletePessoa(id int) error {
	_, err := t.db.Exec("DELETE FROM pessoas WHERE id = ?", id)
	return err
}
