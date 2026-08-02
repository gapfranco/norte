package models

// Negocio representa uma empresa/entidade organizacional (antigo "Empresa").
type Negocio struct {
	ID           int
	Nome         string
	NomeFantasia string
	Documento    string
	Ativo        bool
	Unidades     int // contagem de unidades vinculadas (uso em listagem)
}

// Unidade representa uma filial/unidade operacional de um Negócio.
type Unidade struct {
	ID        int
	NegocioID int
	Nome      string
	Codigo    string
	Ativo     bool
}
