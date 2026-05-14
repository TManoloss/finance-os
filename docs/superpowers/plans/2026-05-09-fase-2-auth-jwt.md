# Plano de Implementação: Personal Finance OS - Fase 2 (Autenticação JWT)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o sistema completo de autenticação JWT, incluindo registro, login, refresh de tokens e rota protegida.

**Architecture:** Camadas Handler-Service-Repository no Go. Bcrypt para senhas e JWT/UUID para tokens.

**Tech Stack:** Go, Echo, JWT-Go, Bcrypt, PostgreSQL (pgx).

---

### Task 1: Repositório de Usuário e JWT Service

**Files:**
- Create: `backend/internal/repository/user_repository.go`
- Create: `backend/internal/service/jwt_service.go`
- Modify: `backend/internal/config/config.go` (adicionar JWT_SECRET)

- [x] **Step 1: Implementar o UserRepository**
Criar métodos `Create`, `FindByEmail`, `FindByID` e métodos para gerenciar `refresh_tokens`.

- [x] **Step 2: Implementar o JWTService**
Criar funções para gerar `AccessToken` (JWT) e `RefreshToken` (UUID), e validar os mesmos.

---

### Task 3: Registro de Usuários (POST /auth/register)

**Files:**
- Modify: `backend/internal/handler/auth.go`
- Modify: `backend/internal/service/auth_service.go`

- [x] **Step 1: Implementar lógica de registro**
Validar política de senha rigorosa, fazer o hash com bcrypt (custo 12) e salvar no banco.

- [x] **Step 2: Validar com cURL**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
-H "Content-Type: application/json" \
-d '{"name": "Manoel", "email": "manoel@example.com", "password": "Senha@Rigorosa123"}'
```
*Esperado: Status 201 com tokens no corpo da resposta.*

---

### Task 4: Login de Usuários (POST /auth/login)

- [x] **Step 1: Implementar lógica de login**
Buscar usuário por email, comparar hash da senha e gerar novo par de tokens.

- [x] **Step 2: Validar com cURL**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
-H "Content-Type: application/json" \
-d '{"email": "manoel@example.com", "password": "Senha@Rigorosa123"}'
```
*Esperado: Status 200 com novos tokens.*

---

### Task 5: Rota Protegida e Middleware JWT (GET /me)

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Modify: `backend/internal/router/router.go`

- [x] **Step 1: Implementar Middleware de Autenticação**
Validar o token JWT no header `Authorization` e injetar o `user_id` no contexto do Echo.

- [x] **Step 2: Implementar handler GET /me**
Retornar dados do usuário extraídos do token.

- [x] **Step 3: Validar com cURL**
```bash
# Substitua <TOKEN> pelo access_token recebido no login
curl -X GET http://localhost:8080/api/v1/me \
-H "Authorization: Bearer <TOKEN>"
```
*Esperado: Status 200 com dados do usuário.*

---

### Task 6: Refresh de Tokens (POST /auth/refresh)

- [x] **Step 1: Implementar lógica de rotação de tokens**
Validar refresh token no banco, deletá-lo e gerar novo par (Access + Refresh).

- [x] **Step 2: Validar com cURL**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
-H "Content-Type: application/json" \
-d '{"refresh_token": "<REFRESH_TOKEN>"}'
```
*Esperado: Status 200 com novos tokens e o antigo invalidado no banco.*
