# Especificação de Design: Personal Finance OS - Autenticação JWT e Segurança

**Data:** 2026-05-09
**Status:** Em Revisão
**Tópico:** Fase 2 - Autenticação JWT, Gestão de Usuários e Segurança

---

## 1. Objetivo
Implementar um sistema de autenticação robusto e seguro para garantir que apenas membros autorizados da família acessem seus próprios dados financeiros, utilizando padrões modernos de mercado (JWT + Refresh Tokens).

## 2. Requisitos de Segurança

### 2.1 Política de Senhas (Rigorosa)
- Mínimo de 10 caracteres.
- Pelo menos uma letra maiúscula.
- Pelo menos uma letra minúscula.
- Pelo menos um número.
- Pelo menos um caractere especial (@, #, $, %, etc.).
- Hashing: **Bcrypt** com custo de processamento 12.

### 2.2 Gestão de Sessão (Tokens)
- **Access Token:** JWT (JSON Web Token), expiração de 15 minutos. Claims: `user_id`, `email`.
- **Refresh Token:** UUID opaco armazenado no banco de dados (`refresh_tokens`), expiração de 7 dias.
- **Rotação de Tokens:** A cada refresh bem-sucedido, o token antigo é invalidado e um novo par (Access + Refresh) é gerado.

## 3. Arquitetura de Software (Backend Go)

A implementação seguirá o padrão de camadas já estabelecido:
1. **Handler (`internal/handler/auth.go`):** Lida com o parsing do JSON, chamadas ao service e formatação da resposta via `internal/response`.
2. **Service (`internal/service/auth_service.go`):** Orquestra a lógica de negócio (validar senha, gerar tokens, revogar sessões).
3. **Repository (`internal/repository/user_repository.go`):** Executa as queries SQL via `pgx` para as tabelas `users` e `refresh_tokens`.

## 4. Endpoints da API

| Método | Endpoint | Proteção | Descrição |
| :--- | :--- | :--- | :--- |
| POST | `/api/v1/auth/register` | Pública | Cria um novo usuário e retorna tokens. |
| POST | `/api/v1/auth/login` | Pública | Valida credenciais e retorna tokens. |
| POST | `/api/v1/auth/refresh` | Pública | Gera novo par de tokens usando Refresh Token válido. |
| GET | `/api/v1/me` | JWT | Retorna dados do perfil do usuário logado. |

## 5. Estratégia de Validação (cURL)
Para cada endpoint implementado, será executado um comando `curl` para validar:
- Sucesso (Status 200/201).
- Erro de validação (Status 400).
- Erro de credenciais/não autorizado (Status 401).

---
**Próximos Passos:**
1. Aprovação do design pelo usuário.
2. Criação do plano de implementação detalhado (Task-by-Task).
