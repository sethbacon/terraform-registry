-- Remove platform_filter column from mirror_configurations
ALTER TABLE mirror_configurations
DROP COLUMN IF EXISTS platform_filter;
