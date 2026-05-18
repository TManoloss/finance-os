-- Phase 12 Database Schema Update: Cache and Inflation
-- Created at: 2026-05-18

-- 1. report_cache: Cache system for expensive agent reports
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
CREATE INDEX idx_report_cache_user_expires ON report_cache(user_id, expires_at);

-- 2. inflation_snapshots: Tracking personal inflation vs official indices
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
CREATE INDEX idx_inflation_snapshots_user_period ON inflation_snapshots(user_id, period_month);

-- 3. day_profiles_cache: Behavioral patterns per day of week
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
CREATE INDEX idx_day_profiles_cache_user ON day_profiles_cache(user_id);
