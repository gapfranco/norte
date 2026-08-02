package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPWARoutes(t *testing.T) {
	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	handler := app.routes()

	t.Run("manifest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		ct := rr.Header().Get("Content-Type")
		if !strings.Contains(ct, "application/manifest+json") {
			t.Fatalf("Content-Type = %q, want application/manifest+json", ct)
		}
		body := rr.Body.String()
		if !strings.Contains(body, `"name": "Norte"`) {
			t.Fatalf("manifest missing name: %s", body)
		}
		if !strings.Contains(body, `"display": "standalone"`) {
			t.Fatalf("manifest missing standalone display: %s", body)
		}
	})

	t.Run("service worker", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		ct := rr.Header().Get("Content-Type")
		if !strings.Contains(ct, "javascript") {
			t.Fatalf("Content-Type = %q, want javascript", ct)
		}
		if rr.Header().Get("Service-Worker-Allowed") != "/" {
			t.Fatalf("Service-Worker-Allowed = %q, want /", rr.Header().Get("Service-Worker-Allowed"))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "skipWaiting") {
			t.Fatalf("sw.js missing skipWaiting: %s", body)
		}
		if strings.Contains(body, "caches") || strings.Contains(body, "fetch") {
			t.Fatalf("sw.js must not use cache/fetch: %s", body)
		}
	})
}

func TestBaseTemplateIncludesPWA(t *testing.T) {
	cache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"home.html", "login.html"} {
		ts, ok := cache[name]
		if !ok {
			t.Fatalf("template %s not in cache", name)
		}
		buf := new(bytes.Buffer)
		data := templateData{
			CSRFToken: "test-token",
			Data:      struct{ NomeEmpresa string }{NomeEmpresa: "Norte"},
		}
		if err := ts.ExecuteTemplate(buf, "base", data); err != nil {
			t.Fatalf("execute %s base: %v", name, err)
		}
		html := buf.String()
		if !strings.Contains(html, `rel="manifest"`) {
			t.Fatalf("%s missing manifest link", name)
		}
		if !strings.Contains(html, "apple-touch-icon") {
			t.Fatalf("%s missing apple-touch-icon", name)
		}
		if !strings.Contains(html, "viewport-fit=cover") {
			t.Fatalf("%s missing viewport-fit=cover", name)
		}
		if !strings.Contains(html, "serviceWorker") {
			t.Fatalf("%s missing service worker registration", name)
		}
	}
}
