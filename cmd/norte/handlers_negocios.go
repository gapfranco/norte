package main

import (
	"net/http"
	"strconv"
	"strings"

	"norte/internal/models"
)

const negociosPageSize = 10
const unidadesPageSize = 10

type negocioForm struct {
	Codigo string `form:"codigo"`
	Nome   string `form:"nome"`
	CNPJ   string `form:"cnpj"`
}

type unidadeForm struct {
	Codigo string `form:"codigo"`
	Nome   string `form:"nome"`
	CNPJ   string `form:"cnpj"`
}

// ─── Negócios ───────────────────────────────────────────────────────────────

func (app *application) negociosList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterCodigo := r.URL.Query().Get("codigo")
	filterNome := r.URL.Query().Get("nome")
	filterCNPJ := r.URL.Query().Get("cnpj")
	// CNPJ é persistido normalizado; o filtro ignora pontuação digitada.
	cnpjQuery := normalizarCNPJ(filterCNPJ)

	negocios, total, err := app.db.ListNegociosFilter(filterCodigo, filterNome, cnpjQuery, negociosPageSize, (page-1)*negociosPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	data.Data = struct {
		Negocios     []models.Negocio
		Pagination   models.PaginationMetadata
		FilterCodigo string
		FilterNome   string
		FilterCNPJ   string
	}{
		Negocios:     negocios,
		Pagination:   app.calculatePagination(total, page, negociosPageSize),
		FilterCodigo: filterCodigo,
		FilterNome:   filterNome,
		FilterCNPJ:   filterCNPJ,
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
	}{EditMode: false, Negocio: &models.Negocio{}}
	app.render(w, r, http.StatusOK, "negocio_form.html", data)
}

