-- backend/internal/db/schema.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    pluggy_client_id TEXT,
    pluggy_client_secret_encrypted TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS connected_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pluggy_item_id TEXT, -- ID da conexão (um item pode ter várias contas)
    pluggy_account_id TEXT UNIQUE, -- ID único da conta na Pluggy
    institution_name TEXT NOT NULL,
    institution_logo TEXT,
    institution_color TEXT,
    account_type TEXT NOT NULL, -- CHECKING, SAVINGS, CREDIT
    subtype TEXT, -- CHECKING_ACCOUNT, SAVINGS_ACCOUNT, CREDIT_CARD
    balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'BRL',
    close_day INT DEFAULT 1,
    due_day INT DEFAULT 10,
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
    priority INT NOT NULL DEFAULT 0,
    UNIQUE(user_id, merchant_pattern)
);

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

-- Feed de eventos financeiros
CREATE TABLE IF NOT EXISTS feed_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    amount NUMERIC(12,2),
    severity TEXT DEFAULT 'info',
    related_tx_ids UUID[],
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_feed_events_user_id_created_at ON feed_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feed_events_user_id_unread ON feed_events(user_id, read_at) WHERE read_at IS NULL;

-- Score de saúde financeira (histórico)
CREATE TABLE IF NOT EXISTS health_score_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    score NUMERIC(5,2),
    cashflow_score NUMERIC(5,2),
    installments_score NUMERIC(5,2),
    consistency_score NUMERIC(5,2),
    subscriptions_score NUMERIC(5,2),
    diversification_score NUMERIC(5,2),
    trend_score NUMERIC(5,2),
    period_month DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, period_month)
);

-- Metas financeiras
CREATE TABLE IF NOT EXISTS financial_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    goal_type TEXT NOT NULL,
    target_amount NUMERIC(12,2),
    current_amount NUMERIC(12,2) DEFAULT 0,
    start_date DATE NOT NULL,
    target_date DATE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Simulações salvas
CREATE TABLE IF NOT EXISTS saved_simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    simulation_type TEXT NOT NULL,
    input_params JSONB,
    result_json JSONB,
    name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
