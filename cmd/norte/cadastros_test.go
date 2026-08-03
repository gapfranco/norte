package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBancosContasCategoriasCRUD(t *testing.T) {
	srv, client := newAuthedServer(t)

	// Bancos — seed já traz 001; lista carrega.
	resp, body := mustGet(t, client, srv.URL+"/cadastros/bancos")
	if resp.StatusCode != 200 || !strings.Contains(body, "Banco do Brasil") {
		t.Fatalf("lista bancos: %d", resp.StatusCode)
	}

	// Cria banco customizado.
	_, body = mustGet(t, client, srv.URL+"/cadastros/bancos/new")
	token := mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/bancos/new", url.Values{
		"csrf_token": {token}, "codigo": {"999"}, "nome": {"Banco Teste"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria banco %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/bancos?codigo=999")
	if !strings.Contains(body, "Banco Teste") {
		t.Fatal("banco não apareceu na lista")
	}

	// Edita banco.
	_, body = mustGet(t, client, srv.URL+"/cadastros/bancos/999/edit")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/bancos/999/edit", url.Values{
		"csrf_token": {token}, "nome": {"Banco Teste SA"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("edita banco %d", resp.StatusCode)
	}

	// Conta sob banco seed 001.
	_, body = mustGet(t, client, srv.URL+"/cadastros/contas/new")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/contas/new", url.Values{
		"csrf_token": {token},
		"banco":      {"001"},
		"conta":      {"12345-6"},
		"agencia":    {"0001"},
		"tipo":       {"Corrente"},
		"descricao":  {"Conta Matriz"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria conta %d", resp.StatusCode)
	}
	resp, body = mustGet(t, client, srv.URL+"/cadastros/contas")
	if resp.StatusCode != 200 || !strings.Contains(body, "Conta Matriz") {
		t.Fatal("conta não apareceu na lista")
	}

	// Excluir banco 001 com conta vinculada deve falhar (FK) e redirecionar com flash.
	_, body = mustGet(t, client, srv.URL+"/cadastros/bancos")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/bancos/001/delete", url.Values{
		"csrf_token": {token},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("delete banco com FK %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/bancos")
	if !strings.Contains(body, "Banco do Brasil") {
		t.Fatal("banco 001 não deveria ter sido excluído")
	}

	// Categoria entrada.
	_, body = mustGet(t, client, srv.URL+"/cadastros/categorias/new")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/categorias/new", url.Values{
		"csrf_token": {token}, "codigo": {"REC01"}, "tipo": {"E"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria categoria E %d", resp.StatusCode)
	}

	// Categoria saída exige classe.
	_, body = mustGet(t, client, srv.URL+"/cadastros/categorias/new")
	token = mustCSRF(t, body)
	resp = mustPost(t, client, srv.URL+"/cadastros/categorias/new", url.Values{
		"csrf_token": {token}, "codigo": {"DES01"}, "tipo": {"S"},
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("categoria S sem classe deveria 422, got %d", resp.StatusCode)
	}
	if resp := mustPost(t, client, srv.URL+"/cadastros/categorias/new", url.Values{
		"csrf_token": {token}, "codigo": {"DES01"}, "tipo": {"S"}, "classe": {"Custo Fixo"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria categoria S %d", resp.StatusCode)
	}

	resp, body = mustGet(t, client, srv.URL+"/cadastros/categorias")
	if !strings.Contains(body, "REC01") || !strings.Contains(body, "DES01") {
		t.Fatal("categorias não apareceram na lista")
	}

	// Exclui conta e depois banco customizado.
	_, body = mustGet(t, client, srv.URL+"/cadastros/contas")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/contas/001/12345-6/delete", url.Values{
		"csrf_token": {token},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("exclui conta %d", resp.StatusCode)
	}
	if resp := mustPost(t, client, srv.URL+"/cadastros/bancos/999/delete", url.Values{
		"csrf_token": {token},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("exclui banco 999 %d", resp.StatusCode)
	}
}

func TestPapeisPessoasProdutosCRUD(t *testing.T) {
	srv, client := newAuthedServer(t)

	// Papéis — seed.
	resp, body := mustGet(t, client, srv.URL+"/cadastros/papeis")
	if resp.StatusCode != 200 || !strings.Contains(body, "Cliente") {
		t.Fatalf("lista papeis: %d", resp.StatusCode)
	}

	_, body = mustGet(t, client, srv.URL+"/cadastros/papeis/new")
	token := mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/papeis/new", url.Values{
		"csrf_token": {token}, "codigo": {"transportadora"}, "nome": {"Transportadora"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria papel %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/papeis?codigo=transportadora")
	if !strings.Contains(body, "Transportadora") {
		t.Fatal("papel customizado não apareceu")
	}

	// Pessoa PF com papéis.
	_, body = mustGet(t, client, srv.URL+"/cadastros/pessoas/new")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/pessoas/new", url.Values{
		"csrf_token":   {token},
		"tipo_pessoa":  {"F"},
		"nome":         {"Maria Silva"},
		"cpf":          {"123.456.789-09"},
		"email":        {"maria@exemplo.com"},
		"papeis":       {"cliente", "vendedor"},
		"cidade":       {"Curitiba"},
		"uf":           {"PR"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria pessoa %d", resp.StatusCode)
	}
	resp, body = mustGet(t, client, srv.URL+"/cadastros/pessoas?nome=Maria")
	if resp.StatusCode != 200 || !strings.Contains(body, "Maria Silva") || !strings.Contains(body, "Cliente") {
		t.Fatal("pessoa não apareceu na lista com papéis")
	}

	// Sem papel → 422.
	_, body = mustGet(t, client, srv.URL+"/cadastros/pessoas/new")
	token = mustCSRF(t, body)
	resp = mustPost(t, client, srv.URL+"/cadastros/pessoas/new", url.Values{
		"csrf_token": {token}, "tipo_pessoa": {"J"}, "nome": {"Empresa X"},
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("pessoa sem papel deveria 422, got %d", resp.StatusCode)
	}

	// Excluir papel seed vinculado deve falhar (FK).
	_, body = mustGet(t, client, srv.URL+"/cadastros/papeis")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/papeis/cliente/delete", url.Values{
		"csrf_token": {token},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("delete papel com FK %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/papeis")
	if !strings.Contains(body, "Cliente") {
		t.Fatal("papel cliente não deveria ter sido excluído")
	}

	// Produto.
	_, body = mustGet(t, client, srv.URL+"/cadastros/produtos/new")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/produtos/new", url.Values{
		"csrf_token":        {token},
		"codigo":            {"SKU-001"},
		"descricao":         {"Parafuso M6"},
		"unidade":           {"UN"},
		"preco_venda":       {"1.50"},
		"preco_custo":       {"0.80"},
		"controla_estoque":  {"1"},
		"estoque_minimo":    {"10"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria produto %d", resp.StatusCode)
	}
	resp, body = mustGet(t, client, srv.URL+"/cadastros/produtos?codigo=SKU-001")
	if resp.StatusCode != 200 || !strings.Contains(body, "Parafuso M6") {
		t.Fatal("produto não apareceu na lista")
	}

	_, body = mustGet(t, client, srv.URL+"/cadastros/produtos/SKU-001/edit")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/produtos/SKU-001/edit", url.Values{
		"csrf_token":       {token},
		"descricao":        {"Parafuso M6 Zincado"},
		"unidade":          {"UN"},
		"preco_venda":      {"1.75"},
		"preco_custo":      {"0.90"},
		"controla_estoque": {"1"},
		"estoque_minimo":   {"20"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("edita produto %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/produtos?codigo=SKU-001")
	if !strings.Contains(body, "Parafuso M6 Zincado") {
		t.Fatal("produto não foi atualizado")
	}

	if resp := mustPost(t, client, srv.URL+"/cadastros/produtos/SKU-001/delete", url.Values{
		"csrf_token": {token},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("exclui produto %d", resp.StatusCode)
	}
}
