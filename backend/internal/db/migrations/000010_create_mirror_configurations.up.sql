-- Migration 010: Create mirror_configurations table
-- This table stores configuration for provider mirroring from upstream registries

CREATE TABLE IF NOT EXISTS mirror_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    upstream_registry_url VARCHAR(512) NOT NULL,
    namespace_filter TEXT, -- JSON array of namespaces to mirror (null = all)
    provider_filter TEXT, -- JSON array of provider names to mirror (null = all)
    enabled BOOLEAN NOT NULL DEFAULT true,
    sync_interval_hours INTEGER NOT NULL DEFAULT 24, -- How often to sync (hours)
    last_sync_at TIMESTAMPTZ,
    last_sync_status VARCHAR(50), -- 'success', 'failed', 'in_progress'
    last_sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    CONSTRAINT valid_sync_interval CHECK (sync_interval_hours > 0),
    CONSTRAINT valid_registry_url CHECK (upstream_registry_url LIKE 'http%')
);

-- Index for querying enabled mirrors that need syncing
CREATE INDEX idx_mirror_enabled_last_sync ON mirror_configurations(enabled, last_sync_at) 
    WHERE enabled = true;

-- Index for filtering by upstream registry
CREATE INDEX idx_mirror_upstream_registry ON mirror_configurations(upstream_registry_url);

-- Track sync history
CREATE TABLE IF NOT EXISTS mirror_sync_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mirror_config_id UUID NOT NULL REFERENCES mirror_configurations(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL, -- 'running', 'success', 'failed', 'cancelled'
    providers_synced INTEGER DEFAULT 0,
    providers_failed INTEGER DEFAULT 0,
    error_message TEXT,
    sync_details JSONB, -- Detailed sync information (versions added, errors, etc.)
    
    CONSTRAINT valid_status CHECK (status IN ('running', 'success', 'failed', 'cancelled'))
);

-- Index for querying sync history by mirror config
CREATE INDEX idx_sync_history_mirror_config ON mirror_sync_history(mirror_config_id, started_at DESC);

-- Index for querying active syncs
CREATE INDEX idx_sync_history_status ON mirror_sync_history(status, started_at) 
    WHERE status = 'running';

COMMENT ON TABLE mirror_configurations IS 'Configuration for provider mirroring from upstream registries';
COMMENT ON TABLE mirror_sync_history IS 'Historical record of mirror synchronization operations';
