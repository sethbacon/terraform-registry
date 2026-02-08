-- Migration 023: Add organization and API key scopes to User Manager role template
-- User Manager needs:
-- - organizations:read/write to add users to organizations
-- - api_keys:manage to revoke compromised API keys for users they manage

UPDATE role_templates
SET scopes = '["users:read", "users:write", "organizations:read", "organizations:write", "api_keys:manage", "modules:read", "providers:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'user_manager';
