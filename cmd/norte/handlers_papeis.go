package main

import (
	"net/http"
	"strconv"

	"norte/internal/models"
)

const papeisPageSize = 10

type papelForm struct {
	Codigo string `form:"codigo"`
	Nome   string `form:"nome"`
}

func (app *application) papeisList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterCodigo := r.URL.Query().Get("codigo")
	filterNome := r.URL.Query().Get("nome")

	papeis, total, err := app.db.ListPapeisFilter(filterCodigo, filterNome, papeisPageSize, (page-1)*papeisPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "papeis"
	data.Data = struct {
		Papeis       []models.Papel
		Pagination   models.PaginationMetadata
		FilterCodigo string
		FilterNome   string
	}{
		Papeis:       papeis,
		Pagination:   app.calculatePagination(total, page, papeisPageSize),
		FilterCodigo: filterCodigo,
		FilterNome:   filterNome,
	}
	app.render(w, r, http.StatusOK, "papeis.html", data)
}

func (app *application) papelNew(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "papeis"
	data.Data = struct {
		EditMode bool
		Papel    *models.Papel
	}{EditMode: false}
	app.render(w, r, http.StatusOK, "papel_form.html", data)
}

func (app *application) papelNewPost(w http.ResponseWriter, r *http.Request) {
	var form papelForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Codigo == "" || form.Nome == "" {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "papeis"
		data.Flash = "Código e nome são obrigatórios"
		data.Data = struct {
			EditMode bool
			Papel    *models.Papel
		}{EditMode: false, Papel: &models.Papel{Codigo: form.Codigo, Nome: form.Nome}}
		app.render(w, r, http.StatusUnprocessableEntity, "papel_form.html", data)
		return
	}
	if err := app.db.CreatePapel(models.Papel{Codigo: form.Codigo, Nome: form.Nome}); err != nil {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "papeis"
		data.Flash = "Erro ao criar papel (código já existe?)"
		data.Data = struct {
			EditMode bool
			Papel    *models.Papel
		}{EditMode: false, Papel: &models.Papel{Codigo: form.Codigo, Nome: form.Nome}}
		app.render(w, r, http.StatusUnprocessableEntity, "papel_form.html", data)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Papel criado com sucesso.")
	http.Redirect(w, r, "/cadastros/papeis", http.StatusSeeOther)
}

func (app *application) papelEdit(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	papel, err := app.db.GetPapel(codigo)
	if err != nil {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "papeis"
	data.Data = struct {
		EditMode bool
		Papel    *models.Papel
	}{EditMode: true, Papel: papel}
	app.render(w, r, http.StatusOK, "papel_form.html", data)
}

func (app *application) papelEditPost(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	var form papelForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Nome == "" {
		papel, _ := app.db.GetPapel(codigo)
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "papeis"
		data.Flash = "Nome é obrigatório"
		data.Data = struct {
			EditMode bool
			Papel    *models.Papel
		}{EditMode: true, Papel: papel}
		app.render(w, r, http.StatusUnprocessableEntity, "papel_form.html", data)
		return
	}
	if err := app.db.UpdatePapel(models.Papel{Codigo: codigo, Nome: form.Nome}); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Papel atualizado com sucesso.")
	http.Redirect(w, r, "/cadastros/papeis", http.StatusSeeOther)
}

func (app *application) papelDelete(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	if err := app.db.DeletePapel(codigo); err != nil {
		if isFKError(err) {
			app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir: papel está vinculado a pessoas.")
			http.Redirect(w, r, "/cadastros/papeis", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Papel excluído.")
	http.Redirect(w, r, "/cadastros/papeis", http.StatusSeeOther)
}
