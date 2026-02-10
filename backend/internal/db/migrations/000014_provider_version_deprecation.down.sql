-- Remove deprecation index
DROP INDEX IF EXISTS idx_provider_versions_deprecated;

-- Remove deprecation fields from provider_versions table
ALTER TABLE provider_versions
DROP COLUMN IF EXISTS deprecated,
DROP COLUMN IF EXISTS deprecated_at,
DROP COLUMN IF EXISTS deprecation_message;
