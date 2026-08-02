package models

// ContaBancaria representa uma conta vinculada a um banco.
type ContaBancaria struct {
	Banco          string
	Conta          string
	Agencia        string
	Tipo           string
	Descricao      string
	NegocioCodigo  string
	UnidadeCodigo  string
}
