-- +migrate Up

-- SCM provider configurations per organization
CREATE TABLE scm_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    provider_type VARCHAR(50) NOT NULL CHECK (provider_type IN ('github', 'azuredevops', 'gitlab')),
    name VARCHAR(255) NOT NULL,
    base_url VARCHAR(512),
    client_id VARCHAR(255) NOT NULL,
    client_secret_encrypted TEXT NOT NULL,
    webhook_secret VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, provider_type, name)
);

CREATE INDEX idx_scm_providers_org ON scm_providers(organization_id);
CREATE INDEX idx_scm_providers_type ON scm_providers(provider_type);

-- User OAuth tokens (one per user per provider)
CREATE TABLE scm_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    scm_provider_id UUID REFERENCES scm_providers(id) ON DELETE CASCADE,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT,
    token_type VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at TIMESTAMP,
    scopes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, scm_provider_id)
);

CREATE INDEX idx_scm_oauth_tokens_user ON scm_oauth_tokens(user_id);
CREATE INDEX idx_scm_oauth_tokens_provider ON scm_oauth_tokens(scm_provider_id);

-- Module <-> SCM repository links
CREATE TABLE module_scm_repos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id UUID REFERENCES modules(id) ON DELETE CASCADE,
    scm_provider_id UUID REFERENCES scm_providers(id) ON DELETE CASCADE,
    repository_owner VARCHAR(255) NOT NULL,
    repository_name VARCHAR(255) NOT NULL,
    repository_url VARCHAR(512),
    default_branch VARCHAR(255) DEFAULT 'main',
    module_path VARCHAR(512) DEFAULT '/',
    tag_pattern VARCHAR(255) DEFAULT 'v*',
    auto_publish BOOLEAN DEFAULT true,
    webhook_id VARCHAR(255),
    webhook_url VARCHAR(512),
    webhook_enabled BOOLEAN DEFAULT false,
    last_sync_at TIMESTAMP,
    last_sync_commit VARCHAR(40),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (module_id)
);

CREATE INDEX idx_module_scm_repos_module ON module_scm_repos(module_id);
CREATE INDEX idx_module_scm_repos_provider ON module_scm_repos(scm_provider_id);

-- Webhook event log for debugging and audit
CREATE TABLE scm_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_scm_repo_id UUID REFERENCES module_scm_repos(id) ON DELETE CASCADE,
    event_id VARCHAR(255),
    event_type VARCHAR(50) NOT NULL,
    ref VARCHAR(255),
    commit_sha VARCHAR(40),
    tag_name VARCHAR(255),
    payload JSONB NOT NULL,
    headers JSONB,
    signature VARCHAR(255),
    signature_valid BOOLEAN,
    processed BOOLEAN DEFAULT false,
    processing_started_at TIMESTAMP,
    processed_at TIMESTAMP,
    result_version_id UUID REFERENCES module_versions(id),
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scm_webhook_events_repo ON scm_webhook_events(module_scm_repo_id);
CREATE INDEX idx_scm_webhook_events_processed ON scm_webhook_events(processed);
CREATE INDEX idx_scm_webhook_events_created ON scm_webhook_events(created_at DESC);
CREATE INDEX idx_scm_webhook_events_event_id ON scm_webhook_events(event_id);

-- Track version immutability violations (tag movements)
CREATE TABLE version_immutability_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_version_id UUID REFERENCES module_versions(id) ON DELETE CASCADE,
    tag_name VARCHAR(255) NOT NULL,
    original_commit_sha VARCHAR(40) NOT NULL,
    detected_commit_sha VARCHAR(40) NOT NULL,
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    alert_sent BOOLEAN DEFAULT false,
    alert_sent_at TIMESTAMP,
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP,
    resolved_by UUID REFERENCES users(id),
    notes TEXT
);

CREATE INDEX idx_immutability_violations_version ON version_immutability_violations(module_version_id);
CREATE INDEX idx_immutability_violations_unresolved ON version_immutability_violations(resolved) WHERE resolved = false;

-- Extend module_versions with SCM metadata
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS commit_sha VARCHAR(40);
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS scm_source VARCHAR(512);
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS tag_name VARCHAR(255);
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS scm_repo_id UUID REFERENCES module_scm_repos(id);

CREATE INDEX idx_module_versions_commit_sha ON module_versions(commit_sha);
CREATE INDEX idx_module_versions_tag_name ON module_versions(tag_name);
CREATE INDEX idx_module_versions_scm_repo ON module_versions(scm_repo_id);

-- +migrate Down

DROP INDEX IF EXISTS idx_module_versions_scm_repo;
DROP INDEX IF EXISTS idx_module_versions_tag_name;
DROP INDEX IF EXISTS idx_module_versions_commit_sha;

ALTER TABLE module_versions DROP COLUMN IF EXISTS scm_repo_id;
ALTER TABLE module_versions DROP COLUMN IF EXISTS tag_name;
ALTER TABLE module_versions DROP COLUMN IF EXISTS scm_source;
ALTER TABLE module_versions DROP COLUMN IF EXISTS commit_sha;

DROP TABLE IF EXISTS version_immutability_violations;
DROP TABLE IF EXISTS scm_webhook_events;
DROP TABLE IF EXISTS module_scm_repos;
DROP TABLE IF EXISTS scm_oauth_tokens;
DROP TABLE IF EXISTS scm_providers;
