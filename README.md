# Norte — Gestão Operacional

**Norte** é um ERP operacional mínimo para PMEs: cadastros, vendas, compras, estoque, contas a pagar/receber e conciliação bancária. Roda como app desktop (webview) ou servidor HTTP headless, com interface web leve (SSR + HTMX).

Produto greenfield inspirado na stack e UX do Norte-Finan. Especificação completa em `recursos/prd.md`.

## Estado atual (Fase 0)

- Setup de primeiro acesso
- Login / logout
- CRUD de usuários
- Alteração de senha
- Shell com menu (itens stub “Em breve”, exceto Usuários e Alterar senha)
- Modos de banco `local`, `remote` e `sync` (Turso / SQLite)

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.25+ (`net/http` stdlib) |
| Banco de dados | Turso / LibSQL / SQLite via `tursogo` e `libsql-client-go` |
| Migrações | Goose v3 (SQL embed, execução no startup) |
| Frontend | Go Templates + HTMX |
| CSS | Tailwind CSS v4 (`ui/input.css` → `ui/static/css/styles.css`) |
| Desktop | `github.com/webview/webview_go` |
| Segurança | CSRF (`nosurf`) + sessões (`alexedwards/scs`) + bcrypt |
| Configuração | Viper (arquivo `norte.conf`) |

## Estrutura do Projeto

```
norte/
├── cmd/norte/           # Entry point, handlers, rotas, middleware, desktop
├── config/              # Carregamento de configuração via Viper
├── internal/
│   ├── migrations/      # Migrações SQL + Goose
│   ├── models/          # Entidades de domínio
│   └── storage/         # Camada de dados — três modos de conexão
├── ui/
│   ├── html/            # Templates (base, pages/, partials/)
│   ├── static/          # CSS, JS e imagens
│   ├── efs.go           # embed.FS dos assets
│   └── input.css        # Fonte Tailwind
├── recursos/            # PRD e docs internas
├── AGENTS.md
├── Makefile
└── norte.conf           # Configuração local (não versionado)
```

## Pré-requisitos

| Requisito | Observação |
|-----------|------------|
| Go 1.25+ | Versão definida em `go.mod` |
| `tailwindcss` no PATH | CLI v4; usado pelo `make build` |
| Linux (desktop) | `libgtk-3-dev`, `libwebkit2gtk-4.1-dev` |

## Configuração

Crie `norte.conf` na raiz do projeto (não commitado):

```env
DB_URL=https://seu-banco.turso.io    # Obrigatório para remote e sync
DB_TOKEN=seu-token-aqui              # Obrigatório para remote e sync
DB_MODE=local                        # local | remote | sync (padrão: remote)
DB_LOCAL_PATH=local.db               # Caminho do SQLite local (padrão: local.db)
NOME_EMPRESA=Norte                   # Nome na tela de login (padrão: Norte)
ADDR=:4000                           # Endereço HTTP (padrão: :4000)
```

Para desenvolvimento local sem Turso Cloud, use `DB_MODE=local`.

### Modos de banco de dados (`DB_MODE`)

| Modo | Descrição | Requisitos |
|------|-----------|------------|
| `remote` | Conecta diretamente ao Turso Cloud | `DB_URL` + `DB_TOKEN` |
| `local` | SQLite puro local, sem rede | `DB_LOCAL_PATH` |
| `sync` | Réplica embutida: leituras locais, sync com Turso Cloud | `DB_URL` + `DB_TOKEN` + `DB_LOCAL_PATH` |

## Build e Execução

### Desktop (janela nativa)

```bash
make build
./build/norte
./build/norte -headless   # servidor + navegador, sem webview
```

### Servidor headless

```bash
make build-server
./build/norte-server
```

Acesse em: `http://localhost:4000`

### Outros

```bash
make tailwind-watch   # CSS em watch
make clean
go test ./...
```

## Documentação

| Documento | Conteúdo |
|-----------|----------|
| `recursos/prd.md` | Visão de produto, roadmap, fases |
| `recursos/menu.md` | Itens de menu (implementados vs. stubs) |
| `AGENTS.md` | Convenções para agentes de código |
