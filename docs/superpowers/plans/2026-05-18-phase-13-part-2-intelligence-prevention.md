# Phase 13 Implementation Plan - Part 2: Intelligence & Prevention

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar a inteligência histórica e preditiva: Timeline de Vida (13.4), Lifestyle Drift (13.5), Memória Financeira (13.7), Radar Anti-Impulso (13.9), Dias Perigosos (13.10), Previsão Comportamental (13.13) e Pequenos Vazamentos (13.17).

**Architecture:** Novos agentes Python seguindo o padrão `BaseAgent`, integração com a tabela `financial_timeline_events` e novos endpoints no Backend Go.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Financial Timeline & Lifestyle Drift Agents (13.4 & 13.5)

**Files:**
- Create: `agents/agents/timeline_drift_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o TimelineDriftAgent em Python**
- [ ] **Step 2: Lógica para detectar eventos significativos (13.4) e drift de estilo de vida (13.5)**
- [ ] **Step 3: Registrar endpoints em main.py**
- [ ] **Step 4: Commit**

---

### Task 2: Financial Memory & Dangerous Days Agents (13.7 & 13.10)

**Files:**
- Create: `agents/agents/memory_prediction_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o MemoryPredictionAgent em Python**
- [ ] **Step 2: Lógica para Year-Over-Year context (13.7) e padrões temporais perigosos (13.10)**
- [ ] **Step 3: Registrar endpoints em main.py**
- [ ] **Step 4: Commit**

---

### Task 3: Behavioral Prediction & Micro-spending Agents (13.13 & 13.17)

**Files:**
- Create: `agents/agents/future_leakage_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o FutureLeakageAgent em Python**
- [ ] **Step 2: Lógica para Monte Carlo prediction (13.13) e detecção de pequenos vazamentos (13.17)**
- [ ] **Step 3: Registrar endpoints em main.py**
- [ ] **Step 4: Commit**

---

### Task 4: Anti-Impulse Radar & Backend Integration (13.9)

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`
- Create: `backend/internal/service/impulse_radar_service.go`

- [ ] **Step 1: Implementar o ImpulseRadarService em Go**
- [ ] **Step 2: Criar endpoints para os novos relatórios (Timeline, Drift, Memory, Prediction, Leakage)**
- [ ] **Step 3: Registrar rotas**
- [ ] **Step 4: Commit**

---

### Task 5: UI Implementation (Web & Mobile) - History & Future

**Files:**
- Create: `web/src/components/FinancialTimeline.tsx`
- Create: `web/src/components/LifestyleDriftCard.tsx`
- Create: `web/src/components/MicroSpendingCard.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais no Next.js**
- [ ] **Step 2: Integrar Timeline e Insights Preditivos no Dashboard Web**
- [ ] **Step 3: Adicionar widget de "Dias Perigosos" e Timeline no Mobile**
- [ ] **Step 4: Commit**
