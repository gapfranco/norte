package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"norte/config"
	"norte/internal/migrations"
	"norte/internal/storage"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

// newAuthedServer sobe um servidor de teste com um usuário admin já autenticado.
func newAuthedServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := storage.NewTursoDB(storage.DBConfig{Mode: "local", LocalPath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := migrations.Run(db.DB()); err != nil {
		t.Fatal(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = time.Hour

	app := &application{
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    form.NewDecoder(),
		db:             db,
		config:         config.Config{NomeEmpresa: "Norte", Addr: ":0", DBMode: "local", DBLocalPath: dbPath},
	}

	hasUsers, err := db.HasUsers()
	if err != nil {
		t.Fatal(err)
	}
	app.setupDone.Store(hasUsers)

	srv := httptest.NewServer(app.routes())
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	_, body := mustGet(t, client, srv.URL+"/setup")
	token := mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/setup", url.Values{
		"csrf_token": {token}, "nome": {"Admin"}, "usuario": {"admin@norte.local"}, "senha": {"senha123"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("setup POST %d", resp.StatusCode)
	}

	_, body = mustGet(t, client, srv.URL+"/login")
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/login", url.Values{
		"csrf_token": {token}, "usuario": {"admin@norte.local"}, "senha": {"senha123"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login POST %d", resp.StatusCode)
	}

	return srv, client
}

// extractID retorna o primeiro grupo capturado por pattern no corpo HTML.
func extractID(t *testing.T, body, pattern string) string {
	t.Helper()
	m := regexp.MustCompile(pattern).FindStringSubmatch(body)
	if m == nil {
		return ""
	}
	return m[1]
}

func TestNegociosUnidadesCRUD(t *testing.T) {
	srv, client := newAuthedServer(t)

	// Lista inicial vazia.
	resp, body := mustGet(t, client, srv.URL+"/cadastros/negocios")
	if resp.StatusCode != 200 {
		t.Fatalf("negocios list %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Nenhum negócio cadastrado") {
		t.Fatal("esperava estado vazio de negócios")
	}

	// Cria negócio com CNPJ válido.
	_, body = mustGet(t, client, srv.URL+"/cadastros/negocios/new")
	token := mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/negocios/new", url.Values{
		"csrf_token": {token},
		"codigo":     {"ACME"},
		"nome":       {"Acme Comércio LTDA"},
		"cnpj":       {"11.444.777/0001-61"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria negócio %d", resp.StatusCode)
	}

	resp, body = mustGet(t, client, srv.URL+"/cadastros/negocios")
	if resp.StatusCode != 200 || !strings.Contains(body, "Acme Comércio LTDA") {
		t.Fatalf("negócio não apareceu na lista: %d", resp.StatusCode)
	}
	if !strings.Contains(body, "ACME") {
		t.Fatal("código do negócio não apareceu na lista")
	}

	// Edita negócio (unidades embutidas no form de edição).
	editURL := srv.URL + "/cadastros/negocios/ACME/edit"
	resp, body = mustGet(t, client, editURL)
	if resp.StatusCode != 200 {
		t.Fatalf("edit negócio %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Nenhuma unidade cadastrada") {
		t.Fatal("esperava seção de unidades vazia no form de edição")
	}
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, editURL, url.Values{
		"csrf_token": {token},
		"nome":       {"Acme Comércio S.A."},
		"cnpj":       {"11.444.777/0001-61"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("edita negócio %d", resp.StatusCode)
	}
	resp, body = mustGet(t, client, srv.URL+"/cadastros/negocios")
	if !strings.Contains(body, "Acme Comércio S.A.") {
		t.Fatal("edição do negócio não refletiu na lista")
	}

	// Cria unidade a partir do formulário de edição.
	unidadeNewURL := srv.URL + "/cadastros/negocios/ACME/unidades/new"
	_, body = mustGet(t, client, unidadeNewURL)
	token = mustCSRF(t, body)
	if resp := mustPost(t, client, unidadeNewURL, url.Values{
		"csrf_token": {token},
		"codigo":     {"matriz"},
		"nome":       {"Matriz São Paulo"},
		"cnpj":       {"04.252.011/0001-10"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria unidade %d", resp.StatusCode)
	}

	resp, body = mustGet(t, client, editURL)
	if !strings.Contains(body, "Matriz São Paulo") {
		t.Fatal("unidade não apareceu no form de edição do negócio")
	}
	if !strings.Contains(body, "04252011000110") {
		t.Fatal("CNPJ da unidade não apareceu na listagem embutida")
	}

	// Exclui unidade.
	unidadeEditPath := extractID(t, body, `/unidades/([^/"]+)/edit`)
	if unidadeEditPath == "" {
		t.Fatal("não encontrou link de edição da unidade")
	}
	if resp := mustPost(t, client, srv.URL+"/cadastros/negocios/ACME/unidades/"+unidadeEditPath+"/delete", url.Values{"csrf_token": {token}}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("exclui unidade %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, editURL)
	if strings.Contains(body, "Matriz São Paulo") {
		t.Fatal("unidade não foi excluída")
	}

	// Exclui negócio.
	if resp := mustPost(t, client, srv.URL+"/cadastros/negocios/ACME/delete", url.Values{"csrf_token": {token}}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("exclui negócio %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/negocios")
	if strings.Contains(body, "Acme Comércio S.A.") {
		t.Fatal("negócio não foi excluído")
	}
}

func TestNegocioCNPJInvalido(t *testing.T) {
	srv, client := newAuthedServer(t)

	_, body := mustGet(t, client, srv.URL+"/cadastros/negocios/new")
	token := mustCSRF(t, body)
	resp := mustPost(t, client, srv.URL+"/cadastros/negocios/new", url.Values{
		"csrf_token": {token},
		"codigo":     {"BAD"},
		"nome":       {"Negócio Inválido"},
		"cnpj":       {"12.345.678/0001-90"},
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422 para CNPJ inválido, got %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/negocios")
	if strings.Contains(body, "Negócio Inválido") {
		t.Fatal("negócio com CNPJ inválido não deveria ser persistido")
	}
}

func TestUnidadeCNPJInvalido(t *testing.T) {
	srv, client := newAuthedServer(t)

	_, body := mustGet(t, client, srv.URL+"/cadastros/negocios/new")
	token := mustCSRF(t, body)
	if resp := mustPost(t, client, srv.URL+"/cadastros/negocios/new", url.Values{
		"csrf_token": {token},
		"codigo":     {"ACME"},
		"nome":       {"Acme"},
	}); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cria negócio %d", resp.StatusCode)
	}

	unidadeNewURL := srv.URL + "/cadastros/negocios/ACME/unidades/new"
	_, body = mustGet(t, client, unidadeNewURL)
	token = mustCSRF(t, body)
	resp := mustPost(t, client, unidadeNewURL, url.Values{
		"csrf_token": {token},
		"codigo":     {"u1"},
		"nome":       {"Unidade Inválida"},
		"cnpj":       {"12.345.678/0001-90"},
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422 para CNPJ inválido na unidade, got %d", resp.StatusCode)
	}
	_, body = mustGet(t, client, srv.URL+"/cadastros/negocios/ACME/edit")
	if strings.Contains(body, "Unidade Inválida") {
		t.Fatal("unidade com CNPJ inválido não deveria ser persistida")
	}
}
