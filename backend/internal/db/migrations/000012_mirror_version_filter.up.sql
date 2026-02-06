-- Migration 012: Add version_filter to mirror_configurations
-- This allows filtering which versions to sync (e.g., "3." for all 3.x, "latest:5" for latest 5)

ALTER TABLE mirror_configurations
ADD COLUMN IF NOT EXISTS version_filter TEXT;

-- Examples of version_filter values:
-- "3." or "3.x" - all versions starting with 3.
-- "latest:5" - only the latest 5 versions
-- "3.74.0,3.73.0,3.72.0" - specific versions only
-- ">=3.0.0" - versions 3.0.0 and higher
-- null/empty - all versions

COMMENT ON COLUMN mirror_configurations.version_filter IS 'Filter for which versions to sync. Supports: prefix (3.), latest:N, comma-separated list, or semver constraint (>=3.0.0)';
