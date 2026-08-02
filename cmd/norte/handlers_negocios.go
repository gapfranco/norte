package main

import (
	"net/http"
	"strconv"
	"strings"

	"norte/internal/models"
)

const negociosPageSize = 10

type negocioForm struct {
	Nome         string `form:"nome"`
	NomeFantasia string `form:"nome_fantasia"`
	Documento    string `form:"documento"`
	Ativo        bool   `form:"ativo"`
}

type unidadeForm struct {
	Nome   string `form:"nome"`
	Codigo string `form:"codigo"`
	Ativo  bool   `form:"ativo"`
}

// ─── Negócios ───────────────────────────────────────────────────────────────

func (app *application) negociosList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterNome := r.URL.Query().Get("nome")

	negocios, total, err := app.db.ListNegociosFilter(filterNome, negociosPageSize, (page-1)*negociosPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	data.Data = struct {
		Negocios   []models.Negocio
		Pagination models.PaginationMetadata
		FilterNome string
	}{
		Negocios:   negocios,
		Pagination: app.calculatePagination(total, page, negociosPageSize),
		FilterNome: filterNome,
	}
	app.render(w, r, http.StatusOK, "negocios.html", data)
}

func (app *application) negocioNew(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	data.Data = struct {
		EditMode bool
		Negocio  *models.Negocio
	}{EditMode: false, Negocio: &models.Negocio{Ativo: true}}
	app.render(w, r, http.StatusOK, "negocio_form.html", data)
}

func (app *application) negocioNewPost(w http.ResponseWriter, r *http.Request) {
	var form negocioForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Nome = strings.TrimSpace(form.Nome)
	if form.Nome == "" {
		app.renderNegocioForm(w, r, false, &models.Negocio{
			Nome: form.Nome, NomeFantasia: form.NomeFantasia, Documento: form.Documento, Ativo: form.Ativo,
		}, "O nome do negócio é obrigatório.")
		return
	}
	if err := app.db.CreateNegocio(form.Nome, strings.TrimSpace(form.NomeFantasia), strings.TrimSpace(form.Documento), form.Ativo); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Negócio criado com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios", http.StatusSeeOther)
}

func (app *application) negocioEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	negocio, err := app.db.GetNegocio(id)
	if err != nil {
		app.notFound(w)
		return
	}
	app.renderNegocioForm(w, r, true, negocio, "")
}

func (app *application) negocioEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	if _, err := app.db.GetNegocio(id); err != nil {
		app.notFound(w)
		return
	}
	var form negocioForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Nome = strings.TrimSpace(form.Nome)
	if form.Nome == "" {
		app.renderNegocioForm(w, r, true, &models.Negocio{
			ID: id, Nome: form.Nome, NomeFantasia: form.NomeFantasia, Documento: form.Documento, Ativo: form.Ativo,
		}, "O nome do negócio é obrigatório.")
		return
	}
	if err := app.db.UpdateNegocio(id, form.Nome, strings.TrimSpace(form.NomeFantasia), strings.TrimSpace(form.Documento), form.Ativo); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Negócio atualizado com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios", http.StatusSeeOther)
}

func (app *application) negocioDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	if _, err := app.db.GetNegocio(id); err != nil {
		app.notFound(w)
		return
	}
	if err := app.db.DeleteNegocio(id); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Negócio excluído.")
	http.Redirect(w, r, "/cadastros/negocios", http.StatusSeeOther)
}

func (app *application) renderNegocioForm(w http.ResponseWriter, r *http.Request, editMode bool, negocio *models.Negocio, flash string) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	if flash != "" {
		data.Flash = flash
	}
	data.Data = struct {
		EditMode bool
		Negocio  *models.Negocio
	}{EditMode: editMode, Negocio: negocio}
	status := http.StatusOK
	if flash != "" {
		status = http.StatusUnprocessableEntity
	}
	app.render(w, r, status, "negocio_form.html", data)
}

// ─── Unidades (escopo de um Negócio) ──────────────────────────────────────────

