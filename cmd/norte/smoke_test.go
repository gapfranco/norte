package main

import (
	"html"
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

func TestSmokeSetupLoginUsers(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := storage.NewTursoDB(storage.DBConfig{
		Mode:      "local",
		LocalPath: dbPath,
	})
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

	resp, body := mustGet(t, client, srv.URL+"/setup")
	if resp.StatusCode != 200 {
		t.Fatalf("setup GET %d", resp.StatusCode)
	}
	token := mustCSRF(t, body)

	resp = mustPost(t, client, srv.URL+"/setup", url.Values{
		"csrf_token": {token},
		"nome":       {"Admin"},
		"usuario":    {"admin@norte.local"},
		"senha":      {"senha123"},
	})
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("setup POST %d", resp.StatusCode)
	}

	resp, body = mustGet(t, client, srv.URL+"/login")
	if resp.StatusCode != 200 {
		t.Fatalf("login GET %d", resp.StatusCode)
	}
	token = mustCSRF(t, body)

	resp = mustPost(t, client, srv.URL+"/login", url.Values{
		"csrf_token": {token},
		"usuario":    {"admin@norte.local"},
		"senha":      {"senha123"},
	})
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login POST %d", resp.StatusCode)
	}

	client.CheckRedirect = nil

	resp, body = mustGet(t, client, srv.URL+"/")
	if resp.StatusCode != 200 {
		t.Fatalf("home %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Em breve") {
		t.Fatal("expected menu stubs")
	}
	if !strings.Contains(body, "/config/usuarios") {
		t.Fatal("expected users link")
	}
	if !strings.Contains(body, "/painel") {
		t.Fatal("expected painel link")
	}

	resp, body = mustGet(t, client, srv.URL+"/painel")
	if resp.StatusCode != 200 || !strings.Contains(body, "Painel") {
		t.Fatalf("painel failed: %d", resp.StatusCode)
	}

	resp, body = mustGet(t, client, srv.URL+"/config/usuarios")
	if resp.StatusCode != 200 || !strings.Contains(body, "admin@norte.local") {
		t.Fatalf("usuarios list failed: %d", resp.StatusCode)
	}
	if !strings.Contains(body, "mobile-card") {
		t.Fatal("expected mobile-card list on usuarios")
	}

	resp, body = mustGet(t, client, srv.URL+"/config/usuarios/new")
	token = mustCSRF(t, body)
	resp = mustPost(t, client, srv.URL+"/config/usuarios/new", url.Values{
		"csrf_token": {token},
		"usuario":    {"ops@norte.local"},
		"nome":       {"Operador"},
		"senha":      {"senha123"},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("create user %d", resp.StatusCode)
	}

	resp, body = mustGet(t, client, srv.URL+"/config/usuarios")
	if !strings.Contains(body, "ops@norte.local") {
		t.Fatal("expected ops user in list")
	}

	resp, body = mustGet(t, client, srv.URL+"/config/senha")
	token = mustCSRF(t, body)
	resp = mustPost(t, client, srv.URL+"/config/senha", url.Values{
		"csrf_token":        {token},
		"senha_atual":       {"senha123"},
		"senha":             {"senha456"},
		"senha_confirmacao": {"senha456"},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("change password %d", resp.StatusCode)
	}
}

func mustGet(t *testing.T, client *http.Client, u string) (*http.Response, string) {
	t.Helper()
	resp, err := client.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, string(b)
}

func mustPost(t *testing.T, client *http.Client, u string, vals url.Values) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(vals.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// nosurf v1.2 exige Origin ou Referer em métodos não seguros.
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	origin := parsed.Scheme + "://" + parsed.Host
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", origin+"/")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp
}

func mustCSRF(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`name=["']csrf_token["'] value=["']([^"']+)["']`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatal("csrf token not found")
	}
	return html.UnescapeString(m[1])
}
