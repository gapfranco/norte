package main

import (
	"net/http"
)

// ─── Setup ───────────────────────────────────────────────────────────────────

type setupForm struct {
	Usuario string `form:"usuario"`
	Nome    string `form:"nome"`
	Senha   string `form:"senha"`
}

func (app *application) setup(w http.ResponseWriter, r *http.Request) {
	if app.setupDone.Load() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	app.render(w, r, http.StatusOK, "setup.html", app.newTemplateData(r))
}

func (app *application) setupPost(w http.ResponseWriter, r *http.Request) {
	if app.setupDone.Load() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var form setupForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if form.Usuario == "" || form.Senha == "" || form.Nome == "" {
		data := app.newTemplateData(r)
		data.Flash = "Todos os campos são obrigatórios"
		app.render(w, r, http.StatusUnprocessableEntity, "setup.html", data)
		return
	}

	email := normalizeEmail(form.Usuario)
	if !isValidEmail(email) {
		data := app.newTemplateData(r)
		data.Flash = "Informe um e-mail válido no campo E-mail."
		app.render(w, r, http.StatusUnprocessableEntity, "setup.html", data)
		return
	}

	if err := app.db.CreateUser(email, form.Nome, form.Senha); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())

	app.setupDone.Store(true)
	app.sessionManager.Put(r.Context(), "flash", "Usuário criado com sucesso! Faça login.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

type userLoginForm struct {
	Usuario string `form:"usuario"`
	Senha   string `form:"senha"`
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Data = struct{ NomeEmpresa string }{NomeEmpresa: app.config.NomeEmpresa}
	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *application) loginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user, err := app.db.Authenticate(normalizeEmail(form.Usuario), form.Senha)
	if err != nil {
		data := app.newTemplateData(r)
		data.Flash = "E-mail ou senha inválidos"
		data.Data = struct{ NomeEmpresa string }{NomeEmpresa: app.config.NomeEmpresa}
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUser", user.Usuario)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logoutPost(w http.ResponseWriter, r *http.Request) {
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUser")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ─── Home ─────────────────────────────────────────────────────────────────────

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) painel(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "painel"
	app.render(w, r, http.StatusOK, "painel.html", data)
}
