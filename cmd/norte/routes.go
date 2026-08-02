package main

import (
	"net/http"

	"norte/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(ui.Files))
	mux.Handle("GET /static/", fileServer)

	mux.HandleFunc("GET /manifest.webmanifest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		http.ServeFileFS(w, r, ui.Files, "static/manifest.webmanifest")
	})
	mux.HandleFunc("GET /sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Service-Worker-Allowed", "/")
		http.ServeFileFS(w, r, ui.Files, "static/sw.js")
	})

	standard := []func(http.Handler) http.Handler{
		app.recoverPanic,
		app.logRequest,
		commonHeaders,
	}

	dynamic := []func(http.Handler) http.Handler{
		app.sessionManager.LoadAndSave,
		preventCSRF,
		app.requireSetup,
		app.authenticate,
	}

	// Setup (primeiro acesso)
	mux.Handle("GET /setup", app.chain(http.HandlerFunc(app.setup), dynamic...))
	mux.Handle("POST /setup", app.chain(http.HandlerFunc(app.setupPost), dynamic...))

	// Rotas públicas
	mux.Handle("GET /login", app.chain(http.HandlerFunc(app.login), dynamic...))
	mux.Handle("POST /login", app.chain(http.HandlerFunc(app.loginPost), dynamic...))

	// Rotas protegidas
	protected := http.NewServeMux()
	protected.HandleFunc("GET /", app.home)
	protected.HandleFunc("GET /painel", app.painel)

	// Cadastros — Negócios e Unidades
	protected.HandleFunc("GET /cadastros/negocios", app.negociosList)
	protected.HandleFunc("GET /cadastros/negocios/new", app.negocioNew)
	protected.HandleFunc("POST /cadastros/negocios/new", app.negocioNewPost)
	protected.HandleFunc("GET /cadastros/negocios/{codigo}/edit", app.negocioEdit)
	protected.HandleFunc("POST /cadastros/negocios/{codigo}/edit", app.negocioEditPost)
	protected.HandleFunc("POST /cadastros/negocios/{codigo}/delete", app.negocioDelete)
	protected.HandleFunc("GET /cadastros/negocios/{codigo}/unidades/new", app.unidadeNew)
	protected.HandleFunc("POST /cadastros/negocios/{codigo}/unidades/new", app.unidadeNewPost)
	protected.HandleFunc("GET /cadastros/negocios/{codigo}/unidades/{ucodigo}/edit", app.unidadeEdit)
	protected.HandleFunc("POST /cadastros/negocios/{codigo}/unidades/{ucodigo}/edit", app.unidadeEditPost)
	protected.HandleFunc("POST /cadastros/negocios/{codigo}/unidades/{ucodigo}/delete", app.unidadeDelete)

	protected.HandleFunc("GET /config/usuarios", app.usuariosList)
	protected.HandleFunc("GET /config/usuarios/new", app.usuarioNew)
	protected.HandleFunc("POST /config/usuarios/new", app.usuarioNewPost)
	protected.HandleFunc("GET /config/usuarios/{id}/edit", app.usuarioEdit)
	protected.HandleFunc("POST /config/usuarios/{id}/edit", app.usuarioEditPost)
	protected.HandleFunc("POST /config/usuarios/{id}/delete", app.usuarioDelete)
	protected.HandleFunc("GET /config/senha", app.changePassword)
	protected.HandleFunc("POST /config/senha", app.changePasswordPost)
	protected.HandleFunc("POST /logout", app.logoutPost)

	mux.Handle("/", app.chain(protected, append(dynamic, app.requireAuthentication)...))

	return app.chain(mux, standard...)
}

func (app *application) chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
