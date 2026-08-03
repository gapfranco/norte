# Modelagem — Norte

DDL magro e decisões de schema. Detalha a diretriz do [PRD §9](prd.md).

**Atualizado:** Agosto/2026

---

## Estratégia de IDs

| Estilo | Uso | Exemplos |
|--------|-----|----------|
| `INTEGER` autoincrement | Entidades surrogate (FKs futuras densas) | `users.id`, `pessoas.id` |
| `TEXT codigo` | Chave natural estável | `bancos`, `categorias`, `papeis`, `produtos`, `negocios` |
| Composite TEXT | Escopo composto | `unidades(negocio_codigo, codigo)`, `contas_bancarias(banco, conta)` |

FKs para pessoa usam `pessoa_id INTEGER`. Não há soft-delete nem flag `ativo` nos cadastros atuais — exclusão é hard delete.

---

## Papéis e Pessoas

### `papeis`

| Coluna | Tipo | Notas |
|--------|------|--------|
| `codigo` | TEXT PK | Ex.: `cliente` |
| `nome` | TEXT NOT NULL | Label exibido |

Seed: `cliente`, `fornecedor`, `vendedor`, `representante`. Novos papéis via CRUD.

### `pessoas`

| Coluna | Tipo | Notas |
|--------|------|--------|
| `id` | INTEGER PK | Autoincrement |
| `tipo_pessoa` | TEXT | `F` física / `J` jurídica |
| `nome` | TEXT NOT NULL | Nome ou razão social |
| `cpf` / `cnpj` | TEXT | Opcionais; unique parcial |
| contato | `email`, `telefone`, `celular` | |
| endereço | `cep`, `logradouro`, `numero`, `complemento`, `bairro`, `cidade`, `uf` | Magro; fiscal completo na Fase 9 |
| `observacoes` | TEXT | |
| `criado_em` / `atualizado_em` | TEXT | |

### `pessoa_papeis`

| Coluna | Tipo | Notas |
|--------|------|--------|
| `pessoa_id` | INTEGER FK → `pessoas(id)` ON DELETE CASCADE | |
| `papel_codigo` | TEXT FK → `papeis(codigo)` | |
| PK | `(pessoa_id, papel_codigo)` | |

Uma pessoa pode ter vários papéis. Pelo menos um papel é obrigatório na UI.

---

## Produtos

### `produtos`

| Coluna | Tipo | Notas |
|--------|------|--------|
| `codigo` | TEXT PK | SKU |
| `descricao` | TEXT NOT NULL | |
| `unidade` | TEXT NOT NULL DEFAULT `'UN'` | |
| `preco_venda` / `preco_custo` | REAL NOT NULL DEFAULT 0 | |
| `controla_estoque` | INTEGER NOT NULL DEFAULT 1 | 0/1 |
| `estoque_minimo` | REAL NOT NULL DEFAULT 0 | |
| `observacoes` | TEXT | |
| `criado_em` / `atualizado_em` | TEXT | |

Saldo e movimentos ficam em `estoque_saldos` / `estoque_movimentos` (Fase 2). NCM/origem na Fase 9.

---

## Já existentes (resumo)

- `users` — auth  
- `negocios` / `unidades` — contexto organizacional (CNPJ opcional)  
- `bancos` / `contas_bancarias` — contas financeiras  
- `categorias` — classificação E/S (+ classe em saídas)

---

## Fases posteriores (orientação)

| Entidade | Fase |
|----------|------|
| `estoque_movimentos` / `estoque_saldos` | 2 |
| `pedidos_compra` / itens | 3 |
| `pedidos_venda` / itens | 4 |
| `titulos` / `titulo_baixas` | 3–5 |
| `movimentacao_banco` | 6 |
| `cobrancas` | 8 |
| `documentos_fiscais` | 9 |
| `orcamentos` | 10+ |

Regras duras (PRD): confirmação compra/venda é transação única; estoque só com movimento rastreável; boleto/NF-e não alteram estoque/título.
