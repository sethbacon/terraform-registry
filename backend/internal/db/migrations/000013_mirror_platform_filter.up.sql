-- Add platform_filter column to mirror_configurations
-- Format: JSON array of "os/arch" strings, e.g. ["linux/amd64", "windows/amd64"]
ALTER TABLE mirror_configurations
ADD COLUMN IF NOT EXISTS platform_filter TEXT;

COMMENT ON COLUMN mirror_configurations.platform_filter IS 'JSON array of platform strings in "os/arch" format (e.g. ["linux/amd64", "windows/amd64"])';
