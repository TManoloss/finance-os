# Phase 13 Implementation Plan - Part 1: Emotional Impact Core

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o núcleo de inteligência emocional: Modo Sobrevivência (13.1), Stress Score (13.2), Explicador Inteligente (13.3) e CFO Pessoal (13.12).

**Architecture:** Novos serviços em Go e Python, tabelas de snapshot de stress e sobrevivência, e agentes proativos que enviam insights via Feed.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Database Schema for Phase 13 (Core)

**Files:**
- Modify: `backend/internal/db/schema.sql`
- Create: `backend/internal/db/migrations/20260518_phase_13_emotional_core.sql`

- [ ] **Step 1: Criar tabelas `survival_mode_snapshots`, `stress_score_snapshots` e `financial_timeline_events`**
- [ ] **Step 2: Atualizar schema.sql**
- [ ] **Step 3: Aplicar migração**
- [ ] **Step 4: Commit**

---

### Task 2: Stress Score & Survival Mode Agents (13.1 & 13.2)

**Files:**
- Create: `agents/agents/stress_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o StressScoreAgent em Python**
- [ ] **Step 2: Implementar a lógica do Modo Sobrevivência (calculada em conjunto ou separada)**
- [ ] **Step 3: Registrar endpoints em main.py**
- [ ] **Step 4: Commit**

---

### Task 3: Expense Explainer & Proactive CFO (13.3 & 13.12)

**Files:**
- Create: `agents/agents/cfo_agent.py`
- Modify: `agents/main.py`
- Modify: `agents/agents/chat.py` (para integração com explainer)

- [ ] **Step 1: Implementar o ExpenseExplainerAgent**
- [ ] **Step 2: Implementar o CFOAgent (insights proativos diários)**
- [ ] **Step 3: Integrar o explainer no ChatAgent existente**
- [ ] **Step 4: Commit**

---

### Task 4: Backend Integration for Emotional Core

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`
- Create: `backend/internal/service/survival_mode_service.go`

- [ ] **Step 1: Implementar o SurvivalModeService em Go**
- [ ] **Step 2: Criar endpoints GET /reports/survival-mode e GET /reports/stress-score**
- [ ] **Step 3: Registrar rotas**
- [ ] **Step 4: Commit**

---

### Task 5: UI Implementation (Web & Mobile) - Emotional Feedback

**Files:**
- Create: `web/src/components/SurvivalModeOverlay.tsx`
- Create: `web/src/components/StressScoreBadge.tsx`
- Modify: `web/src/app/dashboard/layout.tsx` (para mudanças visuais sutis)
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais no Next.js**
- [ ] **Step 2: Implementar mudanças de UI baseadas no nível de sobrevivência**
- [ ] **Step 3: Adicionar Stress Score e Survival Widget no Mobile**
- [ ] **Step 4: Commit**
