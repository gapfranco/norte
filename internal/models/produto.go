package models

// Produto é um item comercializado, com ou sem controle de estoque.
type Produto struct {
	Codigo          string
	Descricao       string
	Unidade         string
	PrecoVenda      float64
	PrecoCusto      float64
	ControlaEstoque bool
	EstoqueMinimo   float64
	Observacoes     string
}