func (app *application) negocioNewPost(w http.ResponseWriter, r *http.Request) {
	var form negocioForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Codigo = strings.TrimSpace(form.Codigo)
	form.Nome = strings.TrimSpace(form.Nome)
	form.CNPJ = strings.TrimSpace(form.CNPJ)

	var flashMsg string
	switch {
	case form.Codigo == "" || form.Nome == "":
		flashMsg = "Código e nome são obrigatórios."
	case !validarCNPJ(form.CNPJ):
		flashMsg = "CNPJ inválido."
	}
	if flashMsg != "" {
		app.renderNegocioForm(w, r, false, &models.Negocio{
			Codigo: form.Codigo, Nome: form.Nome, CNPJ: form.CNPJ,
		}, nil, models.PaginationMetadata{}, flashMsg)
		return
	}

	cnpj := normalizarCNPJ(form.CNPJ)
	if err := app.db.CreateNegocio(models.Negocio{Codigo: form.Codigo, Nome: form.Nome, CNPJ: cnpj}); err != nil {
		msg := "Erro ao criar negócio (código já existe?)."
		if isCNPJUniqueError(err) {
			msg = "CNPJ já cadastrado."
		}
		app.renderNegocioForm(w, r, false, &models.Negocio{
			Codigo: form.Codigo, Nome: form.Nome, CNPJ: form.CNPJ,
		}, nil, models.PaginationMetadata{}, msg)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Negócio criado com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios", http.StatusSeeOther)
}

func (app *application) negocioEdit(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	negocio, err := app.db.GetNegocio(codigo)
	if err != nil {
		app.notFound(w)
		return
	}
	unidades, uniPag, err := app.loadUnidadesPage(codigo, r.URL.Query().Get("unidades_page"))
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.renderNegocioForm(w, r, true, negocio, unidades, uniPag, "")
}

func (app *application) negocioEditPost(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	if _, err := app.db.GetNegocio(codigo); err != nil {
		app.notFound(w)
		return
	}
	var form negocioForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Nome = strings.TrimSpace(form.Nome)
	form.CNPJ = strings.TrimSpace(form.CNPJ)

	var flashMsg string
	switch {
	case form.Nome == "":
		flashMsg = "O nome do negócio é obrigatório."
	case !validarCNPJ(form.CNPJ):
		flashMsg = "CNPJ inválido."
	}
	if flashMsg != "" {
		negocio := &models.Negocio{Codigo: codigo, Nome: form.Nome, CNPJ: form.CNPJ}
		unidades, uniPag, _ := app.loadUnidadesPage(codigo, "1")
		app.renderNegocioForm(w, r, true, negocio, unidades, uniPag, flashMsg)
		return
	}

	cnpj := normalizarCNPJ(form.CNPJ)
	if err := app.db.UpdateNegocio(models.Negocio{Codigo: codigo, Nome: form.Nome, CNPJ: cnpj}); err != nil {
		if isCNPJUniqueError(err) {
			negocio := &models.Negocio{Codigo: codigo, Nome: form.Nome, CNPJ: form.CNPJ}
			unidades, uniPag, _ := app.loadUnidadesPage(codigo, "1")
			app.renderNegocioForm(w, r, true, negocio, unidades, uniPag, "CNPJ já cadastrado.")
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Negócio atualizado com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios", http.StatusSeeOther)
}

func (app *application) negocioDelete(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	if _, err := app.db.GetNegocio(codigo); err != nil {
		app.notFound(w)
		return
	}
	if err := app.db.DeleteNegocio(codigo); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Negócio excluído.")
	http.Redirect(w, r, "/cadastros/negocios", http.StatusSeeOther)
}

func (app *application) renderNegocioForm(
	w http.ResponseWriter, r *http.Request,
	editMode bool, negocio *models.Negocio,
	unidades []models.Unidade, uniPag models.PaginationMetadata,
	flash string,
) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "negocios"
	if flash != "" {
		data.Flash = flash
	}
	data.Data = struct {
		EditMode      bool
		Negocio       *models.Negocio
		Unidades      []models.Unidade
		UniPagination models.PaginationMetadata
	}{
		EditMode:      editMode,
		Negocio:       negocio,
		Unidades:      unidades,
		UniPagination: uniPag,
	}
	status := http.StatusOK
	if flash != "" {
		status = http.StatusUnprocessableEntity
	}
	app.render(w, r, status, "negocio_form.html", data)
}

func (app *application) loadUnidadesPage(negocioCodigo, pageParam string) ([]models.Unidade, models.PaginationMetadata, error) {
	page := 1
	if pageParam != "" {
		if v, err := strconv.Atoi(pageParam); err == nil && v > 0 {
			page = v
		}
	}
	unidades, total, err := app.db.ListUnidades(negocioCodigo, unidadesPageSize, (page-1)*unidadesPageSize)
	if err != nil {
		return nil, models.PaginationMetadata{}, err
	}
	return unidades, app.calculatePagination(total, page, unidadesPageSize), nil
}

// ─── Unidades (escopo de um Negócio) ──────────────────────────────────────────

func (app *application) unidadeNew(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	app.renderUnidadeForm(w, r, false, negocio, &models.Unidade{NegocioCodigo: negocio.Codigo}, "")
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
	form.Codigo = strings.TrimSpace(form.Codigo)
	form.Nome = strings.TrimSpace(form.Nome)
	form.CNPJ = strings.TrimSpace(form.CNPJ)

	var flashMsg string
	switch {
	case form.Codigo == "" || form.Nome == "":
		flashMsg = "Código e nome são obrigatórios."
	case !validarCNPJ(form.CNPJ):
		flashMsg = "CNPJ inválido."
	}
	if flashMsg != "" {
		app.renderUnidadeForm(w, r, false, negocio, &models.Unidade{
			NegocioCodigo: negocio.Codigo, Codigo: form.Codigo, Nome: form.Nome, CNPJ: form.CNPJ,
		}, flashMsg)
		return
	}

	cnpj := normalizarCNPJ(form.CNPJ)
	if err := app.db.CreateUnidade(models.Unidade{
		NegocioCodigo: negocio.Codigo, Codigo: form.Codigo, Nome: form.Nome, CNPJ: cnpj,
	}); err != nil {
		msg := "Erro ao criar unidade (código já existe?)."
		if isCNPJUniqueError(err) {
			msg = "CNPJ já cadastrado."
		} else if isUniqueConstraintError(err) {
			msg = "Código da unidade já existe neste negócio."
		}
		app.renderUnidadeForm(w, r, false, negocio, &models.Unidade{
			NegocioCodigo: negocio.Codigo, Codigo: form.Codigo, Nome: form.Nome, CNPJ: form.CNPJ,
		}, msg)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Unidade criada com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios/"+negocio.Codigo+"/edit", http.StatusSeeOther)
}

func (app *application) unidadeEdit(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	unidade, ok := app.unidadeFromPath(w, r, negocio.Codigo)
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
	unidade, ok := app.unidadeFromPath(w, r, negocio.Codigo)
	if !ok {
		return
	}
	var form unidadeForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Nome = strings.TrimSpace(form.Nome)
	form.CNPJ = strings.TrimSpace(form.CNPJ)

	var flashMsg string
	switch {
	case form.Nome == "":
		flashMsg = "O nome da unidade é obrigatório."
	case !validarCNPJ(form.CNPJ):
		flashMsg = "CNPJ inválido."
	}
	if flashMsg != "" {
		app.renderUnidadeForm(w, r, true, negocio, &models.Unidade{
			NegocioCodigo: negocio.Codigo, Codigo: unidade.Codigo, Nome: form.Nome, CNPJ: form.CNPJ,
		}, flashMsg)
		return
	}

	cnpj := normalizarCNPJ(form.CNPJ)
	if err := app.db.UpdateUnidade(models.Unidade{
		NegocioCodigo: negocio.Codigo, Codigo: unidade.Codigo, Nome: form.Nome, CNPJ: cnpj,
	}); err != nil {
		if isCNPJUniqueError(err) {
			app.renderUnidadeForm(w, r, true, negocio, &models.Unidade{
				NegocioCodigo: negocio.Codigo, Codigo: unidade.Codigo, Nome: form.Nome, CNPJ: form.CNPJ,
			}, "CNPJ já cadastrado.")
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Unidade atualizada com sucesso.")
	http.Redirect(w, r, "/cadastros/negocios/"+negocio.Codigo+"/edit", http.StatusSeeOther)
}

func (app *application) unidadeDelete(w http.ResponseWriter, r *http.Request) {
	negocio, ok := app.negocioFromPath(w, r)
	if !ok {
		return
	}
	unidade, ok := app.unidadeFromPath(w, r, negocio.Codigo)
	if !ok {
		return
	}
	if err := app.db.DeleteUnidade(negocio.Codigo, unidade.Codigo); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Unidade excluída.")
	http.Redirect(w, r, "/cadastros/negocios/"+negocio.Codigo+"/edit", http.StatusSeeOther)
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

// negocioFromPath resolve o {codigo} da rota para um Negócio, respondendo 404 se inválido.
func (app *application) negocioFromPath(w http.ResponseWriter, r *http.Request) (*models.Negocio, bool) {
	codigo := strings.TrimSpace(r.PathValue("codigo"))
	if codigo == "" {
		app.notFound(w)
		return nil, false
	}
	negocio, err := app.db.GetNegocio(codigo)
	if err != nil {
		app.notFound(w)
		return nil, false
	}
	return negocio, true
}

// unidadeFromPath resolve o {ucodigo} da rota e valida que pertence ao negócio informado.
func (app *application) unidadeFromPath(w http.ResponseWriter, r *http.Request, negocioCodigo string) (*models.Unidade, bool) {
	ucodigo := strings.TrimSpace(r.PathValue("ucodigo"))
	if ucodigo == "" {
		app.notFound(w)
		return nil, false
	}
	unidade, err := app.db.GetUnidade(negocioCodigo, ucodigo)
	if err != nil {
		app.notFound(w)
		return nil, false
	}
	return unidade, true
}
