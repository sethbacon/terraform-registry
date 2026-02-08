-- Rollback: Remove organization scopes from User Manager role template

UPDATE role_templates
SET scopes = '["users:read", "users:write", "modules:read", "providers:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'user_manager';
