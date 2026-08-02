---
description:
alwaysApply: true
---

# AGENTS.md — Norte

Instruções para agentes de código trabalhando neste repositório.

## Visão geral

**Norte** é um ERP operacional mínimo em Go (cadastros, compras, vendas, estoque, títulos, conciliação). Roda como app desktop (webview) ou servidor HTTP headless.

Monolito Go com handlers por domínio, templates SSR + HTMX, banco Turso/LibSQL (modos `local`, `remote`, `sync`).

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.25+, `net/http` stdlib |
| Banco | Turso / LibSQL / SQLite (`internal/storage`) |
| Migrações | Goose v3 (`internal/migrations/*.sql`) |
| Frontend | Go Templates + HTMX |
| CSS | Tailwind CSS v4 (`ui/input.css` → `ui/static/css/styles.css`) |
| Desktop | `webview_go` (build tag `!headless`) |
| Segurança | nosurf (CSRF) + scs (sessões) + bcrypt |
| Config | Viper (`norte.conf`) |

## Estrutura do repositório

```
cmd/norte/         # Entry point, handlers, rotas, middleware, desktop
  handlers*.go     # Um arquivo por domínio
  routes.go        # Registro de rotas e cadeias de middleware
  helpers.go       # render, renderPartial, erros, decodePostForm
  desktop.go       # //go:build !headless
  desktop_stub.go  # //go:build headless
config/            # Carregamento de configuração
internal/
  migrations/      # SQL embed + goose.Up no startup
  models/          # Structs de domínio
  storage/         # TursoDB — toda persistência passa por aqui
ui/
  html/            # base.html, pages/, partials/
  static/          # CSS/JS embed via ui/efs.go
recursos/          # PRD, modelagem futura, docs internas
```

## Comandos essenciais

| Ação | Comando |
|------|---------|
| Build desktop | `make build` |
| Build servidor headless | `make build-server` |
| Executar desktop | `./build/norte` |
| Executar headless | `./build/norte -headless` ou `./build/norte-server` |
| Tailwind watch | `make tailwind-watch` |
| Testes | `go test ./...` |
| Limpar build | `make clean` |

**Pré-requisitos Linux (desktop):** `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, binário `tailwindcss` no PATH.

**Desenvolvimento local:** use `DB_MODE=local` em `norte.conf` para evitar dependência de Turso Cloud.

## Arquitetura e convenções

### Handlers

- Struct central `application` em `main.go` — injeta `db`, `sessionManager`, `formDecoder`, `templateCache`.
- Handlers são métodos `(app *application)` no pacote `main`.
- Novos domínios: criar `handlers_<dominio>.go` e registrar rotas em `routes.go`.
- Formulários: struct com tags `form:"campo"` + `app.decodePostForm(r, &form)`.
- Respostas HTML: `app.render()` (página completa) ou `app.renderPartial()` (HTMX).
- Erros: `app.serverError()` (500 + log) ou `app.clientError()` (4xx).

### Rotas e middleware

Cadeia padrão em `routes.go`:

- **standard:** `recoverPanic`, `logRequest`, `commonHeaders`
- **dynamic:** `sessionManager.LoadAndSave`, `preventCSRF`, `requireSetup`, `authenticate`
- Rotas protegidas ficam em sub-mux com `requireAuthentication`.

Rotas usam padrão Go 1.22+: `"GET /path"`, `"POST /config/usuarios/{id}/edit"`.

### Storage

- Toda query SQL fica em `internal/storage` (métodos do `TursoDB`).
- Não espalhar SQL nos handlers.
- Modo `sync`: chamar `db.Sync()` após escritas e antes de leituras críticas.
- IDs: `users.id` é INTEGER; novas entidades devem seguir a modelagem documentada.

### Migrações

- Arquivos numerados: `001_users.sql`, etc.
- Embutidos via `//go:embed` em `migrate.go`; executados no startup.
- Dialeto: `sqlite3`. Nunca alterar migrações já aplicadas — criar nova com número sequencial.

### Build tags

- **Desktop (padrão):** sem tags — compila `desktop.go` com webview.
- **Headless:** `-tags headless` — compila `desktop_stub.go`, sem CGo/webview.

### Idioma

- UI, mensagens de flash e comentários de negócio: **português**.
- Identificadores de código: manter consistência local.

