# Design Spec: Pluggy Multi-Key Architecture

**Date:** 2026-05-11
**Topic:** Suporte a múltiplas chaves de API Pluggy por usuário com criptografia AES-GCM.
**Status:** Draft

## 1. Overview
Atualmente, o sistema utiliza um par único de `PLUGGY_CLIENT_ID` e `PLUGGY_CLIENT_SECRET` via variáveis de ambiente. Para suportar múltiplos usuários da família com suas próprias contas Pluggy, cada usuário deve fornecer suas próprias credenciais. Estas credenciais serão armazenadas de forma segura no banco de dados.

## 2. Architecture Changes

### 2.1 Database (PostgreSQL)
Alteração na tabela `public.users`:
- Adicionar `pluggy_client_id` (TEXT, nullable)
- Adicionar `pluggy_client_secret_encrypted` (TEXT, nullable)

### 2.2 Backend (Go)
- **Encryption Service:** Novo serviço em `internal/service/encryption_service.go` utilizando AES-GCM (256-bit). A chave mestre (`ENCRYPTION_KEY`) será lida do `.env`.
- **Pluggy Client Refactor:** O `pluggy.Client` deixará de ler do config global. O `NewClient` passará a aceitar `clientID` e `clientSecret` explicitamente.
- **Service/Handler Update:**
    - `AuthService` e `UserRepository` serão atualizados para lidar com os novos campos.
    - `AccountsHandler` e `SyncService` buscarão as credenciais do usuário logado, as descriptografarão e instanciarão o cliente Pluggy sob demanda.

### 2.3 Frontend (Next.js & Flutter)
- **Onboarding Middleware:** Implementar verificação no Dashboard. Se `user.pluggy_client_id` for nulo, redirecionar para a tela de "Configuração de Chaves".
- **Setup Screen:** Nova interface para inserir Client ID e Client Secret da Pluggy, com validação básica.

## 3. Security
- **Encryption:** AES-GCM fornece tanto confidencialidade quanto integridade (AEAD).
- **Key Management:** A `ENCRYPTION_KEY` deve ser uma string de 32 bytes (para AES-256) codificada em Base64 no `.env`.

## 4. Implementation Strategy (Micro-tasks)
A implementação será dividida em:
1. Infra de Criptografia (Go).
2. Mudança no Schema e Modelos (Go/SQL).
3. Refatoração do Cliente Pluggy (Go).
4. Handlers de Gerenciamento de Chaves (Go).
5. UI de Onboarding (Next.js).
6. UI de Onboarding (Flutter).

## 5. Success Criteria
- Um usuário pode cadastrar suas chaves e sincronizar suas contas com sucesso.
- As chaves no banco de dados estão ilegíveis sem a `ENCRYPTION_KEY`.
- Usuários sem chaves são bloqueados de acessar funcionalidades que dependem da Pluggy.
