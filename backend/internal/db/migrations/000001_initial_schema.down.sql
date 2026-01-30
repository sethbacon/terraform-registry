-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_created;
DROP INDEX IF EXISTS idx_audit_logs_organization;
DROP INDEX IF EXISTS idx_audit_logs_user;
DROP INDEX IF EXISTS idx_download_events_created;
DROP INDEX IF EXISTS idx_download_events_resource;
DROP INDEX IF EXISTS idx_provider_platforms_os_arch;
DROP INDEX IF EXISTS idx_provider_platforms_version;
DROP INDEX IF EXISTS idx_provider_versions_version;
DROP INDEX IF EXISTS idx_provider_versions_provider;
DROP INDEX IF EXISTS idx_providers_namespace;
DROP INDEX IF EXISTS idx_providers_org;
DROP INDEX IF EXISTS idx_module_versions_version;
DROP INDEX IF EXISTS idx_module_versions_module;
DROP INDEX IF EXISTS idx_modules_namespace;
DROP INDEX IF EXISTS idx_modules_org;
DROP INDEX IF EXISTS idx_organization_members_user_id;
DROP INDEX IF EXISTS idx_api_keys_organization_id;
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_users_oidc_sub;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS download_events;
DROP TABLE IF EXISTS provider_platforms;
DROP TABLE IF EXISTS provider_versions;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS module_versions;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";
