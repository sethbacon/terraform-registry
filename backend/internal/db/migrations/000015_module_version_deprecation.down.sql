-- Remove deprecation index
DROP INDEX IF EXISTS idx_module_versions_deprecated;

-- Remove deprecation fields from module_versions table
ALTER TABLE module_versions
DROP COLUMN IF EXISTS deprecated,
DROP COLUMN IF EXISTS deprecated_at,
DROP COLUMN IF EXISTS deprecation_message;
