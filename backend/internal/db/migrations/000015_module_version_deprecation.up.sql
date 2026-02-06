-- Add deprecation fields to module_versions table
ALTER TABLE module_versions
ADD COLUMN IF NOT EXISTS deprecated BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS deprecated_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS deprecation_message TEXT;

-- Add index for deprecated versions lookup
CREATE INDEX IF NOT EXISTS idx_module_versions_deprecated ON module_versions(deprecated);
