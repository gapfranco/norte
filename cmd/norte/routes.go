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

	// Cadastros — Papéis
	protected.HandleFunc("GET /cadastros/papeis", app.papeisList)
	protected.HandleFunc("GET /cadastros/papeis/new", app.papelNew)
	protected.HandleFunc("POST /cadastros/papeis/new", app.papelNewPost)
	protected.HandleFunc("GET /cadastros/papeis/{codigo}/edit", app.papelEdit)
	protected.HandleFunc("POST /cadastros/papeis/{codigo}/edit", app.papelEditPost)
	protected.HandleFunc("POST /cadastros/papeis/{codigo}/delete", app.papelDelete)

	// Cadastros — Pessoas
	protected.HandleFunc("GET /cadastros/pessoas", app.pessoasList)
	protected.HandleFunc("GET /cadastros/pessoas/new", app.pessoaNew)
	protected.HandleFunc("POST /cadastros/pessoas/new", app.pessoaNewPost)
	protected.HandleFunc("GET /cadastros/pessoas/{id}/edit", app.pessoaEdit)
	protected.HandleFunc("POST /cadastros/pessoas/{id}/edit", app.pessoaEditPost)
	protected.HandleFunc("POST /cadastros/pessoas/{id}/delete", app.pessoaDelete)

	// Cadastros — Produtos
	protected.HandleFunc("GET /cadastros/produtos", app.produtosList)
	protected.HandleFunc("GET /cadastros/produtos/new", app.produtoNew)
	protected.HandleFunc("POST /cadastros/produtos/new", app.produtoNewPost)
	protected.HandleFunc("GET /cadastros/produtos/{codigo}/edit", app.produtoEdit)
	protected.HandleFunc("POST /cadastros/produtos/{codigo}/edit", app.produtoEditPost)
	protected.HandleFunc("POST /cadastros/produtos/{codigo}/delete", app.produtoDelete)

	// Cadastros — Bancos
	protected.HandleFunc("GET /cadastros/bancos", app.bancosList)
	protected.HandleFunc("GET /cadastros/bancos/new", app.bancoNew)
	protected.HandleFunc("POST /cadastros/bancos/new", app.bancoNewPost)
	protected.HandleFunc("GET /cadastros/bancos/{codigo}/edit", app.bancoEdit)
	protected.HandleFunc("POST /cadastros/bancos/{codigo}/edit", app.bancoEditPost)
	protected.HandleFunc("POST /cadastros/bancos/{codigo}/delete", app.bancoDelete)

	// Cadastros — Contas
	protected.HandleFunc("GET /cadastros/contas", app.contasList)
	protected.HandleFunc("GET /cadastros/contas/new", app.contaNew)
	protected.HandleFunc("POST /cadastros/contas/new", app.contaNewPost)
	protected.HandleFunc("GET /cadastros/contas/{banco}/{conta}/edit", app.contaEdit)
	protected.HandleFunc("POST /cadastros/contas/{banco}/{conta}/edit", app.contaEditPost)
	protected.HandleFunc("POST /cadastros/contas/{banco}/{conta}/delete", app.contaDelete)

	// Cadastros — Categorias
	protected.HandleFunc("GET /cadastros/categorias", app.categoriasList)
	protected.HandleFunc("GET /cadastros/categorias/new", app.categoriaNew)
	protected.HandleFunc("POST /cadastros/categorias/new", app.categoriaNewPost)
	protected.HandleFunc("GET /cadastros/categorias/{codigo}/edit", app.categoriaEdit)
	protected.HandleFunc("POST /cadastros/categorias/{codigo}/edit", app.categoriaEditPost)
	protected.HandleFunc("POST /cadastros/categorias/{codigo}/delete", app.categoriaDelete)

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
