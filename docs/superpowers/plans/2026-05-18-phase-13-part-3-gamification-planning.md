# Phase 13 Implementation Plan - Part 3: Gamification & Planning

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o sistema de recompensas e planejamento: Feed de Conquistas (13.6), Missões (13.14), Planejamento de Salário (13.8) e Timeline de Parcelamentos (13.18).

**Architecture:** Novos agentes Python seguindo o padrão `BaseAgent`, integração com as tabelas `achievements_awarded`, `missions` e `salary_plans`, e novos endpoints no Backend Go.

**Tech Stack:** Go (Echo), Python (FastAPI), PostgreSQL (pgx).

---

### Task 1: Achievement & Mission Agents (13.6 & 13.14)

**Files:**
- Create: `agents/agents/gamification_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o GamificationAgent em Python**
- [ ] **Step 2: Lógica para verificar conquistas (ex: 3 dias sem delivery, primeiro investimento)**
- [ ] **Step 3: Lógica para gerar missões mensais baseadas no perfil do usuário**
- [ ] **Step 4: Registrar endpoints em main.py**
- [ ] **Step 5: Commit**

---

### Task 2: Salary Planning Agent (13.8)

**Files:**
- Create: `agents/agents/salary_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o SalaryAgent em Python**
- [ ] **Step 2: Lógica para calcular Safe Daily Limit baseado em compromissos fixos e data do salário**
- [ ] **Step 3: Registrar endpoints em main.py**
- [ ] **Step 4: Commit**

---

### Task 3: Installment Timeline Agent (13.18)

**Files:**
- Create: `agents/agents/installment_timeline_agent.py`
- Modify: `agents/main.py`

- [ ] **Step 1: Implementar o InstallmentTimelineAgent em Python**
- [ ] **Step 2: Gerar visão consolidada de parcelas futuras (meses a frente)**
- [ ] **Step 3: Registrar endpoints em main.py**
- [ ] **Step 4: Commit**

---

### Task 4: Backend Integration for Part 3

**Files:**
- Modify: `backend/internal/handler/reports.go`
- Modify: `backend/internal/router/router.go`
- Create: `backend/internal/service/gamification_service.go`

- [ ] **Step 1: Implementar o GamificationService em Go**
- [ ] **Step 2: Criar endpoints GET para Conquistas, Missões e Plano de Salário**
- [ ] **Step 3: Registrar rotas**
- [ ] **Step 4: Commit**

---

### Task 5: UI Implementation (Web & Mobile) - Rewards & Plans

**Files:**
- Create: `web/src/components/AchievementsFeed.tsx`
- Create: `web/src/components/MissionsCard.tsx`
- Create: `web/src/components/InstallmentTimeline.tsx`
- Modify: `web/src/app/dashboard/reports/page.tsx`
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Criar componentes visuais de Gamificação no Next.js**
- [ ] **Step 2: Integrar Timeline de Parcelamentos no Dashboard Web**
- [ ] **Step 3: Adicionar widget de "Conquistas do Mês" no Mobile**
- [ ] **Step 4: Commit**
