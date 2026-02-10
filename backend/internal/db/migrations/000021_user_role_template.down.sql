-- Migration 021 Down: Remove role_template_id from users table

DROP INDEX IF EXISTS idx_users_role_template;
ALTER TABLE users DROP COLUMN IF EXISTS role_template_id;
