package main

import (
	"net/http"
	"slices"
	"strconv"

	"norte/internal/models"
)

const categoriasPageSize = 10

var classesCategoria = []string{
	"Custo Fixo", "Custo Variável", "Aquisições", "Investimentos", "Perdas e Prejuízos", "Outros",
}

type categoriaForm struct {
	Codigo string `form:"codigo"`
	Tipo   string `form:"tipo"`
	Classe string `form:"classe"`
}

func (app *application) categoriasList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterCodigo := r.URL.Query().Get("codigo")
	filterTipo := r.URL.Query().Get("tipo")

	categorias, total, err := app.db.ListCategoriasFilter(filterCodigo, filterTipo, categoriasPageSize, (page-1)*categoriasPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "categorias"
	data.Data = struct {
		Categorias   []models.Categoria
		Pagination   models.PaginationMetadata
		FilterCodigo string
		FilterTipo   string
	}{
		Categorias:   categorias,
		Pagination:   app.calculatePagination(total, page, categoriasPageSize),
		FilterCodigo: filterCodigo,
		FilterTipo:   filterTipo,
	}
	app.render(w, r, http.StatusOK, "categorias.html", data)
}

func (app *application) categoriaNew(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "categorias"
	data.Data = struct {
		EditMode  bool
		Categoria *models.Categoria
		Classes   []string
	}{EditMode: false, Classes: classesCategoria}
	app.render(w, r, http.StatusOK, "categoria_form.html", data)
}

func applyCategoriaClasseRule(form *categoriaForm) string {
	if form.Tipo == "E" {
		form.Classe = ""
		return ""
	}
	if form.Tipo == "S" {
		if !slices.Contains(classesCategoria, form.Classe) {
			return "Classe é obrigatória para categorias de saída"
		}
	}
	return ""
}

func (app *application) categoriaNewPost(w http.ResponseWriter, r *http.Request) {
	var form categoriaForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	renderErr := func(flash string) {
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "categorias"
		data.Flash = flash
		data.Data = struct {
			EditMode  bool
			Categoria *models.Categoria
			Classes   []string
		}{
			EditMode:  false,
			Categoria: &models.Categoria{Codigo: form.Codigo, Tipo: form.Tipo, Classe: form.Classe},
			Classes:   classesCategoria,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "categoria_form.html", data)
	}
	if form.Codigo == "" || (form.Tipo != "E" && form.Tipo != "S") {
		renderErr("Código e tipo (E/S) são obrigatórios")
		return
	}
	if msg := applyCategoriaClasseRule(&form); msg != "" {
		renderErr(msg)
		return
	}
	if err := app.db.CreateCategoria(models.Categoria{
		Codigo: form.Codigo, Tipo: form.Tipo, Classe: form.Classe,
	}); err != nil {
		renderErr("Erro ao criar categoria (código já existe?)")
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Categoria criada com sucesso.")
	http.Redirect(w, r, "/cadastros/categorias", http.StatusSeeOther)
}

func (app *application) categoriaEdit(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	cat, err := app.db.GetCategoria(codigo)
	if err != nil {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "categorias"
	data.Data = struct {
		EditMode  bool
		Categoria *models.Categoria
		Classes   []string
	}{EditMode: true, Categoria: cat, Classes: classesCategoria}
	app.render(w, r, http.StatusOK, "categoria_form.html", data)
}

func (app *application) categoriaEditPost(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	var form categoriaForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Tipo != "E" && form.Tipo != "S" {
		cat, _ := app.db.GetCategoria(codigo)
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "categorias"
		data.Flash = "Tipo (E/S) é obrigatório"
		data.Data = struct {
			EditMode  bool
			Categoria *models.Categoria
			Classes   []string
		}{EditMode: true, Categoria: cat, Classes: classesCategoria}
		app.render(w, r, http.StatusUnprocessableEntity, "categoria_form.html", data)
		return
	}
	if msg := applyCategoriaClasseRule(&form); msg != "" {
		cat := &models.Categoria{Codigo: codigo, Tipo: form.Tipo, Classe: form.Classe}
		data := app.newTemplateData(r)
		data.ActiveMenu = "cadastros"
		data.ActiveSubmenu = "categorias"
		data.Flash = msg
		data.Data = struct {
			EditMode  bool
			Categoria *models.Categoria
			Classes   []string
		}{EditMode: true, Categoria: cat, Classes: classesCategoria}
		app.render(w, r, http.StatusUnprocessableEntity, "categoria_form.html", data)
		return
	}
	if err := app.db.UpdateCategoria(models.Categoria{
		Codigo: codigo, Tipo: form.Tipo, Classe: form.Classe,
	}); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Categoria atualizada com sucesso.")
	http.Redirect(w, r, "/cadastros/categorias", http.StatusSeeOther)
}

func (app *application) categoriaDelete(w http.ResponseWriter, r *http.Request) {
	codigo := r.PathValue("codigo")
	if err := app.db.DeleteCategoria(codigo); err != nil {
		if isFKError(err) {
			app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir: categoria está em uso.")
			http.Redirect(w, r, "/cadastros/categorias", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Categoria excluída.")
	http.Redirect(w, r, "/cadastros/categorias", http.StatusSeeOther)
}
