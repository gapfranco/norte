package main

import (
	"net/http"
	"strconv"
	"strings"

	"norte/internal/models"
)

const produtosPageSize = 10

var unidadesProduto = []string{"UN", "KG", "CX", "MT", "LT", "PC", "DZ", "PAR", "CJ"}

type produtoForm struct {
	Codigo          string  `form:"codigo"`
	Descricao       string  `form:"descricao"`
	Unidade         string  `form:"unidade"`
	PrecoVenda      float64 `form:"preco_venda"`
	PrecoCusto      float64 `form:"preco_custo"`
	ControlaEstoque string  `form:"controla_estoque"`
	EstoqueMinimo   float64 `form:"estoque_minimo"`
	Observacoes     string  `form:"observacoes"`
}

func (f produtoForm) toProduto(codigo string) models.Produto {
	if codigo == "" {
		codigo = strings.TrimSpace(f.Codigo)
	}
	return models.Produto{
		Codigo:          codigo,
		Descricao:       strings.TrimSpace(f.Descricao),
		Unidade:         strings.TrimSpace(f.Unidade),
		PrecoVenda:      f.PrecoVenda,
		PrecoCusto:      f.PrecoCusto,
		ControlaEstoque: f.ControlaEstoque == "1",
		EstoqueMinimo:   f.EstoqueMinimo,
		Observacoes:     strings.TrimSpace(f.Observacoes),
	}
}

func validUnidadeProduto(u string) bool {
	for _, v := range unidadesProduto {
		if v == u {
			return true
		}
	}
	return false
}

func validateProdutoForm(f *produtoForm, requireCodigo bool) string {
	if requireCodigo && strings.TrimSpace(f.Codigo) == "" {
		return "Código é obrigatório"
	}
	if strings.TrimSpace(f.Descricao) == "" {
		return "Descrição é obrigatória"
	}
	if !validUnidadeProduto(strings.TrimSpace(f.Unidade)) {
		return "Unidade inválida"
	}
	if f.PrecoVenda < 0 || f.PrecoCusto < 0 || f.EstoqueMinimo < 0 {
		return "Valores não podem ser negativos"
	}
	return ""
}

func (app *application) renderProdutoForm(w http.ResponseWriter, r *http.Request, status int, flash string, editMode bool, produto *models.Produto) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "produtos"
	data.Flash = flash
	data.Data = struct {
		EditMode bool
		Produto  *models.Produto
		Unidades []string
	}{EditMode: editMode, Produto: produto, Unidades: unidadesProduto}
	app.render(w, r, status, "produto_form.html", data)
}

func (app *application) produtosList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterCodigo := r.URL.Query().Get("codigo")
	filterDescricao := r.URL.Query().Get("descricao")

	produtos, total, err := app.db.ListProdutosFilter(filterCodigo, filterDescricao, produtosPageSize, (page-1)*produtosPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "produtos"
	data.Data = struct {
		Produtos        []models.Produto
		Pagination      models.PaginationMetadata
		FilterCodigo    string
		FilterDescricao string
	}{
		Produtos: produtos, Pagination: app.calculatePagination(total, page, produtosPageSize),
		FilterCodigo: filterCodigo, FilterDescricao: filterDescricao,
	}
	app.render(w, r, http.StatusOK, "produtos.html", data)
}

func (app *application) produtoNew(w http.ResponseWriter, r *http.Request) {
	app.renderProdutoForm(w, r, http.StatusOK, "", false, &models.Produto{Unidade: "UN", ControlaEstoque: true})
}

func (app *application) produtoNewPost(w http.ResponseWriter, r *http.Request) {
	var form produtoForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	produto := form.toProduto("")
	if msg := validateProdutoForm(&form, true); msg != "" {
		app.renderProdutoForm(w, r, http.StatusUnprocessableEntity, msg, false, &produto)
		return
	}
	if err := app.db.CreateProduto(produto); err != nil {
		flash := "Erro ao criar produto (código já existe?)"
		app.renderProdutoForm(w, r, http.StatusUnprocessableEntity, flash, false, &produto)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Produto criado com sucesso.")
	http.Redirect(w, r, "/cadastros/produtos", http.StatusSeeOther)
}

func (app *application) produtoEdit(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	produto, err := app.db.GetProduto(codigo)
	if err != nil {
		app.notFound(w)
		return
	}
	app.renderProdutoForm(w, r, http.StatusOK, "", true, produto)
}

func (app *application) produtoEditPost(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	var form produtoForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	produto := form.toProduto(codigo)
	if msg := validateProdutoForm(&form, false); msg != "" {
		app.renderProdutoForm(w, r, http.StatusUnprocessableEntity, msg, true, &produto)
		return
	}
	if err := app.db.UpdateProduto(produto); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Produto atualizado com sucesso.")
	http.Redirect(w, r, "/cadastros/produtos", http.StatusSeeOther)
}

func (app *application) produtoDelete(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	if err := app.db.DeleteProduto(codigo); err != nil {
		if isFKError(err) {
			app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir: produto está vinculado a outros registros.")
			http.Redirect(w, r, "/cadastros/produtos", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Produto excluído.")
	http.Redirect(w, r, "/cadastros/produtos", http.StatusSeeOther)
}