## Frontend

- Templates em `ui/html/`; herdam de `base.html` via `{{template "base" .}}`.
- Partials HTMX em `ui/html/partials/`.
- Assets estáticos em `ui/static/`; servidos em `/static/`.
- Após mudanças em classes Tailwind: rodar `make tailwind-build`.
- Não adicionar framework JS — interatividade via HTMX apenas.

### Mobile e PWA

- **Listas:** dual `mobile-card-list` (`sm:hidden`) + `hidden sm:block table-scroll` / `list-table`.
- **Toolbars/filtros:** `page-toolbar`, `filter-bar`; CTAs com `w-full sm:w-auto`.
- **Forms:** `page-form*` + `form-actions` (botões full-width no mobile via CSS).
- **Shell:** `safe-main` / drawer já no `base.html` — não reinventar safe-area nem nav mobile.
- **PWA:** instalável (manifest + SW em `/`); ícones via `recursos/build-icons.sh`. Service worker **sem** cache offline (`caches`/`fetch` proibidos).

## Configuração

Arquivo `norte.conf` na raiz (não commitar tokens reais):

```env
DB_URL=
DB_TOKEN=
DB_MODE=local
DB_LOCAL_PATH=local.db
NOME_EMPRESA=Norte
ADDR=:4000
```

## Testes e verificação

Antes de concluir alterações:

1. `go test ./...`
2. `go build -tags headless ./cmd/norte` (ou `make build-server`)
3. Se alterou templates/CSS: `make tailwind-build`
4. Se alterou schema: nova migração SQL + método em `storage`

## Limites e cuidados

- **Não commitar** `norte.conf` com tokens, `.env` ou credenciais.
- **Não** introduzir frameworks web (Gin, Echo, etc.) — stdlib `net/http`.
- **Não** quebrar os três modos de banco (`local`, `remote`, `sync`).
- **Não** editar migrações antigas; sempre criar arquivo novo.
- **Não** criar commits ou PRs sem solicitação explícita do usuário.
- Escopo mínimo: mudanças focadas, sem refatorações não solicitadas.
- Schema e domínio próprios — não copiar schema do Norte-Finan sem modelagem.

## Estado do produto

**Fase 0** (fundação): auth + usuários + shell com menu stubs.

Consulte antes de implementar features novas:

| Documento | Conteúdo |
|-----------|----------|
| `recursos/prd.md` | Visão de produto, roadmap, fases |
| `recursos/menu.md` | Itens de menu planejados vs. implementados |
| `README.md` | Setup, build, modos de banco |

## Cursor Cloud specific instructions

Ambiente de desenvolvimento para agentes na nuvem. O update script roda `go mod download` a cada startup; as observações abaixo são caveats não óbvios.

- **Rode em modo headless.** A VM não tem GTK/WebKit; o build/execução desktop (webview, sem build tag) falha. Sempre use a tag `headless`: `go test -tags headless ./...`, `go build -tags headless ./cmd/norte` (ou `make build-server`). O binário resultante fica em `build/norte-server` e sobe em `:4000`.
- **`norte.conf` é obrigatório e não versionado** (está no `.gitignore`). O default de `DB_MODE` é `remote`, que exige Turso Cloud (`DB_URL`/`DB_TOKEN`) e faz o app abortar no startup sem credenciais. Para dev, mantenha um `norte.conf` na raiz com `DB_MODE=local` e `DB_LOCAL_PATH=local.db` (SQLite puro, sem rede). Este arquivo já existe na VM; recrie-o se sumir.
- **`tailwindcss` (CLI standalone v4) está instalado em `/usr/local/bin`** e não vem do `go.mod`. É necessário para gerar `ui/static/css/styles.css` (gitignored, mas embutido via `//go:embed "static"`). `make build-server`/`make build` chamam `tailwind-build` automaticamente. O `go build` funciona mesmo sem o CSS (o embed não falha), mas a UI fica sem estilos até rodar `make tailwind-build`.
- **Executar o servidor de dev:** `./build/norte-server` (ou `go run -tags headless ./cmd/norte -headless`). Não há hot reload; após alterar Go, rebuild. Após alterar classes Tailwind, rode `make tailwind-build`. No primeiro acesso (banco vazio) o app redireciona para `/setup` para criar o admin inicial antes do `/login`.
