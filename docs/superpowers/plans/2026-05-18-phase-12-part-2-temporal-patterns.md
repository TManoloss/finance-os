# Phase 12 Implementation Plan - Part 2: Temporal & Lifestyle Patterns

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar os agentes de análise temporal e de estilo de vida: Fim de Semana vs Útil (12.2), Perfil dos 7 Dias (12.9), Efeito do Salário (12.3) e Ciclo Mensal (12.12).

**Architecture:** Novos agentes Python seguindo o padrão `BaseAgent`, integração com a tabela de cache `report_cache` e novos endpoints no Backend Go.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Weekday vs Weekend & 7-Day Profile Agents (12.2 & 12.9)

**Files:**
- Create: `agents/agents/weekly_profile.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o WeeklyProfileAgent em Python (cobrindo 12.2 e 12.9)**
- [ ] **Step 2: Registrar endpoints POST /agents/weekly-profile e POST /agents/weekday-weekend**
- [ ] **Step 3: Commit**

---

### Task 2: Salary Effect & Monthly Weeks Agents (12.3 & 12.12)

**Files:**
- Create: `agents/agents/monthly_cycle.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o MonthlyCycleAgent em Python (cobrindo 12.3 e 12.12)**
- [ ] **Step 2: Registrar endpoints POST /agents/salary-effect e POST /agents/monthly-weeks**
- [ ] **Step 3: Commit**

---

### Task 3: Backend Endpoints for Part 2

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Implementar handlers GET para os 4 novos relatórios (com lógica de cache)**
- [ ] **Step 2: Registrar rotas no router**
- [ ] **Step 3: Commit**

---

### Task 4: UI Implementation (Web & Mobile) - Temporal Insights

**Files:**
- Create: `web/src/components/WeeklyProfileChart.tsx`
- Create: `web/src/components/MonthlyCycleChart.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais no Next.js**
- [ ] **Step 2: Integrar na página de Relatórios Web**
- [ ] **Step 3: Adicionar widget de "Ciclo Mensal" no Mobile**
- [ ] **Step 4: Commit**
