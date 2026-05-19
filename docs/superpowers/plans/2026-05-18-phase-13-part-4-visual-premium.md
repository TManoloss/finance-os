# Phase 13 Implementation Plan - Part 4: Visual & Premium Experience

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finalizar a Fase 13 implementando as experiências visuais premium: Mapa de Dependência (13.11), Replay Financeiro do Mês (13.15) e Heatmap de Gastos (13.16).

**Architecture:** Criação dos últimos agentes Python (`dependency_map_agent.py`, `replay_agent.py`), exposição via Backend Go e desenvolvimento de componentes visuais complexos no Next.js e Flutter.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx), Recharts/D3, Flutter CustomPaint.

---

### Task 1: Dependency Map Agent (13.11)

**Files:**
- Create: `agents/agents/dependency_map_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o DependencyMapAgent em Python**
- [ ] **Step 2: Gerar estrutura hierárquica (Categoria -> Merchant -> Gasto)**
- [ ] **Step 3: Determinar o dependency_level (Crítica, Alta, Normal)**
- [ ] **Step 4: Registrar endpoint POST /agents/dependency-map e GET /reports/dependency-map**
- [ ] **Step 5: Commit**

---

### Task 2: Monthly Replay Agent (13.15)

**Files:**
- Create: `agents/agents/monthly_replay_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o MonthlyReplayAgent em Python**
- [ ] **Step 2: Coletar dados mensais completos (cashflow, top merchants, recordes)**
- [ ] **Step 3: Usar Claude para gerar a narrativa estilo Spotify Wrapped**
- [ ] **Step 4: Salvar em `monthly_replays`**
- [ ] **Step 5: Registrar endpoints POST /agents/monthly-replay e GET /reports/monthly-replay**
- [ ] **Step 6: Commit**

---

### Task 3: Spending Heatmap Data (13.16) & Backend Endpoints

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`
- Create: `backend/internal/service/visual_reports_service.go`

- [ ] **Step 1: Implementar `GetSpendingHeatmap` no Go (agrega transações diárias do último ano)**
- [ ] **Step 2: Expor `GetDependencyMap` e `GetMonthlyReplay` no Go via cache/Python**
- [ ] **Step 3: Registrar as 3 rotas no router**
- [ ] **Step 4: Commit**

---

### Task 4: UI Implementation (Web) - Visual Premium

**Files:**
- Create: `web/src/components/DependencyTreemap.tsx`
- Create: `web/src/components/SpendingHeatmap.tsx`
- Create: `web/src/app/dashboard/reports/replay/[month]/page.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`

- [ ] **Step 1: Implementar DependencyTreemap usando Recharts**
- [ ] **Step 2: Implementar SpendingHeatmap estilo GitHub**
- [ ] **Step 3: Criar página isolada e imersiva para o Replay Financeiro (`/replay/[month]`)**
- [ ] **Step 4: Integrar componentes na página de Relatórios**
- [ ] **Step 5: Commit**

---

### Task 5: UI Implementation (Mobile) - Visual Premium

**Files:**
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`
- Create: `mobile/lib/features/reports/presentation/replay_screen.dart`

- [ ] **Step 1: Criar a tela imersiva `ReplayScreen` no Flutter**
- [ ] **Step 2: Adicionar chamada/botão para o Replay e Heatmap no Dashboard**
- [ ] **Step 3: Commit**
