-- Migration 025: Ensure admin users have proper role template assignments
-- This fixes users who were created before the RBAC migration or whose
-- organization memberships weren't properly assigned admin role templates.

-- ============================================================================
-- 1. Ensure the default organization exists
-- ============================================================================
INSERT INTO organizations (name, display_name, created_at, updated_at)
VALUES ('default', 'Default organization for system administration', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- 2. Ensure admin@dev.local user has an organization membership
-- ============================================================================
DO $$
DECLARE
    v_user_id UUID;
    v_org_id UUID;
    v_admin_role_id UUID;
BEGIN
    -- Get the dev admin user ID
    SELECT id INTO v_user_id FROM users WHERE email = 'admin@dev.local';

    -- Get the default organization ID
    SELECT id INTO v_org_id FROM organizations WHERE name = 'default';

    -- Get the admin role template ID
    SELECT id INTO v_admin_role_id FROM role_templates WHERE name = 'admin';

    -- Skip if any of these don't exist
    IF v_user_id IS NULL OR v_org_id IS NULL OR v_admin_role_id IS NULL THEN
        RAISE NOTICE 'Skipping: user=%, org=%, role=%', v_user_id, v_org_id, v_admin_role_id;
        RETURN;
    END IF;

    -- Ensure the user is a member of the default organization with admin role
    INSERT INTO organization_members (organization_id, user_id, role_template_id, created_at)
    VALUES (v_org_id, v_user_id, v_admin_role_id, NOW())
    ON CONFLICT (organization_id, user_id)
    DO UPDATE SET role_template_id = EXCLUDED.role_template_id;

    RAISE NOTICE 'Admin user % assigned to organization % with admin role %', v_user_id, v_org_id, v_admin_role_id;
END $$;

-- ============================================================================
-- 3. Ensure any organization_members without a role_template_id get the viewer role
-- This is a safety net for any members that slipped through previous migrations
-- ============================================================================
UPDATE organization_members om
SET role_template_id = rt.id
FROM role_templates rt
WHERE om.role_template_id IS NULL
  AND rt.name = 'viewer';