func (app *application) unidadesList(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	unidades, err := app.db.ListUnidades(negocio.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	data.Data = struct {
		Negocio  *models.Negocio
		Unidades []models.Unidade
	}{Negocio: negocio, Unidades: unidades}
	app.render(w, r, http.StatusOK, "unidades.html", data)
}

func (app *application) unidadeNew(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	app.renderUnidadeForm(w, r, false, negocio, &models.Unidade{NegocioID: negocio.ID, Ativo: true}, "")
}

func (app *application) unidadeNewPost(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	var form unidadeForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Nome = strings.TrimSpace(form.Nome)
	if form.Nome == "" {
		app.renderUnidadeForm(w, r, false, negocio, &models.Unidade{
			NegocioID: negocio.ID, Nome: form.Nome, Codigo: form.Codigo, Ativo: form.Ativo,
		}, "O nome da unidade é obrigatório.")
		return
	}
	if err := app.db.CreateUnidade(negocio.ID, form.Nome, strings.TrimSpace(form.Codigo), form.Ativo); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Unidade criada com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios/"+strconv.Itoa(negocio.ID)+"/unidades", http.StatusSeeOther)
}

func (app *application) unidadeEdit(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	unidade, ok := app.unidadeFromPath(w, r, negocio.ID)
	if !ok {
		return
	}
	app.renderUnidadeForm(w, r, true, negocio, unidade, "")
}

func (app *application) unidadeEditPost(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	unidade, ok := app.unidadeFromPath(w, r, negocio.ID)
	if !ok {
		return
	}
	var form unidadeForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Nome = strings.TrimSpace(form.Nome)
	if form.Nome == "" {
		app.renderUnidadeForm(w, r, true, negocio, &models.Unidade{
			ID: unidade.ID, NegocioID: negocio.ID, Nome: form.Nome, Codigo: form.Codigo, Ativo: form.Ativo,
		}, "O nome da unidade é obrigatório.")
		return
	}
	if err := app.db.UpdateUnidade(unidade.ID, form.Nome, strings.TrimSpace(form.Codigo), form.Ativo); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Unidade atualizada com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios/"+strconv.Itoa(negocio.ID)+"/unidades", http.StatusSeeOther)
}

func (app *application) unidadeDelete(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	unidade, ok := app.unidadeFromPath(w, r, negocio.ID)
	if !ok {
		return
	}
	if err := app.db.DeleteUnidade(unidade.ID); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Unidade excluída.")
	http.Redirect(w, r, "/cadastros/negocios/"+strconv.Itoa(negocio.ID)+"/unidades", http.StatusSeeOther)
}

func (app *application) renderUnidadeForm(w http.ResponseWriter, r *http.Request, editMode bool, negocio *models.Negocio, unidade *models.Unidade, flash string) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	if flash != "" {
		data.Flash = flash
	}
	data.Data = struct {
		EditMode bool
		Negocio  *models.Negocio
		Unidade  *models.Unidade
	}{EditMode: editMode, Negocio: negocio, Unidade: unidade}
	status := http.StatusOK
	if flash != "" {
		status = http.StatusUnprocessableEntity
	}
	app.render(w, r, status, "unidade_form.html", data)
}

// negocioFromPath resolve o {id} da rota para um Negócio, respondendo 404 se inválido.
func (app *application) negocioFromPath(w http.ResponseWriter, r *http.Request) (*models.Negocio, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return nil, false
	}
	negocio, err := app.db.GetNegocio(id)
	if err != nil {
		app.notFound(w)
		return nil, false
	}
	return negocio, true
}

// unidadeFromPath resolve o {uid} da rota e valida que pertence ao negócio informado.
func (app *application) unidadeFromPath(w http.ResponseWriter, r *http.Request, negocioID int) (*models.Unidade, bool) {
	uid, err := strconv.Atoi(r.PathValue("uid"))
	if err != nil || uid < 1 {
		app.notFound(w)
		return nil, false
	}
	unidade, err := app.db.GetUnidade(uid)
	if err != nil || unidade.NegocioID != negocioID {
		app.notFound(w)
		return nil, false
	}
	return unidade, true
}
