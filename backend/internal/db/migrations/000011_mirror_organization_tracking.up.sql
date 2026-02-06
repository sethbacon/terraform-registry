-- Migration 011: Add organization support to mirror configurations and track mirrored providers
-- This migration adds the organization_id to mirror configurations and creates a table
-- to track which providers were mirrored from which mirror configuration.

-- Add organization_id to mirror_configurations
-- In multi-tenant mode, mirrored providers will belong to this organization
-- In single-tenant mode, this can be null (all providers are accessible)
ALTER TABLE mirror_configurations
ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Create index for querying mirrors by organization
CREATE INDEX IF NOT EXISTS idx_mirror_organization ON mirror_configurations(organization_id);

-- Track which providers came from which mirror
-- This allows us to:
-- 1. Know which providers are mirrored vs manually uploaded
-- 2. Re-sync specific providers
-- 3. Remove all providers from a mirror when the mirror is deleted
CREATE TABLE IF NOT EXISTS mirrored_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mirror_config_id UUID NOT NULL REFERENCES mirror_configurations(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    upstream_namespace VARCHAR(255) NOT NULL,
    upstream_type VARCHAR(255) NOT NULL,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_sync_version VARCHAR(50), -- Last version that was synced
    sync_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each provider can only be mirrored from one mirror config
    UNIQUE(provider_id),
    -- Index for finding all providers from a specific mirror
    CONSTRAINT unique_mirror_provider UNIQUE(mirror_config_id, upstream_namespace, upstream_type)
);

-- Index for querying mirrored providers by mirror config
CREATE INDEX IF NOT EXISTS idx_mirrored_providers_mirror ON mirrored_providers(mirror_config_id);

-- Index for looking up if a provider is mirrored
CREATE INDEX IF NOT EXISTS idx_mirrored_providers_provider ON mirrored_providers(provider_id);

-- Track individual version sync status
CREATE TABLE IF NOT EXISTS mirrored_provider_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mirrored_provider_id UUID NOT NULL REFERENCES mirrored_providers(id) ON DELETE CASCADE,
    provider_version_id UUID NOT NULL REFERENCES provider_versions(id) ON DELETE CASCADE,
    upstream_version VARCHAR(50) NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    shasum_verified BOOLEAN DEFAULT false,
    gpg_verified BOOLEAN DEFAULT false,

    UNIQUE(mirrored_provider_id, upstream_version)
);

-- Index for querying versions by mirrored provider
CREATE INDEX IF NOT EXISTS idx_mirrored_versions_provider ON mirrored_provider_versions(mirrored_provider_id);

COMMENT ON TABLE mirrored_providers IS 'Tracks which providers were mirrored from which mirror configuration';
COMMENT ON TABLE mirrored_provider_versions IS 'Tracks individual version sync status for mirrored providers';
COMMENT ON COLUMN mirror_configurations.organization_id IS 'Organization that owns the mirrored providers (null = global/single-tenant)';
