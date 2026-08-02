package models

// Negocio representa uma empresa/entidade organizacional.
type Negocio struct {
	Codigo string
	Nome   string
	CNPJ   string
}

// Unidade representa uma filial/unidade operacional de um Negócio.
type Unidade struct {
	NegocioCodigo string
	Codigo        string
	Nome          string
	CNPJ          string
}
