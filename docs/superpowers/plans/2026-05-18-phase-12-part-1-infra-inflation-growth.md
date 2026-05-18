# Phase 12 Implementation Plan - Part 1: Core Infra & Early Insights

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o sistema de cache de relatórios e os primeiros agentes de análise da Fase 12: Inflação Pessoal (12.1) e Crescimento Silencioso (12.7).

**Architecture:** Atualização do schema PostgreSQL para suporte a cache e snapshots. Implementação de novos serviços de agentes em Python seguindo o padrão `BaseAgent`. Endpoints em Go para exposição dos novos relatórios.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Database Schema Update (Phase 12 Core)

**Files:**
- Modify: `backend/internal/db/schema.sql`
- Create: `backend/internal/db/migrations/20260518_phase_12_cache_and_inflation.sql`

- [ ] **Step 1: Criar arquivo de migração com as tabelas de cache e inflação**
- [ ] **Step 2: Atualizar o schema.sql principal**
- [ ] **Step 3: Aplicar a migração ao banco de dados**
- [ ] **Step 4: Commit**

---

### Task 2: Personal Inflation Agent (12.1) - Implementation

**Files:**
- Create: `agents/agents/personal_inflation.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o PersonalInflationAgent em Python**
- [ ] **Step 2: Registrar o endpoint POST /agents/personal-inflation em main.py**
- [ ] **Step 3: Commit**

---

### Task 4: Silent Growth Agent (12.7) - Implementation

**Files:**
- Create: `agents/agents/silent_growth.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o SilentGrowthAgent em Python**
- [ ] **Step 2: Registrar o endpoint POST /agents/silent-growth em main.py**
- [ ] **Step 3: Commit**

---

### Task 5: Backend Endpoints for Phase 12 (Part 1)

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Implementar handlers GET /reports/personal-inflation e GET /reports/silent-growth**
- [ ] **Step 2: Registrar rotas no router**
- [ ] **Step 3: Commit**

---

### Task 6: UI Implementation (Next.js & Flutter) - Early Previews

**Files:**
- Create: `web/src/components/PersonalInflationCard.tsx`
- Create: `web/src/components/SilentGrowthCard.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais no Next.js**
- [ ] **Step 2: Integrar no Dashboard Web**
- [ ] **Step 3: Adicionar prévia no Dashboard Mobile**
- [ ] **Step 4: Commit**
