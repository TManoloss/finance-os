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
