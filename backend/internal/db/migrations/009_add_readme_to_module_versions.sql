-- Add readme column to module_versions table
ALTER TABLE module_versions ADD COLUMN readme TEXT;

-- Add comment
COMMENT ON COLUMN module_versions.readme IS 'README content extracted from module tarball';
