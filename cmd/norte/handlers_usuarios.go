package main

import (
	"net/http"
	"strconv"

	"norte/internal/models"
)

const usuariosPageSize = 10

type usuarioForm struct {
	Usuario string `form:"usuario"`
	Nome    string `form:"nome"`
	Senha   string `form:"senha"`
}

type usuarioEditForm struct {
	Nome string `form:"nome"`
}

func (app *application) usuariosList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterUsuario := r.URL.Query().Get("usuario")
	filterNome := r.URL.Query().Get("nome")

	users, total, err := app.db.ListUsersFilter(filterUsuario, filterNome, usuariosPageSize, (page-1)*usuariosPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "config"
	data.Data = struct {
		Users         []models.User
		Pagination    models.PaginationMetadata
		CurrentUser   string
		FilterUsuario string
		FilterNome    string
	}{
		Users:         users,
		Pagination:    app.calculatePagination(total, page, usuariosPageSize),
		CurrentUser:   app.sessionManager.GetString(r.Context(), "authenticatedUser"),
		FilterUsuario: filterUsuario,
		FilterNome:    filterNome,
	}
	app.render(w, r, http.StatusOK, "usuarios.html", data)
}

func (app *application) usuarioNew(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "config"
	data.Data = struct {
		EditMode bool
		User     *models.User
	}{EditMode: false}
	app.render(w, r, http.StatusOK, "usuario_form.html", data)
}

func (app *application) usuarioNewPost(w http.ResponseWriter, r *http.Request) {
	var form usuarioForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Usuario == "" || form.Nome == "" || form.Senha == "" {
		data := app.newTemplateData(r)
		data.ActiveMenu = "config"
		data.Flash = "Todos os campos são obrigatórios"
		data.Data = struct {
			EditMode bool
			User     *models.User
		}{EditMode: false}
		app.render(w, r, http.StatusUnprocessableEntity, "usuario_form.html", data)
		return
	}
	email := normalizeEmail(form.Usuario)
	if !isValidEmail(email) {
		data := app.newTemplateData(r)
		data.ActiveMenu = "config"
		data.Flash = "Informe um e-mail válido."
		data.Data = struct {
			EditMode bool
			User     *models.User
		}{EditMode: false}
		app.render(w, r, http.StatusUnprocessableEntity, "usuario_form.html", data)
		return
	}
	if err := app.db.CreateUser(email, form.Nome, form.Senha); err != nil {
		data := app.newTemplateData(r)
		data.ActiveMenu = "config"
		data.Flash = "Erro ao criar usuário (e-mail já existe?)"
		data.Data = struct {
			EditMode bool
			User     *models.User
		}{EditMode: false}
		app.render(w, r, http.StatusUnprocessableEntity, "usuario_form.html", data)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Usuário criado com sucesso.")
	http.Redirect(w, r, "/config/usuarios", http.StatusSeeOther)
}

func (app *application) usuarioEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	user, err := app.db.GetUser(id)
	if err != nil {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "config"
	data.Data = struct {
		EditMode bool
		User     *models.User
	}{EditMode: true, User: user}
	app.render(w, r, http.StatusOK, "usuario_form.html", data)
}

func (app *application) usuarioEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	var form usuarioEditForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if form.Nome == "" {
		user, _ := app.db.GetUser(id)
		data := app.newTemplateData(r)
		data.ActiveMenu = "config"
		data.Flash = "Nome é obrigatório"
		data.Data = struct {
			EditMode bool
			User     *models.User
		}{EditMode: true, User: user}
		app.render(w, r, http.StatusUnprocessableEntity, "usuario_form.html", data)
		return
	}
	if err := app.db.UpdateUserNome(id, form.Nome); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Usuário atualizado com sucesso.")
	http.Redirect(w, r, "/config/usuarios", http.StatusSeeOther)
}

func (app *application) usuarioDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	user, err := app.db.GetUser(id)
	if err != nil {
		app.notFound(w)
		return
	}
	currentUser := app.sessionManager.GetString(r.Context(), "authenticatedUser")
	if user.Usuario == currentUser {
		app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir o usuário atual.")
		http.Redirect(w, r, "/config/usuarios", http.StatusSeeOther)
		return
	}
	if err := app.db.DeleteUser(id); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Usuário excluído.")
	http.Redirect(w, r, "/config/usuarios", http.StatusSeeOther)
}
