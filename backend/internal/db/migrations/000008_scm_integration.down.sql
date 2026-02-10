-- Reverse migration for SCM integration
DROP TABLE IF EXISTS scm_webhook_events CASCADE;
DROP TABLE IF EXISTS module_source_repos CASCADE;
DROP TABLE IF EXISTS scm_oauth_tokens CASCADE;
DROP TABLE IF EXISTS scm_providers CASCADE;
