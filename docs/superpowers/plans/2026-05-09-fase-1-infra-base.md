# Plano de Implementação: Personal Finance OS - Fase 1 (Infraestrutura Base)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configurar o ambiente de desenvolvimento local com Docker, definir o schema do banco de dados PostgreSQL e criar o esqueleto do backend em Go.

**Architecture:** Docker Compose para serviços de infra (DB/Admin), PostgreSQL 16 para dados e Backend Go com Echo Framework.

**Tech Stack:** Docker, PostgreSQL, Go 1.23+, Echo v4.

---

### Task 1: Configuração do Ambiente Docker e Git

**Files:**
- Create: `docker-compose.yml`
- Create: `.env.example`
- Create: `.gitignore`
- Create: `README.md`

- [ ] **Step 1: Criar o arquivo .env.example**
Definir as variáveis de ambiente necessárias para o banco de dados.

- [ ] **Step 2: Criar o .gitignore**
Ignorar arquivos sensíveis e binários.

- [ ] **Step 3: Criar o docker-compose.yml**
Configurar PostgreSQL e Adminer.

- [ ] **Step 4: Criar o README.md inicial**
Instruções de como rodar o projeto.

---

### Task 2: Schema do Banco de Dados PostgreSQL

**Files:**
- Create: `backend/internal/db/schema.sql`
- Create: `backend/internal/db/seed.sql`

- [ ] **Step 1: Definir o schema.sql**
Criar tabelas `users`, `connected_accounts`, `transactions`, `categories`, `category_rules`, `installments`, `agent_reports` e `refresh_tokens`.
*Nota: Incluir os campos `needs_review` e `confidence_score` na tabela `transactions`.*

- [ ] **Step 2: Criar o seed.sql**
Inserir categorias padrão e um usuário de teste.

---

### Task 3: Skeleton do Backend Go com Echo

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/router/router.go`
- Create: `backend/internal/response/response.go`

- [ ] **Step 1: Inicializar o módulo Go**
Executar `go mod init github.com/finance-os/backend`.

- [ ] **Step 2: Criar a estrutura de configuração**
Implementar o carregamento de variáveis de ambiente no `internal/config/config.go`.

- [ ] **Step 3: Implementar o main.go**
Configurar o servidor Echo, conexão com banco e middlewares básicos (Logger, Recover).

- [ ] **Step 4: Definir as rotas stub no router.go**
Configurar o grupo `/api/v1` com handlers que retornam `501 Not Implemented`.

---

### Task 4: Validação da Infraestrutura

- [ ] **Step 1: Subir o Docker**
Rodar `docker compose up -d` e verificar se os containers estão rodando.

- [ ] **Step 2: Validar o servidor Go**
Tentar rodar `go run cmd/server/main.go` dentro da pasta `backend` (após instalar dependências) e testar um endpoint stub via curl.
