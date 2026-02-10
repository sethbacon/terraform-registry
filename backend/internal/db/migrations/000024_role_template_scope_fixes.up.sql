-- Migration 024: Fix role template scopes for least privilege
-- - Publisher: Add organizations:read (needed to select org when uploading)
-- - DevOps: Add organizations:read and modules:write (SCM integration works for both modules and providers)
-- - Add new Auditor role for security/compliance review

-- Update Publisher role to include organizations:read
UPDATE role_templates
SET scopes = '["modules:read", "modules:write", "providers:read", "providers:write", "organizations:read", "scm:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'publisher';

-- Update DevOps role to include organizations:read and modules:write
UPDATE role_templates
SET scopes = '["modules:read", "modules:write", "providers:read", "providers:write", "mirrors:read", "mirrors:manage", "organizations:read", "scm:read", "scm:manage"]'::jsonb,
    updated_at = NOW()
WHERE name = 'devops';

-- Add Auditor role for security/compliance personnel
INSERT INTO role_templates (name, display_name, description, scopes, is_system) VALUES
('auditor', 'Auditor', 'Read-only access with audit log visibility for security and compliance review',
 '["modules:read", "providers:read", "mirrors:read", "organizations:read", "scm:read", "audit:read"]'::jsonb, true)
ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    scopes = EXCLUDED.scopes,
    updated_at = NOW();
