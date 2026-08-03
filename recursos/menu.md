# Menu de Opções — Norte

Espelho do menu atual da UI ([`ui/html/partials/nav.html`](../ui/html/partials/nav.html)).

**Fase:** 1 (cadastros em andamento)  
**Atualizado:** Agosto/2026

| Status | Significado |
|--------|-------------|
| **Implementado** | Item clicável com rota e tela |
| **Em breve** | Stub no dropdown/drawer (`nav-dropdown-item-disabled` / badge), sem rota |

Desktop e mobile exibem os mesmos itens.

---

## Cadastros

| Item | Status | Rota |
|------|--------|------|
| Negócios / Unidades | Implementado | `/cadastros/negocios` |
| Papéis | Implementado | `/cadastros/papeis` |
| Pessoas | Implementado | `/cadastros/pessoas` |
| Produtos | Implementado | `/cadastros/produtos` |
| Categorias | Implementado | `/cadastros/categorias` |
| Bancos | Implementado | `/cadastros/bancos` |
| Contas | Implementado | `/cadastros/contas` |

## Compras

| Item | Status |
|------|--------|
| Pedidos | Em breve |

## Vendas

| Item | Status |
|------|--------|
| Pedidos / Faturas | Em breve |

## Estoque

| Item | Status |
|------|--------|
| Saldos | Em breve |
| Movimentos | Em breve |
| Ajustes | Em breve |

## Financeiro

| Item | Status |
|------|--------|
| A Pagar | Em breve |
| A Receber | Em breve |
| Baixas | Em breve |
| Cobranças | Em breve (Fase 8) |

## Banco

| Item | Status |
|------|--------|
| Importar OFX | Em breve |
| Conciliação | Em breve |

## Relatórios

| Item | Status |
|------|--------|
| Fluxo de Caixa | Em breve |
| Abertos | Em breve |
| Vencidos | Em breve |
| Extrato | Em breve |

## Configurações

| Item | Status | Rota |
|------|--------|------|
| Usuários | **Implementado** | `/config/usuarios` |
| Alterar senha | **Implementado** | `/config/senha` |
| Integrações | Em breve | — |

## Fora do menu atual

Previstos no PRD, ainda **não** aparecem na nav:

- Fiscal (Documentos, Emitir, Configuração fiscal) — Fase 9
- Condições de pagamento (cadastro)

---

## Home (`/`)

Atalhos na página inicial (além do menu superior):

| Atalho | Rota |
|--------|------|
| Usuários | `/config/usuarios` |
| Alterar senha | `/config/senha` |

---

## Resumo

- **Cadastros implementados:** Negócios, Papéis, Pessoas, Produtos, Categorias, Bancos, Contas  
- **Stubs no menu:** Compras, Vendas, Estoque, Financeiro, Banco, Relatórios (exceto config)  
- **Shell:** grupos Cadastros, Compras, Vendas, Estoque, Financeiro, Banco, Relatórios, Configurações
