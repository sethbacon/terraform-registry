-- Migration 021: Add role_template_id to users table
-- This enables assigning role templates directly to users, which defines their permission ceiling
-- Users can only create API keys with scopes that are a subset of their role template's scopes

-- Add role_template_id column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_template_id UUID REFERENCES role_templates(id) ON DELETE SET NULL;

-- Create index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_users_role_template ON users(role_template_id);

-- By default, assign 'viewer' role to existing users without a role template
-- This ensures existing users have a baseline permission level
UPDATE users u
SET role_template_id = rt.id
FROM role_templates rt
WHERE u.role_template_id IS NULL
  AND rt.name = 'viewer';
