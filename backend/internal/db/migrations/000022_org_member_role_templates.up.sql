-- Migration 022: Move role templates to organization_members
-- This changes from global user roles to per-organization role templates
-- and adds the User Manager role template

-- ============================================================================
-- 1. Add User Manager role template
-- ============================================================================
INSERT INTO role_templates (name, display_name, description, scopes, is_system) VALUES
('user_manager', 'User Manager', 'Can manage user accounts and memberships',
 '["users:read", "users:write", "modules:read", "providers:read"]'::jsonb, true)
ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    scopes = EXCLUDED.scopes,
    updated_at = NOW();

-- ============================================================================
-- 2. Add role_template_id to organization_members
-- ============================================================================
ALTER TABLE organization_members ADD COLUMN IF NOT EXISTS role_template_id UUID REFERENCES role_templates(id) ON DELETE SET NULL;

-- Create index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_org_members_role_template ON organization_members(role_template_id);

-- ============================================================================
-- 3. Migrate existing role values to role template IDs
-- Map: owner/admin -> admin, member -> publisher, viewer -> viewer
-- ============================================================================
UPDATE organization_members om
SET role_template_id = rt.id
FROM role_templates rt
WHERE om.role_template_id IS NULL
  AND ((om.role IN ('owner', 'admin') AND rt.name = 'admin')
    OR (om.role = 'member' AND rt.name = 'publisher')
    OR (om.role = 'viewer' AND rt.name = 'viewer'));

-- Assign viewer as default for any remaining without a role template
UPDATE organization_members om
SET role_template_id = rt.id
FROM role_templates rt
WHERE om.role_template_id IS NULL
  AND rt.name = 'viewer';

-- ============================================================================
-- 4. Drop the old role column from organization_members
-- ============================================================================
ALTER TABLE organization_members DROP COLUMN IF EXISTS role;

-- ============================================================================
-- 5. Remove role_template_id from users table (was added in migration 021)
-- ============================================================================
DROP INDEX IF EXISTS idx_users_role_template;
ALTER TABLE users DROP COLUMN IF EXISTS role_template_id;
