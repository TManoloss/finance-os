# Pluggy Multi-Key Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrar o sistema de uma chave global da Pluggy para chaves individuais por usuário, armazenadas de forma segura com criptografia AES-GCM.

**Architecture:** Implementação de um `EncryptionService` no Go para lidar com segredos, migração da tabela `users` no PostgreSQL, refatoração do `pluggy.Client` para injeção dinâmica de credenciais e criação de fluxo de onboarding no Frontend.

**Tech Stack:** Go (Echo), PostgreSQL (pgx), AES-GCM (Go standard library), Next.js 15, Flutter.

---

### Task 1: Setup de Segurança (Encryption Key) [DONE]

**Files:**
- Modify: `.env.example`
- Modify: `.env`

- [x] **Step 1: Adicionar a chave de criptografia ao .env.example**
- [x] **Step 2: Gerar e adicionar a chave ao .env local**
- [x] **Step 3: Commit**

---

### Task 2: Implementação do Encryption Service (Go) [DONE]

**Files:**
- Create: `backend/internal/service/encryption_service.go`
- Create: `backend/internal/service/encryption_service_test.go`

- [x] **Step 1: Escrever teste unitário para criptografia e descriptografia**
- [x] **Step 2: Implementar o serviço AES-GCM**
- [x] **Step 3: Rodar testes e verificar sucesso**
- [x] **Step 4: Commit**

---

### Task 3: Migração do Banco de Dados [DONE]

**Files:**
- Modify: `backend/internal/db/schema.sql`
- Create: `backend/internal/db/migrations/20260511_add_pluggy_keys_to_users.sql`

- [x] **Step 1: Criar arquivo de migração**
- [x] **Step 2: Atualizar schema.sql principal**
- [x] **Step 3: Commit**

---

### Task 4: Atualização de Modelos e Repositório de Usuário (Go) [DONE]

**Files:**
- Modify: `backend/internal/models/user.go`
- Modify: `backend/internal/repository/user_repository.go`

- [x] **Step 1: Atualizar struct User**
- [x] **Step 2: Atualizar métodos do UserRepository**
- [x] **Step 3: Commit**

---

### Task 5: Refatoração do Pluggy Client (Go) [DONE]

**Files:**
- Modify: `backend/internal/pluggy/client.go`

- [x] **Step 1: Mudar assinatura de NewClient**
- [x] **Step 2: Atualizar método Authenticate**
- [x] **Step 3: Commit**

---

### Task 6: Implementação do Handler de Configuração de Chaves (Go) [DONE]

**Files:**
- Modify: `backend/internal/handler/accounts.go`
- Modify: `backend/internal/router/router.go`

- [x] **Step 1: Criar endpoint POST /api/v1/accounts/keys**
- [x] **Step 2: Registrar rota no router**
- [x] **Step 3: Commit**

---

### Task 7: Atualização do Sync Service e Handlers Existentes (Go) [DONE]

**Files:**
- Modify: `backend/internal/service/sync_service.go`
- Modify: `backend/internal/handler/accounts.go`

- [x] **Step 1: Atualizar SyncUserAccounts**
- [x] **Step 2: Atualizar ConnectToken Handler**
- [x] **Step 3: Commit**

---

### Task 8: UI de Onboarding - Next.js [DONE]

**Files:**
- Create: `web/src/app/onboarding/pluggy/page.tsx`
- Modify: `web/src/middleware.ts`

- [x] **Step 1: Criar formulário de setup de chaves**
- [x] **Step 2: Implementar lógica de bloqueio no middleware**
- [x] **Step 3: Commit**

---

### Task 9: UI de Onboarding - Flutter [DONE]

**Files:**
- Create: `mobile/lib/features/auth/presentation/pluggy_setup_screen.dart`
- Modify: `mobile/lib/core/router/app_router.dart`

- [x] **Step 1: Criar tela de setup no Flutter**
- [x] **Step 2: Atualizar lógica de roteamento**
- [x] **Step 3: Commit**
