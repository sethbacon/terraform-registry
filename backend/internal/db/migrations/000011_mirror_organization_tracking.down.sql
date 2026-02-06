-- Migration 011: Rollback - Remove organization support and mirrored provider tracking

-- Drop the tracking tables first (they reference mirror_configurations)
DROP TABLE IF EXISTS mirrored_provider_versions;
DROP TABLE IF EXISTS mirrored_providers;

-- Drop the index on organization_id
DROP INDEX IF EXISTS idx_mirror_organization;

-- Remove organization_id column from mirror_configurations
ALTER TABLE mirror_configurations DROP COLUMN IF EXISTS organization_id;
