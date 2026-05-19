-- backend/internal/db/migrations/20260519_add_llm_keys_to_users.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS groq_api_key_encrypted TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS gemini_api_key_encrypted TEXT;
