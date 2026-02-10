-- Migration 022 Down: Revert organization_members role template changes

-- ============================================================================
-- 1. Re-add role_template_id to users table
-- ============================================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_template_id UUID REFERENCES role_templates(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_role_template ON users(role_template_id);

-- ============================================================================
-- 2. Re-add role column to organization_members
-- ============================================================================
ALTER TABLE organization_members ADD COLUMN IF NOT EXISTS role VARCHAR(50);

-- ============================================================================
-- 3. Migrate role_template_id back to role strings
-- ============================================================================
UPDATE organization_members om
SET role = CASE
    WHEN rt.name = 'admin' THEN 'admin'
    WHEN rt.name = 'publisher' THEN 'member'
    WHEN rt.name = 'devops' THEN 'member'
    WHEN rt.name = 'user_manager' THEN 'admin'
    ELSE 'viewer'
END
FROM role_templates rt
WHERE om.role_template_id = rt.id;

-- Set default role for any without role_template_id
UPDATE organization_members
SET role = 'viewer'
WHERE role IS NULL;

-- Make role NOT NULL after populating
ALTER TABLE organization_members ALTER COLUMN role SET NOT NULL;

-- ============================================================================
-- 4. Drop role_template_id from organization_members
-- ============================================================================
DROP INDEX IF EXISTS idx_org_members_role_template;
ALTER TABLE organization_members DROP COLUMN IF EXISTS role_template_id;

-- ============================================================================
-- 5. Remove User Manager role template
-- ============================================================================
DELETE FROM role_templates WHERE name = 'user_manager';
