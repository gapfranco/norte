package main

import (
	"net/http"
	"strconv"

	"norte/internal/models"
)

const contasPageSize = 10

var tiposContaBancaria = []string{"Corrente", "Poupança", "Investimento", "Caixa", "Outro"}

type contaBancariaForm struct {
	Banco         string `form:"banco"`
	Conta         string `form:"conta"`
	Agencia       string `form:"agencia"`
	Tipo          string `form:"tipo"`
	Descricao     string `form:"descricao"`
	NegocioCodigo string `form:"negocio_codigo"`
	UnidadeCodigo string `form:"unidade_codigo"`
}

type contaFormData struct {
	EditMode bool
	Conta    *models.ContaBancaria
	Bancos   []models.Banco
	Negocios []models.Negocio
	Unidades []models.Unidade
	Tipos    []string
}

func (app *application) loadContaFormData(conta *models.ContaBancaria, editMode bool) (contaFormData, error) {
	bancos, _, err := app.db.ListBancosFilter("", "", 1000, 0)
	if err != nil {
		return contaFormData{}, err
	}
	negocios, _, err := app.db.ListNegociosFilter("", "", "", 1000, 0)
	if err != nil {
		return contaFormData{}, err
	}
	unidades, err := app.db.ListAllUnidades()
	if err != nil {
		return contaFormData{}, err
	}
	return contaFormData{
		EditMode: editMode,
		Conta:    conta,
		Bancos:   bancos,
		Negocios: negocios,
		Unidades: unidades,
		Tipos:    tiposContaBancaria,
	}, nil
}

func (app *application) contasList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterBanco := r.URL.Query().Get("banco")
	filterDescricao := r.URL.Query().Get("descricao")

	contas, total, err := app.db.ListContasBancariasFilter(filterBanco, filterDescricao, contasPageSize, (page-1)*contasPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	bancos, _, err := app.db.ListBancosFilter("", "", 1000, 0)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "contas"
	data.Data = struct {
		Contas          []models.ContaBancaria
		Bancos          []models.Banco
		Pagination      models.PaginationMetadata
		FilterBanco     string
		FilterDescricao string
	}{
		Contas:          contas,
		Bancos:          bancos,
		Pagination:      app.calculatePagination(total, page, contasPageSize),
		FilterBanco:     filterBanco,
		FilterDescricao: filterDescricao,
	}
	app.render(w, r, http.StatusOK, "contas.html", data)
}

func (app *application) contaNew(w http.ResponseWriter, r *http.Request) {
	fd, err := app.loadContaFormData(nil, false)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "contas"
	data.Data = fd
	app.render(w, r, http.StatusOK, "conta_form.html", data)
}

func (app *application) contaNewPost(w http.ResponseWriter, r *http.Request) {
	var form contaBancariaForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	renderForm := func(flash string) {
		cb := &models.ContaBancaria{
			Banco: form.Banco, Conta: form.Conta, Agencia: form.Agencia,
			Tipo: form.Tipo, Descricao: form.Descricao,
			NegocioCodigo: form.NegocioCodigo, UnidadeCodigo: form.UnidadeCodigo,
		}
		fd, err := app.loadContaFormData(cb, false)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		d := app.newTemplateData(r)
		d.ActiveMenu = "cadastros"
		d.ActiveSubmenu = "contas"
		d.Flash = flash
		d.Data = fd
		app.render(w, r, http.StatusUnprocessableEntity, "conta_form.html", d)
	}

	switch {
	case form.Banco == "":
		renderForm("Banco é obrigatório")
		return
	case form.Conta == "":
		renderForm("Número da conta é obrigatório")
		return
	case form.Descricao == "":
		renderForm("Descrição é obrigatória")
		return
	}

	err := app.db.CreateContaBancaria(models.ContaBancaria{
		Banco: form.Banco, Conta: form.Conta, Agencia: form.Agencia,
		Tipo: form.Tipo, Descricao: form.Descricao,
		NegocioCodigo: form.NegocioCodigo, UnidadeCodigo: form.UnidadeCodigo,
	})
	if err != nil {
		renderForm("Erro ao criar conta (banco/conta já existe?)")
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Conta bancária criada com sucesso.")
	http.Redirect(w, r, "/cadastros/contas", http.StatusSeeOther)
}

func (app *application) contaEdit(w http.ResponseWriter, r *http.Request) {
	banco := r.PathValue("banco")
	conta := r.PathValue("conta")
	cb, err := app.db.GetContaBancaria(banco, conta)
	if err != nil {
		app.notFound(w)
		return
	}
	fd, err := app.loadContaFormData(cb, true)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "contas"
	data.Data = fd
	app.render(w, r, http.StatusOK, "conta_form.html", data)
}

func (app *application) contaEditPost(w http.ResponseWriter, r *http.Request) {
	banco := r.PathValue("banco")
	conta := r.PathValue("conta")
	var form contaBancariaForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Descricao == "" {
		cb, _ := app.db.GetContaBancaria(banco, conta)
		fd, err := app.loadContaFormData(cb, true)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		d := app.newTemplateData(r)
		d.ActiveMenu = "cadastros"
		d.ActiveSubmenu = "contas"
		d.Flash = "Descrição é obrigatória"
		d.Data = fd
		app.render(w, r, http.StatusUnprocessableEntity, "conta_form.html", d)
		return
	}
	if err := app.db.UpdateContaBancaria(models.ContaBancaria{
		Banco: banco, Conta: conta, Agencia: form.Agencia,
		Tipo: form.Tipo, Descricao: form.Descricao,
		NegocioCodigo: form.NegocioCodigo, UnidadeCodigo: form.UnidadeCodigo,
	}); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Conta bancária atualizada com sucesso.")
	http.Redirect(w, r, "/cadastros/contas", http.StatusSeeOther)
}

func (app *application) contaDelete(w http.ResponseWriter, r *http.Request) {
	banco := r.PathValue("banco")
	conta := r.PathValue("conta")
	if err := app.db.DeleteContaBancaria(banco, conta); err != nil {
		if isFKError(err) {
			app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir: conta está vinculada a outros registros.")
			http.Redirect(w, r, "/cadastros/contas", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Conta bancária excluída.")
	http.Redirect(w, r, "/cadastros/contas", http.StatusSeeOther)
}
