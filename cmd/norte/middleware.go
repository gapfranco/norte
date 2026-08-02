package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src fonts.gstatic.com; script-src 'self' 'unsafe-inline'")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info("received request", "ip", r.RemoteAddr, "proto", r.Proto, "method", r.Method, "uri", r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if pv := recover(); pv != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%v", pv))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requireSetup redireciona para /setup enquanto não houver usuários.
func (app *application) requireSetup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.setupDone.Load() {
			if r.URL.Path != "/setup" {
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func preventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
	})
	// nosurf v1.2 assume HTTPS por padrão; em HTTP local (desktop/dev) isso
	// quebra Origin/Referer. Detectar TLS real por request.
	csrfHandler.SetIsTLSFunc(func(r *http.Request) bool {
		return r.TLS != nil
	})
	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetString(r.Context(), "authenticatedUser")
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
