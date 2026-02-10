-- Add readme column to module_versions table (if not exists)
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS readme TEXT;

-- Add comment
COMMENT ON COLUMN module_versions.readme IS 'README content extracted from module tarball';
