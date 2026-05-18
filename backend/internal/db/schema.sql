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

-- report_cache: Cache system for expensive agent reports
CREATE TABLE IF NOT EXISTS report_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    report_type TEXT NOT NULL,
    period_key TEXT NOT NULL, -- e.g., "2024-05", "2024-W20"
    result_json JSONB NOT NULL DEFAULT '{}',
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE(user_id, report_type, period_key)
);
CREATE INDEX IF NOT EXISTS idx_report_cache_user_expires ON report_cache(user_id, expires_at);

-- inflation_snapshots: Tracking personal inflation vs official indices
CREATE TABLE IF NOT EXISTS inflation_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_month DATE NOT NULL,
    personal_inflation_rate NUMERIC(5,2) NOT NULL,
    ipca_rate NUMERIC(5,2), -- Official Brazilian IPCA
    category_breakdown JSONB NOT NULL DEFAULT '{}', -- Inflation impact per category
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, period_month)
);
CREATE INDEX IF NOT EXISTS idx_inflation_snapshots_user_period ON inflation_snapshots(user_id, period_month);

-- day_profiles_cache: Behavioral patterns per day of week
CREATE TABLE IF NOT EXISTS day_profiles_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 6=Saturday
    avg_spending NUMERIC(12,2) NOT NULL DEFAULT 0,
    top_categories JSONB NOT NULL DEFAULT '[]',
    top_merchants JSONB NOT NULL DEFAULT '[]',
    peak_hour_range TEXT, -- e.g., "18:00-20:00"
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, day_of_week)
);
CREATE INDEX IF NOT EXISTS idx_day_profiles_cache_user ON day_profiles_cache(user_id);

-- Survival mode snapshots
CREATE TABLE IF NOT EXISTS survival_mode_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    risk_score NUMERIC(5,2),
    level TEXT,
    is_active BOOLEAN DEFAULT false,
    projected_shortfall NUMERIC(10,2),
    days_until_salary INT,
    top_risks JSONB DEFAULT '{}',
    computed_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_survival_mode_user_computed ON survival_mode_snapshots(user_id, computed_at DESC);

-- Stress score snapshots
CREATE TABLE IF NOT EXISTS stress_score_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score NUMERIC(5,2),
    level TEXT,
    components JSONB DEFAULT '{}',
    trend TEXT,
    computed_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stress_score_user_computed ON stress_score_snapshots(user_id, computed_at DESC);

-- Conquistas
CREATE TABLE IF NOT EXISTS achievements_awarded (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id TEXT NOT NULL,
    awarded_at TIMESTAMPTZ DEFAULT NOW(),
    context_data JSONB DEFAULT '{}'
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_achievement_monthly ON achievements_awarded (user_id, achievement_id, (date_trunc('month', awarded_at AT TIME ZONE 'UTC')));

-- Missões
CREATE TABLE IF NOT EXISTS missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id TEXT NOT NULL,
    title TEXT,
    description TEXT,
    target_value NUMERIC(10,2),
    current_value NUMERIC(10,2) DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ends_at TIMESTAMPTZ,
    status TEXT DEFAULT 'active',
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_missions_user_status ON missions(user_id, status);

-- Monthly replay
CREATE TABLE IF NOT EXISTS monthly_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_month DATE NOT NULL,
    narrative TEXT,
    highlight_stat TEXT,
    replay_data JSONB DEFAULT '{}',
    generated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, period_month)
);

-- Salary plans
CREATE TABLE IF NOT EXISTS salary_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    salary_detected NUMERIC(10,2),
    fixed_commitments NUMERIC(10,2),
    safe_daily_limit NUMERIC(8,2),
    plan_data JSONB DEFAULT '{}',
    valid_until DATE,
    generated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Financial timeline events
CREATE TABLE IF NOT EXISTS financial_timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    event_date DATE NOT NULL,
    title TEXT,
    narrative TEXT,
    event_data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_financial_timeline_user_date ON financial_timeline_events(user_id, event_date DESC);
