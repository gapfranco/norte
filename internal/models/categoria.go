package models

// Categoria representa uma categoria financeira de movimento.
// Tipo: E = entrada, S = saída. Classe é obrigatória apenas para saídas.
type Categoria struct {
	Codigo string
	Tipo   string
	Classe string
}
