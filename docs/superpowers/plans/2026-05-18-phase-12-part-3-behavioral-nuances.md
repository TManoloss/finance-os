# Phase 12 Implementation Plan - Part 3: Behavioral Nuances & Specific Costs

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar os agentes de inteligência comportamental e análise de custos específicos: Análise de Impulso (12.4), Custo Real por Refeição (12.5), Índice de Conveniência (12.6) e Padrão de Compensação (12.11).

**Architecture:** Novos agentes Python seguindo o padrão `BaseAgent`, integração com a tabela de cache `report_cache` e novos endpoints no Backend Go.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Impulse & Compensation Agents (12.4 & 12.11)

**Files:**
- Create: `agents/agents/behavioral_nuances.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o BehavioralNuancesAgent em Python (cobrindo 12.4 e 12.11)**
- [ ] **Step 2: Registrar endpoints POST /agents/impulse e POST /agents/compensation**
- [ ] **Step 3: Commit**

---

### Task 2: Meal Cost & Convenience Index Agents (12.5 & 12.6)

**Files:**
- Create: `agents/agents/specific_costs.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o SpecificCostsAgent em Python (cobrindo 12.5 e 12.6)**
- [ ] **Step 2: Registrar endpoints POST /agents/meal-cost e POST /agents/convenience-index**
- [ ] **Step 3: Commit**

---

### Task 3: Backend Endpoints for Part 3

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Implementar handlers GET para os 4 novos relatórios (com lógica de cache)**
- [ ] **Step 2: Registrar rotas no router**
- [ ] **Step 3: Commit**

---

### Task 4: UI Implementation (Web & Mobile) - Nuanced Insights

**Files:**
- Create: `web/src/components/ImpulseAnalysisCard.tsx`
- Create: `web/src/components/MealCostAnalysisCard.tsx`
- Create: `web/src/components/ConvenienceIndexCard.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais no Next.js**
- [ ] **Step 2: Integrar na página de Relatórios Web**
- [ ] **Step 3: Adicionar widget de "Gasto por Refeição" no Mobile**
- [ ] **Step 4: Commit**
