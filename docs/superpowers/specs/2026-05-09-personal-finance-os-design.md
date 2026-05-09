# Especificação de Design: Personal Finance OS - MVP de Automação

**Data:** 2026-05-09
**Status:** Em Revisão
**Tópico:** Infraestrutura Base e Automação de Transações (Fase 1 e 2 do GEMINI.md adaptado)

---

## 1. Objetivo
Criar um gestor financeiro familiar onde a entrada de dados é 100% automatizada via Open Finance (Pluggy) e a categorização é assistida por IA, garantindo que o usuário tenha controle sobre as dúvidas do sistema sem ser interrompido constantemente.

## 2. Arquitetura do Sistema

O sistema é composto por 4 camadas principais:
1.  **Backend (Go/Echo):** Orquestrador, lida com autenticação, integração Pluggy e API para os frontends.
2.  **Serviço de Agentes (Python/FastAPI):** Inteligência Artificial utilizando Claude (Anthropic) para classificar transações.
3.  **Banco de Dados (PostgreSQL):** Persistência de usuários, contas, transações e regras de categoria.
4.  **Clientes (Next.js & Flutter):** Interfaces para visualização e revisão de dados.

### Fluxo de Dados de Transação:
`Pluggy API` -> `Backend Go` -> `Agente Python (LLM)` -> `PostgreSQL` -> `Dashboard (Web/Mobile)`

## 3. Lógica de Classificação e Revisão (Feature Estrela)

Para cada transação importada:
1.  O **Agente Python** recebe o `merchant_name` e a `description`.
2.  O Claude retorna um JSON com `category_id` e `confidence_score` (0.0 a 1.0).
3.  **Regra de Negócio:**
    *   Se `confidence_score` >= 0.85: A transação é marcada como `needs_review = false`.
    *   Se `confidence_score` < 0.85: A transação é marcada como `needs_review = true`.
4.  No app, transações com `needs_review = true` exibem um **balão/badge roxo** indicando dúvida.
5.  O usuário pode confirmar ou corrigir. Correções geram uma entrada na tabela `category_rules` para futuras automações.

## 4. Modelo de Acesso e Segurança
*   **Privacidade:** Isolamento total por `user_id`. Um membro da família não vê os dados do outro.
*   **Auth:** JWT com Refresh Tokens.
*   **Infra:** Docker Compose para desenvolvimento local.

## 5. Componentes de UI (Pierre Style)
*   **Tema:** Dark Mode exclusivo (#0A0A0F).
*   **Central de Revisão:** Um card de destaque no dashboard que agrupa todas as transações com `needs_review = true`.
*   **Transações:** Lista com ícones de categoria e o "balão de dúvida" roxo para as incertas.

## 6. Plano de Dados (Schema Simplificado)
*   `users`: id, name, email, password_hash.
*   `transactions`: id, user_id, amount, description, merchant_name, category_id, needs_review, confidence_score, date.
*   `category_rules`: id, user_id, merchant_pattern, category_id.

---
**Próximos Passos:**
1.  Aprovação do design pelo usuário.
2.  Criação do plano de implementação detalhado (Fase 1 e 2).
