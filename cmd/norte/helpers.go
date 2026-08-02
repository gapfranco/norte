package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"norte/internal/models"

	"github.com/justinas/nosurf"
)

// syncDB replica alterações locais ao Turso Cloud (modo sync) e puxa o delta remoto.
// No-op em modos local e remote. Em falha, registra Error e deixa flash para o usuário.
func (app *application) syncDB(ctx context.Context) {
	if err := app.db.Sync(ctx); err != nil {
		app.logger.Error("db sync failed", "error", err)
		app.sessionManager.Put(ctx, "flash_error",
			"Falha ao sincronizar com a nuvem. As alterações estão salvas localmente e serão reenviadas.")
	}
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) templateData {
	menu, submenu := menuFromPath(r.URL.Path)
	data := templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		FlashError:      app.sessionManager.PopString(r.Context(), "flash_error"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		ActiveMenu:      menu,
		ActiveSubmenu:   submenu,
	}
	if data.IsAuthenticated {
		data.CurrentUser = app.sessionManager.GetString(r.Context(), "authenticatedUser")
	}
	return data
}

func menuFromPath(path string) (menu, submenu string) {
	switch {
	case path == "/" || path == "":
		return "", ""
	case strings.HasPrefix(path, "/painel"):
		return "painel", ""
	case strings.HasPrefix(path, "/config/usuarios"):
		return "config", "usuarios"
	case strings.HasPrefix(path, "/config/senha"):
		return "config", "senha"
	default:
		return "", ""
	}
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	return app.formDecoder.Decode(dst, r.PostForm)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (app *application) currentUserID(r *http.Request) int {
	usuario := app.sessionManager.GetString(r.Context(), "authenticatedUser")
	if usuario == "" {
		return 0
	}
	user, err := app.db.GetUserByUsuario(usuario)
	if err != nil {
		return 0
	}
	return user.ID
}

func (app *application) calculatePagination(totalItems, currentPage, pageSize int) models.PaginationMetadata {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}
	return models.PaginationMetadata{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasPrev:     currentPage > 1,
		HasNext:     currentPage < totalPages,
		PrevPage:    currentPage - 1,
		NextPage:    currentPage + 1,
		Pages:       buildPaginationPages(currentPage, totalPages),
	}
}

func buildPaginationPages(currentPage, totalPages int) []models.PaginationPageItem {
	if totalPages < 1 {
		totalPages = 1
	}
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	include := make(map[int]bool, 7)
	include[1] = true
	include[totalPages] = true
	for p := currentPage - 2; p <= currentPage+2; p++ {
		if p >= 1 && p <= totalPages {
			include[p] = true
		}
	}

	pages := make([]int, 0, len(include))
	for p := 1; p <= totalPages; p++ {
		if include[p] {
			pages = append(pages, p)
		}
	}

	items := make([]models.PaginationPageItem, 0, len(pages)*2)
	prev := 0
	for _, p := range pages {
		if prev > 0 && p-prev > 1 {
			items = append(items, models.PaginationPageItem{Ellipsis: true})
		}
		items = append(items, models.PaginationPageItem{
			Page:    p,
			Current: p == currentPage,
		})
		prev = p
	}
	return items
}

// openBrowser identifica o SO e abre o navegador padrão
func openBrowser(url string) {
	var err error

	time.Sleep(100 * time.Millisecond)

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("sistema operacional não suportado")
	}

	if err != nil {
		fmt.Printf("Erro ao abrir o navegador: %v\n", err)
	}
}
