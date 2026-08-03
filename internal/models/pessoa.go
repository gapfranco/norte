package models

// Pessoa é um cadastro unificado (PF/PJ) usado como cliente, fornecedor, etc.
// TipoPessoa: F = física, J = jurídica.
type Pessoa struct {
	ID          int
	TipoPessoa  string
	Nome        string
	CPF         string
	CNPJ        string
	Email       string
	Telefone    string
	Celular     string
	CEP         string
	Logradouro  string
	Numero      string
	Complemento string
	Bairro      string
	Cidade      string
	UF          string
	Observacoes string
	Papeis      []Papel
}
