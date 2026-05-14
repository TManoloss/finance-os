# Task 2: Schema do Banco de Dados PostgreSQL Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Definir o schema completo do banco de dados e dados iniciais para o Personal Finance OS.

**Architecture:** PostgreSQL com UUID v4 para chaves primárias e relacionamentos.

**Tech Stack:** PostgreSQL 16.

---

### Task 2.1: Preparar diretórios e Extensões

**Files:**
- Create: `backend/internal/db/schema.sql`

- [ ] **Step 1: Criar diretório backend/internal/db**
Executar `mkdir -p backend/internal/db`.

- [ ] **Step 2: Definir schema.sql com extensões e tabelas base (users)**
Adicionar ativação do `pgcrypto` e criação da tabela `users`.

```sql
-- backend/internal/db/schema.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 3: Commit**
`git add backend/internal/db/schema.sql && git commit -m "db: init schema with users table and pgcrypto"`

---

### Task 2.2: Tabelas de Contas e Categorias

**Files:**
- Modify: `backend/internal/db/schema.sql`

- [ ] **Step 1: Adicionar tabelas connected_accounts e categories**

```sql
-- backend/internal/db/schema.sql
-- (Abaixo da tabela users)

CREATE TABLE IF NOT EXISTS connected_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pluggy_item_id TEXT,
    institution_name TEXT NOT NULL,
    account_type TEXT NOT NULL, -- checking, savings, credit
    balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'BRL',
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL = system category
    name TEXT NOT NULL,
    color TEXT,
    icon TEXT,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_connected_accounts_user_id ON connected_accounts(user_id);
```

- [ ] **Step 2: Commit**
`git add backend/internal/db/schema.sql && git commit -m "db: add connected_accounts and categories tables"`

---

### Task 2.3: Tabela de Transações (Feature Estrela)

**Files:**
- Modify: `backend/internal/db/schema.sql`

- [ ] **Step 1: Adicionar tabela transactions com campos de IA**

```sql
-- backend/internal/db/schema.sql

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES connected_accounts(id) ON DELETE CASCADE,
    pluggy_transaction_id TEXT UNIQUE,
    amount NUMERIC(12,2) NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('debit', 'credit')),
    description TEXT NOT NULL,
    merchant_name TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    -- IA Fields (Feature Estrela)
    needs_review BOOLEAN NOT NULL DEFAULT false,
    confidence_score NUMERIC(5,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_account_id_date ON transactions(account_id, date);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
CREATE INDEX idx_transactions_merchant_name ON transactions(merchant_name);
```

- [ ] **Step 2: Commit**
`git add backend/internal/db/schema.sql && git commit -m "db: add transactions table with AI fields"`

---

### Task 2.4: Tabelas de Suporte (Parcelas, Regras, Relatórios, Tokens)

**Files:**
- Modify: `backend/internal/db/schema.sql`

- [ ] **Step 1: Adicionar parcelas e regras**

```sql
-- backend/internal/db/schema.sql

CREATE TABLE IF NOT EXISTS installments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES connected_accounts(id) ON DELETE CASCADE,
    total_amount NUMERIC(12,2) NOT NULL,
    installments_total INT NOT NULL,
    installment_current INT NOT NULL,
    merchant_name TEXT NOT NULL,
    start_date DATE NOT NULL,
    next_due_date DATE
);

CREATE TABLE IF NOT EXISTS category_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    merchant_pattern TEXT NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    priority INT NOT NULL DEFAULT 0
);
```

- [ ] **Step 2: Adicionar relatórios e tokens**

```sql
-- backend/internal/db/schema.sql

CREATE TABLE IF NOT EXISTS agent_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_type TEXT NOT NULL, -- daily, weekly, monthly
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    summary_markdown TEXT NOT NULL,
    insights JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 3: Commit**
`git add backend/internal/db/schema.sql && git commit -m "db: add remaining tables (installments, rules, reports, tokens)"`

---

### Task 2.5: Dados Iniciais (Seed)

**Files:**
- Create: `backend/internal/db/seed.sql`

- [ ] **Step 1: Inserir categorias padrão e usuário de teste**

```sql
-- backend/internal/db/seed.sql

-- Categorias do sistema (user_id IS NULL)
INSERT INTO categories (name, color, icon) VALUES
('Alimentação', '#FF6B6B', 'restaurant'),
('Transporte', '#4D96FF', 'directions_car'),
('Saúde', '#6BCB77', 'medical_services'),
('Lazer', '#FFD93D', 'celebration'),
('Assinaturas', '#7C6FFF', 'subscriptions'),
('Moradia', '#FF9F45', 'home'),
('Educação', '#A084E8', 'school'),
('Investimentos', '#4ECDC4', 'trending_up'),
('Renda', '#19A7CE', 'payments'),
('Outros', '#8888A0', 'more_horiz')
ON CONFLICT DO NOTHING;

-- Usuário de teste
-- Password: admin123 (hash fixo para exemplo, idealmente seria gerado)
INSERT INTO users (id, name, email, password_hash)
VALUES ('00000000-0000-4000-a000-000000000001', 'Admin Teste', 'admin@example.com', '$2a$12$6/76yN.77uE.uO.uO.uO.uO.uO.uO.uO.uO.uO.uO.uO.uO.uO.uO')
ON CONFLICT (email) DO NOTHING;
```

- [ ] **Step 2: Commit**
`git add backend/internal/db/seed.sql && git commit -m "db: add default categories and test user seed"`
