package main

import "testing"

func TestValidarCNPJ(t *testing.T) {
	valid := []string{
		"11.444.777/0001-61", // numérico clássico
		"11444777000161",
		"", // campo vazio = válido (opcional)
	}
	invalid := []string{
		"11.111.111/1111-11", // todos iguais
		"00.000.000/0000-00",
		"12345678000100", // dígitos verificadores errados
		"1234567800010",  // tamanho errado
	}

	for _, cnpj := range valid {
		if !validarCNPJ(cnpj) {
			t.Errorf("esperado válido, rejeitado: %q", cnpj)
		}
	}
	for _, cnpj := range invalid {
		if validarCNPJ(cnpj) {
			t.Errorf("esperado inválido, aceito: %q", cnpj)
		}
	}
}

func TestNormalizarCNPJ(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"11.444.777/0001-61", "11444777000161"},
		{"  11444777000161  ", "11444777000161"},
		{"   ", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := normalizarCNPJ(tt.in); got != tt.want {
			t.Errorf("normalizarCNPJ(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
