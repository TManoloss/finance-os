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
