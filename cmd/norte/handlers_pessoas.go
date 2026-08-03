package main

import (
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"norte/internal/models"
)

const pessoasPageSize = 10

type pessoaForm struct {
	TipoPessoa  string   `form:"tipo_pessoa"`
	Nome        string   `form:"nome"`
	CPF         string   `form:"cpf"`
	CNPJ        string   `form:"cnpj"`
	Email       string   `form:"email"`
	Telefone    string   `form:"telefone"`
	Celular     string   `form:"celular"`
	CEP         string   `form:"cep"`
	Logradouro  string   `form:"logradouro"`
	Numero      string   `form:"numero"`
	Complemento string   `form:"complemento"`
	Bairro      string   `form:"bairro"`
	Cidade      string   `form:"cidade"`
	UF          string   `form:"uf"`
	Observacoes string   `form:"observacoes"`
	Papeis      []string `form:"papeis"`
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (f *pessoaForm) toPessoa(id int) models.Pessoa {
	cpf, cnpj := onlyDigits(f.CPF), onlyDigits(f.CNPJ)
	if f.TipoPessoa == "F" {
		cnpj = ""
	} else if f.TipoPessoa == "J" {
		cpf = ""
	}
	return models.Pessoa{
		ID: id, TipoPessoa: f.TipoPessoa, Nome: strings.TrimSpace(f.Nome),
		CPF: cpf, CNPJ: cnpj,
		Email: strings.TrimSpace(f.Email), Telefone: strings.TrimSpace(f.Telefone),
		Celular: strings.TrimSpace(f.Celular), CEP: onlyDigits(f.CEP),
		Logradouro: strings.TrimSpace(f.Logradouro), Numero: strings.TrimSpace(f.Numero),
		Complemento: strings.TrimSpace(f.Complemento), Bairro: strings.TrimSpace(f.Bairro),
		Cidade: strings.TrimSpace(f.Cidade), UF: strings.ToUpper(strings.TrimSpace(f.UF)),
		Observacoes: strings.TrimSpace(f.Observacoes),
	}
}

func validatePessoaForm(f *pessoaForm) string {
	if f.TipoPessoa != "F" && f.TipoPessoa != "J" {
		return "Tipo de pessoa inválido"
	}
	if strings.TrimSpace(f.Nome) == "" {
		return "Nome é obrigatório"
	}
	if len(f.Papeis) == 0 {
		return "Selecione pelo menos um papel"
	}
	cpf, cnpj := onlyDigits(f.CPF), onlyDigits(f.CNPJ)
	if f.TipoPessoa == "F" && cpf != "" && len(cpf) != 11 {
		return "CPF deve ter 11 dígitos"
	}
	if f.TipoPessoa == "J" && cnpj != "" && len(cnpj) != 14 {
		return "CNPJ deve ter 14 dígitos"
	}
	uf := strings.ToUpper(strings.TrimSpace(f.UF))
	if uf != "" && len(uf) != 2 {
		return "UF deve ter 2 letras"
	}
	return ""
}

func (app *application) renderPessoaForm(w http.ResponseWriter, r *http.Request, status int, flash string, editMode bool, pessoa *models.Pessoa, selected []string) {
	papeis, err := app.db.ListPapeisAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	sel := map[string]bool{}
	for _, c := range selected {
		sel[c] = true
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "pessoas"
	data.Flash = flash
	data.Data = struct {
		EditMode       bool
		Pessoa         *models.Pessoa
		PapeisOpcoes   []models.Papel
		PapeisSelected map[string]bool
	}{EditMode: editMode, Pessoa: pessoa, PapeisOpcoes: papeis, PapeisSelected: sel}
	app.render(w, r, status, "pessoa_form.html", data)
}

func (app *application) pessoasList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	filterNome := r.URL.Query().Get("nome")
	filterDoc := r.URL.Query().Get("documento")
	filterPapel := r.URL.Query().Get("papel")

	pessoas, total, err := app.db.ListPessoasFilter(filterNome, filterDoc, filterPapel, pessoasPageSize, (page-1)*pessoasPageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	papeis, err := app.db.ListPapeisAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.ActiveMenu = "cadastros"
	data.ActiveSubmenu = "pessoas"
	data.Data = struct {
		Pessoas         []models.Pessoa
		Papeis          []models.Papel
		Pagination      models.PaginationMetadata
		FilterNome      string
		FilterDocumento string
		FilterPapel     string
	}{
		Pessoas: pessoas, Papeis: papeis,
		Pagination: app.calculatePagination(total, page, pessoasPageSize),
		FilterNome: filterNome, FilterDocumento: filterDoc, FilterPapel: filterPapel,
	}
	app.render(w, r, http.StatusOK, "pessoas.html", data)
}

func (app *application) pessoaNew(w http.ResponseWriter, r *http.Request) {
	app.renderPessoaForm(w, r, http.StatusOK, "", false, &models.Pessoa{TipoPessoa: "F"}, nil)
}

func (app *application) pessoaNewPost(w http.ResponseWriter, r *http.Request) {
	var form pessoaForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	pessoa := form.toPessoa(0)
	if msg := validatePessoaForm(&form); msg != "" {
		app.renderPessoaForm(w, r, http.StatusUnprocessableEntity, msg, false, &pessoa, form.Papeis)
		return
	}
	if _, err := app.db.CreatePessoa(pessoa, form.Papeis); err != nil {
		flash := "Erro ao criar pessoa"
		if isUniqueConstraintError(err) {
			flash = "CPF ou CNPJ já cadastrado"
		} else if isFKError(err) {
			flash = "Papel inválido"
		}
		app.renderPessoaForm(w, r, http.StatusUnprocessableEntity, flash, false, &pessoa, form.Papeis)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Pessoa criada com sucesso.")
	http.Redirect(w, r, "/cadastros/pessoas", http.StatusSeeOther)
}

func (app *application) pessoaEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	pessoa, err := app.db.GetPessoa(id)
	if err != nil {
		app.notFound(w)
		return
	}
	var selected []string
	for _, p := range pessoa.Papeis {
		selected = append(selected, p.Codigo)
	}
	app.renderPessoaForm(w, r, http.StatusOK, "", true, pessoa, selected)
}

func (app *application) pessoaEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	var form pessoaForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	pessoa := form.toPessoa(id)
	if msg := validatePessoaForm(&form); msg != "" {
		app.renderPessoaForm(w, r, http.StatusUnprocessableEntity, msg, true, &pessoa, form.Papeis)
		return
	}
	if err := app.db.UpdatePessoa(pessoa, form.Papeis); err != nil {
		flash := "Erro ao atualizar pessoa"
		if isUniqueConstraintError(err) {
			flash = "CPF ou CNPJ já cadastrado"
		} else if isFKError(err) {
			flash = "Papel inválido"
		}
		app.renderPessoaForm(w, r, http.StatusUnprocessableEntity, flash, true, &pessoa, form.Papeis)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Pessoa atualizada com sucesso.")
	http.Redirect(w, r, "/cadastros/pessoas", http.StatusSeeOther)
}

func (app *application) pessoaDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.notFound(w)
		return
	}
	if err := app.db.DeletePessoa(id); err != nil {
		if isFKError(err) {
			app.sessionManager.Put(r.Context(), "flash", "Não é possível excluir: pessoa está vinculada a outros registros.")
			http.Redirect(w, r, "/cadastros/pessoas", http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}
	app.syncDB(r.Context())
	app.sessionManager.Put(r.Context(), "flash", "Pessoa excluída.")
	http.Redirect(w, r, "/cadastros/pessoas", http.StatusSeeOther)
}
