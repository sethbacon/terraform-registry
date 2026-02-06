-- Migration 012 down: Remove version_filter from mirror_configurations

ALTER TABLE mirror_configurations
DROP COLUMN IF EXISTS version_filter;
