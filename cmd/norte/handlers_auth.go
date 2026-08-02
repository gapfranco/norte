package main

import (
	"fmt"
	"net/http"
)

const minPasswordLen = 6

type changePasswordForm struct {
	SenhaAtual       string `form:"senha_atual"`
	Senha            string `form:"senha"`
	SenhaConfirmacao string `form:"senha_confirmacao"`
}

func (app *application) changePassword(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.ActiveMenu = "config"
	data.ActiveSubmenu = "senha"
	app.render(w, r, http.StatusOK, "change_password.html", data)
}

func (app *application) changePasswordPost(w http.ResponseWriter, r *http.Request) {
	var form changePasswordForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	renderChange := func(status int, flash string) {
		data := app.newTemplateData(r)
		data.ActiveMenu = "config"
		data.ActiveSubmenu = "senha"
		data.Flash = flash
		app.render(w, r, status, "change_password.html", data)
	}

	if form.SenhaAtual == "" || form.Senha == "" {
		renderChange(http.StatusUnprocessableEntity, "Preencha todos os campos.")
		return
	}
	if len(form.Senha) < minPasswordLen {
		renderChange(http.StatusUnprocessableEntity, fmt.Sprintf("A nova senha deve ter pelo menos %d caracteres.", minPasswordLen))
		return
	}
	if form.Senha != form.SenhaConfirmacao {
		renderChange(http.StatusUnprocessableEntity, "A confirmação de senha não confere.")
		return
	}

	userID := app.currentUserID(r)
	if userID == 0 {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	if err := app.db.CheckUserPassword(userID, form.SenhaAtual); err != nil {
		renderChange(http.StatusUnprocessableEntity, "Senha atual incorreta.")
		return
	}

	if err := app.db.UpdateUserPassword(userID, form.Senha); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())

	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Senha alterada com sucesso.")
	http.Redirect(w, r, "/config/senha", http.StatusSeeOther)
}
