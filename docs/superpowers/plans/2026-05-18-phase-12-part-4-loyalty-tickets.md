# Phase 12 Implementation Plan - Part 4: Merchant Loyalty & Ticket Analysis

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar os agentes finais da Fase 12: Análise de Ticket Médio (12.8) e Lealdade e Abandono de Merchants (12.10).

**Architecture:** Novos agentes Python seguindo o padrão `BaseAgent`, integração com a tabela de cache `report_cache` e novos endpoints no Backend Go.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Ticket Analysis Agent (12.8)

**Files:**
- Create: `agents/agents/ticket_analysis.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o TicketAnalysisAgent em Python**
- [ ] **Step 2: Implementar lógica de decomposição de gastos (Frequência vs Preço)**
- [ ] **Step 3: Registrar endpoints POST /agents/ticket-analysis e GET /reports/ticket-analysis**
- [ ] **Step 4: Commit**

---

### Task 2: Merchant Loyalty Agent (12.10)

**Files:**
- Create: `agents/agents/loyalty_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o LoyaltyAgent em Python**
- [ ] **Step 2: Implementar classificação de lealdade (LEAL, FREQUENTE, ABANDONADO, etc.)**
- [ ] **Step 3: Implementar cálculo de Abandonment Cost e Customer Lifetime Value**
- [ ] **Step 4: Registrar endpoints POST /agents/loyalty e GET /reports/loyalty**
- [ ] **Step 5: Commit**

---

### Task 3: Backend Endpoints for Part 4

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Implementar handlers GET /api/v1/reports/ticket-analysis e GET /api/v1/reports/loyalty**
- [ ] **Step 2: Registrar rotas no router**
- [ ] **Step 3: Commit**

---

### Task 4: UI Implementation (Web & Mobile) - Loyalty & Decomposition

**Files:**
- Create: `web/src/components/TicketAnalysisCard.tsx`
- Create: `web/src/components/LoyaltyAnalysisCard.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais no Next.js**
- [ ] **Step 2: Integrar na página de Relatórios Web**
- [ ] **Step 3: Adicionar widget de "Lealdade" no Mobile**
- [ ] **Step 4: Commit**
