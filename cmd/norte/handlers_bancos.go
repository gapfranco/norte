package main

import (
	"net/http"
	"strconv"

	"norte/internal/models"
)

const bancosPageSize = 10

type bancoForm struct {
	Codigo string `form:"codigo"`
	Nome   string `form:"nome"`
}

func (app *application) bancosList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterCodigo := r.URL.Query().Get("codigo")
	filterNome := r.URL.Query().Get("nome")

	bancos, total, err := app.db.ListBancosFilter(filterCodigo, filterNome, bancosPageSize, (page-1)*bancosPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "bancos"
	data.Data = struct {
		Bancos       []models.Banco
		Pagination   models.PaginationMetadata
		FilterCodigo string
		FilterNome   string
	}{
		Bancos:       bancos,
		Pagination:   app.calculatePagination(total, page, bancosPageSize),
		FilterCodigo: filterCodigo,
		FilterNome:   filterNome,
	}
	app.render(w, r, http.StatusOK, "bancos.html", data)
}

func (app *application) bancoNew(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "bancos"
	data.Data = struct {
		EditMode bool
		Banco    *models.Banco
	}{EditMode: false}
	app.render(w, r, http.StatusOK, "banco_form.html", data)
}

func (app *application) bancoNewPost(w http.ResponseWriter, r *http.Request) {
	var form bancoForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Codigo == "" || form.Nome == "" {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "bancos"
		data.Flash = "Código e nome são obrigatórios"
		data.Data = struct {
			EditMode bool
			Banco    *models.Banco
		}{EditMode: false, Banco: &models.Banco{Codigo: form.Codigo, Nome: form.Nome}}
		app.render(w, r, http.StatusUnprocessableEntity, "banco_form.html", data)
		return
	}
	if err := app.db.CreateBanco(models.Banco{Codigo: form.Codigo, Nome: form.Nome}); err != nil {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "bancos"
		data.Flash = "Erro ao criar banco (código já existe?)"
		data.Data = struct {
			EditMode bool
			Banco    *models.Banco
		}{EditMode: false, Banco: &models.Banco{Codigo: form.Codigo, Nome: form.Nome}}
		app.render(w, r, http.StatusUnprocessableEntity, "banco_form.html", data)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Banco criado com sucesso.")
	http.Redirect(w, r, "/cadastros/bancos", http.StatusSeeOther)
}

func (app *application) bancoEdit(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	banco, err := app.db.GetBanco(codigo)
	if err != nil {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "bancos"
	data.Data = struct {
		EditMode bool
		Banco    *models.Banco
	}{EditMode: true, Banco: banco}
	app.render(w, r, http.StatusOK, "banco_form.html", data)
}

func (app *application) bancoEditPost(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	var form bancoForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Nome == "" {
		banco, _ := app.db.GetBanco(codigo)
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "bancos"
		data.Flash = "Nome é obrigatório"
		data.Data = struct {
			EditMode bool
			Banco    *models.Banco
		}{EditMode: true, Banco: banco}
		app.render(w, r, http.StatusUnprocessableEntity, "banco_form.html", data)
		return
	}
	if err := app.db.UpdateBanco(models.Banco{Codigo: codigo, Nome: form.Nome}); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Banco atualizado com sucesso.")
	http.Redirect(w, r, "/cadastros/bancos", http.StatusSeeOther)
}

func (app *application) bancoDelete(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	if err := app.db.DeleteBanco(codigo); err != nil {
		if isFKError(err) {
			app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir: banco está vinculado a contas bancárias.")
			http.Redirect(w, r, "/cadastros/bancos", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Banco excluído.")
	http.Redirect(w, r, "/cadastros/bancos", http.StatusSeeOther)
}
